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
	"github.com/rs/xid"
	"github.com/samber/lo"
	"go.uber.org/zap"
)

func (s *Server) indexMessages(w http.ResponseWriter, r *http.Request) {
	logger := ctxzap.Extract(r.Context())

	messageSlice, err := db.NewFetcher[model.Message](s.DB).Query(r.Context(), func(query *firestore.CollectionRef) firestore.Query {
		return query.
			Where("channel_id", "==", chi.URLParam(r, "channel_id")).
			OrderBy("created_at", firestore.Asc).
			LimitToLast(100)
	})
	if err != nil {
		logger.Error("failed to get messages", zap.Error(err))
		s.render.JSON(w, http.StatusInternalServerError, errorMap(err))
		return
	}

	messages := lo.Associate(messageSlice, func(message model.Message) (string, model.Message) {
		return message.ID, message
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
	if _, err := db.NewFetcher[model.Channel](s.DB).FetchFirst(r.Context(), func(query *firestore.CollectionRef) firestore.Query {
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
	batch.Update(s.DB.CollectionFor(model.TypeChannel).Doc(channelID), []firestore.Update{{
		Path:  "last_message_sent_at",
		Value: now,
	}})

	if _, err := batch.Commit(r.Context()); err != nil {
		logger.Error("failed to create message", zap.Error(err))
		s.render.JSON(w, http.StatusBadRequest, errorMap(err))
		return
	}

	s.render.JSON(w, http.StatusCreated, util.Map{"message": message})
}
