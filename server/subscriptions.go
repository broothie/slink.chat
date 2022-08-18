package server

import (
	"encoding/json"
	"net/http"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/broothie/slink.chat/model"
	"github.com/broothie/slink.chat/util"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/rs/xid"
	"go.uber.org/zap"
)

func (s *Server) indexSubscriptions(w http.ResponseWriter, r *http.Request) {
	type Subscription struct {
		model.Subscription
		Name string `json:"name"`
	}

	logger := ctxzap.Extract(r.Context())

	user, _ := model.UserFromContext(r.Context())
	subscriptionSnapshots, err := s.db.
		Collection("users").
		Doc(user.ID).
		Collection("subscriptions").
		Documents(r.Context()).
		GetAll()
	if err != nil {
		logger.Error("failed to get subscriptions", zap.Error(err))
		s.render.JSON(w, http.StatusInternalServerError, errorMap(err))
		return
	}

	subscriptions := make(map[string]Subscription, len(subscriptionSnapshots))
	channelRefs := make([]*firestore.DocumentRef, 0, len(subscriptionSnapshots))
	for _, snapshot := range subscriptionSnapshots {
		var subscription model.Subscription
		if err := snapshot.DataTo(&subscription); err != nil {
			logger.Error("failed to read subscription data", zap.Error(err))
			continue
		}

		subscriptions[subscription.ChannelID] = Subscription{Subscription: subscription}
		channelRefs = append(channelRefs, s.db.Collection("channels").Doc(subscription.ChannelID))
	}

	channelSnapshots, err := s.db.GetAll(r.Context(), channelRefs)
	if err != nil {
		logger.Error("failed to get channels", zap.Error(err))
		s.render.JSON(w, http.StatusInternalServerError, errorMap(err))
		return
	}

	for _, snapshot := range channelSnapshots {
		var channel model.Channel
		if err := snapshot.DataTo(&channel); err != nil {
			logger.Error("failed to read channel data", zap.Error(err))
			continue
		}

		subscription, found := subscriptions[channel.ID]
		if found {
			subscription.Name = channel.Name
			subscriptions[channel.ID] = subscription
		}
	}

	s.render.JSON(w, http.StatusOK, util.Map{"subscriptions": subscriptions})
}

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
		ChannelID: params.ChannelID,
	}

	_, err := s.db.
		Collection("users").
		Doc(user.ID).
		Collection("subscriptions").
		Doc(subscription.ID).
		Create(r.Context(), subscription)
	if err != nil {
		logger.Error("failed to create subscription", zap.Error(err))
		s.render.JSON(w, http.StatusInternalServerError, errorMap(err))
		return
	}

	s.render.JSON(w, http.StatusCreated, util.Map{"subscription": subscription})
}
