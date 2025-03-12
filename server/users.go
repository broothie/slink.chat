package server

import (
	"context"
	"encoding/json"
	"net/http"
	"sort"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/broothie/slink.chat/async/job"
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
	if _, err := db.NewFetcher[model.User](s.DB).FetchFirst(r.Context(), func(query firestore.Query) firestore.Query {
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
		ID:         xid.New().String(),
		CreatedAt:  now,
		UpdatedAt:  now,
		Screenname: params.Screenname,
	}

	if err := user.UpdatePassword(params.Password); err != nil {
		logger.Error("failed to update password", zap.Error(err))
		s.render.JSON(w, http.StatusInternalServerError, errorMap(err))
		return
	}

	if _, err := s.DB.CollectionFor(user.Type()).Doc(user.ID).Create(r.Context(), user); err != nil {
		logger.Error("failed to create user", zap.Error(err))
		s.render.JSON(w, http.StatusInternalServerError, errorMap(err))
		return
	}

	if err := s.Async.Do(r.Context(), job.NewUserJob{UserID: user.ID}); err != nil {
		logger.Error("failed to queue NewUserJob", zap.Error(err))
	}

	jwt, err := s.newJWTToken(user.ID)
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

	s.render.JSON(w, http.StatusCreated, util.Map{"user": user})
}

func (s *Server) showCurrentUser(w http.ResponseWriter, r *http.Request) {
	user, _ := model.UserFromContext(r.Context())
	s.render.JSON(w, http.StatusOK, util.Map{"user": user})
}

func (s *Server) showUser(w http.ResponseWriter, r *http.Request) {
	logger := ctxzap.Extract(r.Context())

	userID := chi.URLParam(r, "user_id")
	user, err := db.NewFetcher[model.User](s.DB).Fetch(r.Context(), userID)
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

func (s *Server) showUsers(w http.ResponseWriter, r *http.Request) {
	logger := ctxzap.Extract(r.Context())

	userIDsQuery := r.URL.Query().Get("user_ids")
	if userIDsQuery == "" {
		s.render.JSON(w, http.StatusOK, util.Map{"users": []model.User{}})
		return
	}

	userIDs := lo.Uniq(strings.Split(userIDsQuery, ","))
	userSlice, err := db.NewFetcher[model.User](s.DB).FetchMany(r.Context(), userIDs...)
	if err != nil {
		logger.Error("failed to fetch users", zap.Error(err))
		s.render.JSON(w, http.StatusInternalServerError, errorMap(err))
		return
	}

	users := lo.Associate(userSlice, func(user model.User) (string, model.User) {
		return user.ID, user
	})

	s.render.JSON(w, http.StatusOK, util.Map{"users": users})
}

func (s *Server) searchUsers(w http.ResponseWriter, r *http.Request) {
	logger := ctxzap.Extract(r.Context())

	query := r.URL.Query().Get("query")
	if query == "" {
		s.render.JSON(w, http.StatusOK, util.Map{"users": []model.User{}})
		return
	}

	users, err := s.Search.SearchUsers(query)
	if err != nil {
		logger.Error("failed to search screennames index", zap.Error(err))
		s.render.JSON(w, http.StatusInternalServerError, errorMap(err))
		return
	}

	user, _ := model.UserFromContext(r.Context())
	users = lo.Reject(users, func(u model.User, _ int) bool { return u.ID == user.ID })
	sort.Slice(users, func(i, j int) bool { return users[i].ID < users[j].ID })
	s.render.JSON(w, http.StatusOK, util.Map{"users": users})
}

func (s *Server) joinWorldChat(ctx context.Context, userID string) error {
	worldChat, err := db.NewFetcher[model.Channel](s.DB).FetchFirst(ctx, func(query firestore.Query) firestore.Query {
		return query.Where("name", "==", model.ChannelNameWorldChat).OrderBy("created_at", firestore.Asc)
	})
	if err != nil {
		return errors.Wrap(err, "failed to get world chat")
	}

	updates := []firestore.Update{{Path: "user_ids", Value: firestore.ArrayUnion(userID)}}
	if _, err = s.DB.CollectionFor(worldChat.Type()).Doc(worldChat.ID).Update(ctx, updates); err != nil {
		return errors.Wrap(err, "failed to create world chat subscription")
	}

	return nil
}
