package server

import (
	"net/http"

	"cloud.google.com/go/firestore"
	"github.com/broothie/slink.chat/db"
	"github.com/broothie/slink.chat/model"
	"github.com/broothie/slink.chat/util"
	"github.com/go-chi/chi/v5"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/samber/lo"
	"go.uber.org/zap"
)

func (s *Server) indexMessages(w http.ResponseWriter, r *http.Request) {
	logger := ctxzap.Extract(r.Context())

	messageSlice, err := db.NewFetcher[model.Message](s.DB).Query(r.Context(), func(query firestore.Query) firestore.Query {
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
