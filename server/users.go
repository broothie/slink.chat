package server

import (
	"context"
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
	"go.uber.org/zap"
)

type userParams struct {
	model.User
	Password string `json:"password"`
}

func (s *Server) createUser(w http.ResponseWriter, r *http.Request) {
	logger := ctxzap.Extract(r.Context())

	var params userParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		logger.Error("failed to decode body", zap.Error(err))
		s.render.JSON(w, http.StatusBadRequest, errorMap(err))
		return
	}

	logger = logger.With(zap.String("screenname", params.Screenname))
	if _, err := db.NewFetcher[model.User](s.db).FetchFirst(r.Context(), func(query firestore.Query) firestore.Query {
		return query.Where("screenname", "==", params.Screenname)
	}); err == nil {
		logger.Info("screenname is taken")
		s.render.JSON(w, http.StatusBadRequest, errorMap(errors.New("screenname is taken")))
		return
	} else if err != db.NotFound {
		logger.Error("failed to look for users", zap.Error(err))
		s.render.JSON(w, http.StatusInternalServerError, errorMap(err))
		return
	}

	now := time.Now()
	user := model.User{
		UserID:     model.TypeUser.NewID(),
		Type:       model.TypeUser,
		CreatedAt:  now,
		UpdatedAt:  now,
		Screenname: params.Screenname,
	}

	if err := user.UpdatePassword(params.Password); err != nil {
		logger.Error("failed to update password", zap.Error(err))
		s.render.JSON(w, http.StatusInternalServerError, errorMap(err))
		return
	}

	if _, err := s.db.Collection().Doc(user.UserID).Create(r.Context(), user); err != nil {
		logger.Error("failed to create user", zap.Error(err))
		s.render.JSON(w, http.StatusInternalServerError, errorMap(err))
		return
	}

	indexErr := make(chan error)
	go func() { indexErr <- s.search.IndexUser(user) }()

	if err := s.createWorldChatSubscription(r.Context(), user.UserID); err != nil {
		logger.Error("failed to create world chat subscription", zap.Error(err))
	}

	jwt, err := s.newJWTToken(user.UserID)
	if err != nil {
		logger.Error("failed to create jwt", zap.Error(err))
		s.render.JSON(w, http.StatusInternalServerError, errorMap(err))
		return
	}

	session, _ := s.sessions.Get(r, authSessionName)
	session.Values["jwt"] = jwt
	if err := session.Save(r, w); err != nil {
		logger.Error("failed to save session", zap.Error(err))
		s.render.JSON(w, http.StatusInternalServerError, errorMap(err))
		return
	}

	if err := <-indexErr; err != nil {
		logger.Error("failed to add screenname to index", zap.Error(err))
	}

	s.render.JSON(w, http.StatusCreated, util.Map{"user": user})
}

func (s *Server) showCurrentUser(w http.ResponseWriter, r *http.Request) {
	user, _ := model.UserFromContext(r.Context())
	s.render.JSON(w, http.StatusOK, util.Map{"user": user})
}

func (s *Server) showUser(w http.ResponseWriter, r *http.Request) {
	logger := ctxzap.Extract(r.Context())

	userID := chi.URLParam(r, "user_id")
	user, err := db.NewFetcher[model.User](s.db).Fetch(r.Context(), userID)
	if err != nil {
		if err == db.NotFound {
			logger.Info("user not found", zap.String("user_id", userID))
			s.render.JSON(w, http.StatusBadRequest, errorMap(err))
			return
		}

		logger.Error("failed to get user", zap.Error(err))
		s.render.JSON(w, http.StatusInternalServerError, errorMap(err))
		return
	}

	s.render.JSON(w, http.StatusOK, util.Map{"user": user})
}

func (s *Server) searchUsers(w http.ResponseWriter, r *http.Request) {
	logger := ctxzap.Extract(r.Context())

	query := r.URL.Query().Get("query")
	if query == "" {
		s.render.JSON(w, http.StatusOK, util.Map{"users": []model.User{}})
		return
	}

	users, err := s.search.SearchUsers(query)
	if err != nil {
		logger.Error("failed to search screennames index", zap.Error(err))
		s.render.JSON(w, http.StatusInternalServerError, errorMap(err))
		return
	}

	s.render.JSON(w, http.StatusOK, util.Map{"users": users})
}

func (s *Server) createWorldChatSubscription(ctx context.Context, userID string) error {
	worldChat, err := db.NewFetcher[model.Channel](s.db).FetchFirst(ctx, func(query firestore.Query) firestore.Query {
		return query.Where("name", "==", model.WorldChatName)
	})
	if err != nil {
		return errors.Wrap(err, "failed to get world chat")
	}

	now := time.Now()
	worldChatSubscription := model.Subscription{
		SubscriptionID: model.TypeSubscription.NewID(),
		Type:           model.TypeSubscription,
		CreatedAt:      now,
		UpdatedAt:      now,
		UserID:         userID,
		ChannelID:      worldChat.ChannelID,
	}

	if _, err = s.db.Collection().Doc(worldChatSubscription.SubscriptionID).Create(ctx, worldChatSubscription); err != nil {
		return errors.Wrap(err, "failed to create world chat subscription")
	}

	return nil
}
