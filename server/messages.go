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
	logger := ctxzap.Extract(r.Context())

	snapshots, err := s.db.
		Collection("channels").
		Doc(chi.URLParam(r, "channel_id")).
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

	messages := make([]model.Message, 0, len(snapshots))
	for _, snapshot := range snapshots {
		var message model.Message
		if err := snapshot.DataTo(&message); err != nil {
			logger.Error("failed to read message", zap.Error(err))
			continue
		}

		messages = append(messages, message)
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

	now := time.Now()
	message := model.Message{
		ID:        xid.New().String(),
		CreatedAt: now,
		UpdatedAt: now,
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
