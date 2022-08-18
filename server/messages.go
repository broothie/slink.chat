package server

import (
	"encoding/json"
	"net/http"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/broothie/slink.chat/model"
	"github.com/broothie/slink.chat/util"
	"github.com/go-chi/chi/v5"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/rs/xid"
	"go.uber.org/zap"
)

func (s *Server) indexMessages(w http.ResponseWriter, r *http.Request) {
	type Message struct {
		model.Message
		ChannelID string `json:"channelID"`
	}

	logger := ctxzap.Extract(r.Context())

	channelID := chi.URLParam(r, "channel_id")
	snapshots, err := s.db.
		Collection("channels").
		Doc(channelID).
		Collection("messages").
		OrderBy("created_at", firestore.Desc).
		Limit(100).
		Documents(r.Context()).
		GetAll()
	if err != nil {
		logger.Error("failed to read json body", zap.Error(err))
		s.render.JSON(w, http.StatusBadRequest, errorMap(err))
		return
	}

	messages := make(map[string]Message, len(snapshots))
	for _, snapshot := range snapshots {
		var message Message
		if err := snapshot.DataTo(&message); err != nil {
			logger.Error("failed to read message", zap.Error(err))
			continue
		}

		message.ChannelID = channelID
		messages[message.ID] = message
	}

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

	user, _ := model.UserFromContext(r.Context())
	now := time.Now()
	message := model.Message{
		ID:        xid.New().String(),
		CreatedAt: now,
		UpdatedAt: now,
		UserID:    user.ID,
		Body:      params.Body,
	}

	_, err := s.db.Collection("channels").
		Doc(chi.URLParam(r, "channel_id")).
		Collection("messages").
		Doc(message.ID).
		Create(r.Context(), message)
	if err != nil {
		logger.Error("failed to create message", zap.Error(err))
		s.render.JSON(w, http.StatusBadRequest, errorMap(err))
		return
	}

	s.render.JSON(w, http.StatusCreated, util.Map{"message": message})
}
