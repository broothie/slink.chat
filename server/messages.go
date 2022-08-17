package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/broothie/slink.chat/model"
	"github.com/broothie/slink.chat/util"
	"github.com/go-chi/chi/v5"
	"github.com/rs/xid"
)

func (s *Server) createMessage(w http.ResponseWriter, r *http.Request) {
	var params model.Message
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
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

	_, err := s.Collection("channels").
		Doc(chi.URLParam(r, "channel_id")).
		Collection("messages").
		Doc(message.ID).
		Create(r.Context(), message)
	if err != nil {
		s.render.JSON(w, http.StatusBadRequest, errorMap(err))
		return
	}

	s.render.JSON(w, http.StatusCreated, util.Map{"message": message})
}
