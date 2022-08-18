package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/broothie/slink.chat/model"
	"github.com/broothie/slink.chat/util"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/rs/xid"
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
		ID:        xid.New().String(),
		CreatedAt: now,
		UpdatedAt: now,
		UserID:    user.ID,
		ChannelID: params.ChannelID,
	}

	if _, err := s.db.Collection("subscriptions").Doc(subscription.ID).Create(r.Context(), subscription); err != nil {
		logger.Error("failed to create subscription", zap.Error(err))
		s.render.JSON(w, http.StatusInternalServerError, errorMap(err))
		return
	}

	s.render.JSON(w, http.StatusCreated, util.Map{"subscription": subscription})
}
