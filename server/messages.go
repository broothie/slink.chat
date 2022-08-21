package server

import (
	"encoding/json"
	"net/http"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/broothie/slink.chat/db"
	"github.com/broothie/slink.chat/model"
	"github.com/broothie/slink.chat/util"
	"github.com/go-chi/chi/v5"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"go.uber.org/zap"
)

func (s *Server) indexMessages(w http.ResponseWriter, r *http.Request) {
	logger := ctxzap.Extract(r.Context())

	messageSlice, err := db.NewFetcher[model.Message](s.db).Query(r.Context(), func(query firestore.Query) firestore.Query {
		return query.Where("channel_id", "==", chi.URLParam(r, "channel_id"))
	})
	if err != nil {
		logger.Error("failed to get messages", zap.Error(err))
		s.render.JSON(w, http.StatusInternalServerError, errorMap(err))
		return
	}

	messages := lo.Associate(messageSlice, func(message model.Message) (string, model.Message) {
		return message.MessageID, message
	})

	s.render.JSON(w, http.StatusOK, util.Map{"messages": messages})
}

func (s *Server) createMessage(w http.ResponseWriter, r *http.Request) {
	logger := ctxzap.Extract(r.Context())

	var params model.Message
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		logger.Error("failed to read json body", zap.Error(err))
		s.render.JSON(w, http.StatusBadRequest, errorMap(err))
		return
	}

	channelID := chi.URLParam(r, "channel_id")
	user, _ := model.UserFromContext(r.Context())
	if _, err := db.NewFetcher[model.Channel](s.db).FetchFirst(r.Context(), func(query firestore.Query) firestore.Query {
		return query.Where("user_ids", "array-contains", user.UserID)
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

	now := time.Now()
	message := model.Message{
		MessageID: model.TypeMessage.NewID(),
		Type:      model.TypeMessage,
		CreatedAt: now,
		UpdatedAt: now,
		UserID:    user.UserID,
		ChannelID: channelID,
		Body:      params.Body,
	}

	if _, err := s.db.Collection().Doc(message.MessageID).Create(r.Context(), message); err != nil {
		logger.Error("failed to create message", zap.Error(err))
		s.render.JSON(w, http.StatusBadRequest, errorMap(err))
		return
	}

	s.render.JSON(w, http.StatusCreated, util.Map{"message": message})
}
