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
	"go.uber.org/zap"
)

func (s *Server) createSubscription(w http.ResponseWriter, r *http.Request) {
	logger := ctxzap.Extract(r.Context())

	var params model.Subscription
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		logger.Error("failed to decode subscription json", zap.Error(err))
		s.render.JSON(w, http.StatusBadRequest, errorMap(err))
		return
	}

	user, _ := model.UserFromContext(r.Context())
	now := time.Now()
	subscription := model.Subscription{
		SubscriptionID: model.TypeSubscription.NewID(),
		Type:           model.TypeSubscription,
		CreatedAt:      now,
		UpdatedAt:      now,
		UserID:         user.UserID,
		ChannelID:      params.ChannelID,
	}

	if _, err := s.db.Collection().Doc(subscription.SubscriptionID).Create(r.Context(), subscription); err != nil {
		logger.Error("failed to create subscription", zap.Error(err))
		s.render.JSON(w, http.StatusInternalServerError, errorMap(err))
		return
	}

	s.render.JSON(w, http.StatusCreated, util.Map{"subscription": subscription})
}

func (s *Server) destroySubscription(w http.ResponseWriter, r *http.Request) {
	logger := ctxzap.Extract(r.Context())

	channelID := chi.URLParam(r, "channel_id")
	user, _ := model.UserFromContext(r.Context())
	subscription, err := db.NewFetcher[model.Subscription](s.db).FetchFirst(r.Context(), func(query firestore.Query) firestore.Query {
		return query.Where("user_id", "==", user.UserID).Where("channel_id", "==", channelID)
	})
	if err != nil {
		logger.Error("failed to get subscription", zap.Error(err))
		s.render.JSON(w, http.StatusInternalServerError, errorMap(err))
		return
	}

	if _, err := s.db.Collection().Doc(subscription.SubscriptionID).Delete(r.Context()); err != nil {
		logger.Error("failed to delete subscription", zap.Error(err))
		s.render.JSON(w, http.StatusInternalServerError, errorMap(err))
		return
	}

	s.render.JSON(w, http.StatusOK, util.Map{"channelID": channelID})
}
