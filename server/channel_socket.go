package server

import (
	"encoding/json"
	"net/http"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/broothie/slink.chat/db"
	"github.com/broothie/slink.chat/model"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/pkg/errors"
	"github.com/rs/xid"
	"github.com/samber/lo"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) channelSocket(w http.ResponseWriter, r *http.Request) {
	logger := ctxzap.Extract(r.Context()).With(zap.String("at", "channelSocket"))

	user, _ := model.UserFromContext(r.Context())
	channelID := chi.URLParam(r, "channel_id")
	if _, err := db.NewFetcher[model.Channel](s.DB).FetchFirst(r.Context(), func(query firestore.Query) firestore.Query {
		return query.Where("user_ids", "array-contains", user.ID)
	}); err != nil {
		if err == db.NotFound {
			logger.Info("user not in channel")
			s.render.JSON(w, http.StatusUnauthorized, errorMap(errors.New("user not in channel")))
			return
		}

		logger.Error("error finding subscription")
		s.render.JSON(w, http.StatusInternalServerError, errorMap(err))
		return
	}

	upgrader := &websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error("failed to upgrade request", zap.Error(err))
		s.render.JSON(w, http.StatusInternalServerError, errorMap(err))
		return
	}

	defer func() {
		if err := conn.Close(); err != nil {
			logger.Error("failed to close connection", zap.Error(err))
		}
	}()

	socketCloseChan := make(chan struct{})
	go func() {
		defer close(socketCloseChan)

		for {
			messageType, socketReader, err := conn.NextReader()
			if err != nil {
				if _, isCloseErr := err.(*websocket.CloseError); !isCloseErr {
					logger.Error("next reader error", zap.Error(err))
				}

				return
			}

			if messageType != websocket.TextMessage {
				logger.Info("reader received non-text message type", zap.Int("message_type", messageType))
				continue
			}

			var params model.Message
			if err := json.NewDecoder(socketReader).Decode(&params); err != nil {
				logger.Error("failed to decode message", zap.Error(err))
				continue
			}

			now := time.Now()
			message := model.Message{
				ID:        xid.New().String(),
				CreatedAt: now,
				UpdatedAt: now,
				UserID:    user.ID,
				ChannelID: channelID,
				Body:      params.Body,
			}

			batch := s.DB.Batch()
			batch.Create(s.DB.CollectionFor(message.Type()).Doc(message.ID), message)
			batch.Update(s.DB.CollectionFor(model.TypeChannel).Doc(channelID), []firestore.Update{
				{Path: "updated_at", Value: now},
				{Path: "last_message_sent_at", Value: now},
			})

			if _, err := batch.Commit(r.Context()); err != nil {
				logger.Error("failed to create message", zap.Error(err))
				return
			}
		}
	}()

	dbCloseChan := make(chan struct{})
	messages := make(chan model.Message)
	go func() {
		defer close(dbCloseChan)

		logger.Debug("listening for messages")
		snapshots := s.DB.
			CollectionFor(model.TypeMessage).
			Where("channel_id", "==", channelID).
			Where("created_at", ">", time.Now()).
			Snapshots(r.Context())
		defer snapshots.Stop()

		for {
			snapshot, err := snapshots.Next()
			if err != nil {
				if status.Code(err) == codes.DeadlineExceeded {
					logger.Debug("db listen timeout", zap.Error(err))
					return
				} else if status.Code(err) == codes.Canceled {
					return
				}

				logger.Error("next snapshot error", zap.Error(err))
				return
			}

			if snapshot == nil {
				continue
			}

			change, found := lo.Find(snapshot.Changes, func(change firestore.DocumentChange) bool {
				return change.Kind == firestore.DocumentAdded
			})

			if !found {
				continue
			}

			var message model.Message
			if err := change.Doc.DataTo(&message); err != nil {
				logger.Error("failed to read message", zap.Error(err))
				continue
			}

			messages <- message
		}
	}()

	logger.Debug("socket opened")
	for {
		select {
		case <-socketCloseChan:
			logger.Info("client closed socket")
			return

		case <-dbCloseChan:
			logger.Info("db closed stream")
			return

		case message := <-messages:
			socketWriter, err := conn.NextWriter(websocket.TextMessage)
			if err != nil {
				logger.Error("failed to get writer", zap.Error(err))
				continue
			}

			if err := json.NewEncoder(socketWriter).Encode(message); err != nil {
				logger.Error("failed to write json to socket", zap.Error(err))
				continue
			}

			if err := socketWriter.Close(); err != nil {
				logger.Error("failed to close socket writer", zap.Error(err))
			}
		}
	}
}
