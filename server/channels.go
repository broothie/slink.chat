package server

import (
	"encoding/json"
	"net/http"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/broothie/slink.chat/model"
	"github.com/broothie/slink.chat/util"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/rs/xid"
	"github.com/samber/lo"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) createChannel(w http.ResponseWriter, r *http.Request) {
	var params model.Channel
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		s.render.JSON(w, http.StatusBadRequest, errorMap(err))
		return
	}

	now := time.Now()
	channel := model.Channel{
		ID:        xid.New().String(),
		CreatedAt: now,
		UpdatedAt: now,
		Name:      params.Name,
	}

	if _, err := s.db.Collection("channels").Doc(channel.ID).Create(r.Context(), channel); err != nil {
		s.render.JSON(w, http.StatusBadRequest, errorMap(err))
		return
	}

	s.render.JSON(w, http.StatusCreated, util.Map{"channel": channel})
}

func (s *Server) channelSubscribe(w http.ResponseWriter, r *http.Request) {
	logger := ctxzap.Extract(r.Context())

	upgrader := &websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
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
			if _, _, err := conn.NextReader(); err != nil {
				if _, isCloseErr := err.(*websocket.CloseError); !isCloseErr {
					logger.Error("next reader error", zap.Error(err))
				}

				return
			}
		}
	}()

	dbCloseChan := make(chan struct{})
	messages := make(chan model.Message)
	go func() {
		defer close(dbCloseChan)

		snapshots := s.db.Collection("channels").
			Doc(chi.URLParam(r, "channel_id")).
			Collection("messages").
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

	for {
		select {
		case <-socketCloseChan:
			return

		case <-dbCloseChan:
			return

		case message := <-messages:
			socket, err := conn.NextWriter(websocket.TextMessage)
			if err != nil {
				logger.Error("failed to get writer", zap.Error(err))
				continue
			}

			if err := json.NewEncoder(socket).Encode(message); err != nil {
				logger.Error("failed to write json to socket", zap.Error(err))
				continue
			}

			if err := socket.Close(); err != nil {
				logger.Error("failed to close socket", zap.Error(err))
			}
		}
	}
}
