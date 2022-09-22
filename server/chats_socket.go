package server

import (
	"encoding/json"
	"net/http"

	"cloud.google.com/go/firestore"
	"github.com/broothie/slink.chat/model"
	"github.com/gorilla/websocket"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/samber/lo"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) channelsSocket(w http.ResponseWriter, r *http.Request) {
	logger := ctxzap.Extract(r.Context()).With(zap.String("at", "channelsSocket"))

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
			if _, _, err := conn.NextReader(); err != nil {
				if _, isCloseErr := err.(*websocket.CloseError); !isCloseErr {
					logger.Error("next reader error", zap.Error(err))
				}

				return
			}
		}
	}()

	user, _ := model.UserFromContext(r.Context())
	dbCloseChan := make(chan struct{})
	channels := make(chan model.Channel)
	go func() {
		defer close(dbCloseChan)

		logger.Debug("listening for channel updates")
		snapshots := s.DB.CollectionFor(model.TypeChannel).
			Where("private", "==", true).
			Where("user_ids", "array-contains", user.ID).
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
				return change.Kind == firestore.DocumentModified
			})

			if !found {
				continue
			}

			var channel model.Channel
			if err := change.Doc.DataTo(&channel); err != nil {
				logger.Error("failed to read channel change", zap.Error(err))
				continue
			}

			channels <- channel
		}
	}()

	logger.Debug("channels socket opened")
	for {
		select {
		case <-socketCloseChan:
			logger.Info("client closed socket")
			return

		case <-dbCloseChan:
			logger.Info("db closed stream")
			return

		case channel := <-channels:
			socketWriter, err := conn.NextWriter(websocket.TextMessage)
			if err != nil {
				logger.Error("failed to get writer", zap.Error(err))
				continue
			}

			if err := json.NewEncoder(socketWriter).Encode(channel); err != nil {
				logger.Error("failed to write json to socket", zap.Error(err))
				continue
			}

			if err := socketWriter.Close(); err != nil {
				logger.Error("failed to close socket writer", zap.Error(err))
			}
		}
	}
}
