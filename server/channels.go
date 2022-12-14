package server

import (
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
	"github.com/rs/xid"
	"github.com/samber/lo"
	"go.uber.org/zap"
)

func (s *Server) createChannel(w http.ResponseWriter, r *http.Request) {
	logger := ctxzap.Extract(r.Context())

	var params model.Channel
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		logger.Error("failed to decode channel", zap.Error(err))
		s.render.JSON(w, http.StatusBadRequest, errorMap(err))
		return
	}

	user, _ := model.UserFromContext(r.Context())
	now := time.Now()
	channel := model.Channel{
		ID:        xid.New().String(),
		CreatedAt: now,
		UpdatedAt: now,
		Name:      params.Name,
		UserID:    user.ID,
		UserIDs:   []string{user.ID},
		Private:   params.Private,
	}

	if _, err := s.DB.CollectionFor(channel.Type()).Doc(channel.ID).Create(r.Context(), channel); err != nil {
		logger.Error("failed to create channel", zap.Error(err))
		s.render.JSON(w, http.StatusBadRequest, errorMap(err))
		return
	}

	if !channel.Private {
		if err := s.Async.Do(r.Context(), job.NewChannelJob{ChannelID: channel.ID}); err != nil {
			logger.Error("failed to queue NewChannelJob", zap.Error(err))
		}
	}

	s.render.JSON(w, http.StatusCreated, util.Map{"channel": channel})
}

func (s *Server) upsertChat(w http.ResponseWriter, r *http.Request) {
	logger := ctxzap.Extract(r.Context())

	var userIDs []string
	if err := json.NewDecoder(r.Body).Decode(&userIDs); err != nil {
		logger.Error("failed to decode user ids", zap.Error(err))
		s.render.JSON(w, http.StatusBadRequest, errorMap(err))
		return
	}

	user, _ := model.UserFromContext(r.Context())
	if lo.NoneBy(userIDs, func(userID string) bool { return userID == user.ID }) {
		userIDs = append(userIDs, user.ID)
	}

	sort.Slice(userIDs, func(i, j int) bool { return userIDs[i] < userIDs[j] })
	if channel, err := db.NewFetcher[model.Channel](s.DB).FetchFirst(r.Context(), func(query firestore.Query) firestore.Query {
		return query.Where("user_ids", "==", userIDs).Where("private", "==", true)
	}); err == nil {
		logger.Info("chat already exists")
		s.render.JSON(w, http.StatusOK, util.Map{"channel": channel})
		return
	}

	users, err := db.NewFetcher[model.User](s.DB).FetchMany(r.Context(), userIDs...)
	if err != nil {
		logger.Error("failed to fetch users", zap.Error(err))
		s.render.JSON(w, http.StatusInternalServerError, errorMap(err))
		return
	}

	sort.Slice(users, func(i, j int) bool { return users[i].ID < users[j].ID })
	name := strings.Join(lo.Map(users, func(user model.User, _ int) string { return user.Screenname }), ", ")

	now := time.Now()
	channel := model.Channel{
		ID:        xid.New().String(),
		CreatedAt: now,
		UpdatedAt: now,
		Name:      name,
		UserID:    user.ID,
		UserIDs:   userIDs,
		Private:   true,
	}

	if _, err := s.DB.CollectionFor(channel.Type()).Doc(channel.ID).Create(r.Context(), channel); err != nil {
		logger.Error("failed to create channel", zap.Error(err))
		s.render.JSON(w, http.StatusBadRequest, errorMap(err))
		return
	}

	s.render.JSON(w, http.StatusCreated, util.Map{"channel": channel})
}

func (s *Server) indexChannels(w http.ResponseWriter, r *http.Request) {
	logger := ctxzap.Extract(r.Context())

	user, _ := model.UserFromContext(r.Context())
	channelSlice, err := db.NewFetcher[model.Channel](s.DB).Query(r.Context(), func(query firestore.Query) firestore.Query {
		return query.Where("user_ids", "array-contains", user.ID)
	})
	if err != nil {
		logger.Error("failed to fetch channels", zap.Error(err))
		s.render.JSON(w, http.StatusInternalServerError, errorMap(err))
		return
	}

	channels := lo.Associate(channelSlice, func(channel model.Channel) (string, model.Channel) { return channel.ID, channel })
	s.render.JSON(w, http.StatusOK, util.Map{"channels": channels})
}

func (s *Server) showChannel(w http.ResponseWriter, r *http.Request) {
	logger := ctxzap.Extract(r.Context())

	channel, err := db.NewFetcher[model.Channel](s.DB).Fetch(r.Context(), chi.URLParam(r, "channel_id"))
	if err != nil {
		if err == db.NotFound {
			s.render.JSON(w, http.StatusBadRequest, errorMap(err))
			return
		}

		logger.Error("failed to fetch channel", zap.Error(err))
		s.render.JSON(w, http.StatusInternalServerError, errorMap(err))
		return
	}

	s.render.JSON(w, http.StatusOK, util.Map{"channel": channel})
}

func (s *Server) searchChannels(w http.ResponseWriter, r *http.Request) {
	logger := ctxzap.Extract(r.Context())

	query := r.URL.Query().Get("query")
	if query == "" {
		s.render.JSON(w, http.StatusOK, util.Map{"channels": []model.Channel{}})
		return
	}

	channels, err := s.Search.SearchChannels(query)
	if err != nil {
		logger.Error("failed to search channels", zap.Error(err))
		s.render.JSON(w, http.StatusInternalServerError, errorMap(err))
		return
	}

	s.render.JSON(w, http.StatusOK, util.Map{"channels": channels})
}

func (s *Server) joinChannel(w http.ResponseWriter, r *http.Request) {
	logger := ctxzap.Extract(r.Context())

	user, _ := model.UserFromContext(r.Context())
	channelID := chi.URLParam(r, "channel_id")
	updates := []firestore.Update{{Path: "user_ids", Value: firestore.ArrayUnion(user.ID)}}
	if _, err := s.DB.CollectionFor(model.TypeChannel).Doc(channelID).Update(r.Context(), updates); err != nil {
		logger.Error("failed to create subscription", zap.Error(err))
		s.render.JSON(w, http.StatusInternalServerError, errorMap(err))
		return
	}

	s.render.JSON(w, http.StatusCreated, util.Map{"channelID": channelID})
}

func (s *Server) leaveChannel(w http.ResponseWriter, r *http.Request) {
	logger := ctxzap.Extract(r.Context())

	user, _ := model.UserFromContext(r.Context())
	channelID := chi.URLParam(r, "channel_id")
	updates := []firestore.Update{{Path: "user_ids", Value: firestore.ArrayRemove(user.ID)}}
	if _, err := s.DB.CollectionFor(model.TypeChannel).Doc(channelID).Update(r.Context(), updates); err != nil {
		logger.Error("failed to delete subscription", zap.Error(err))
		s.render.JSON(w, http.StatusInternalServerError, errorMap(err))
		return
	}

	s.render.JSON(w, http.StatusOK, util.Map{"channelID": channelID})
}

func (s *Server) indexChannelUsers(w http.ResponseWriter, r *http.Request) {
	logger := ctxzap.Extract(r.Context())

	channel, err := db.NewFetcher[model.Channel](s.DB).Fetch(r.Context(), chi.URLParam(r, "channel_id"))
	if err != nil {
		if err == db.NotFound {
			s.render.JSON(w, http.StatusBadRequest, errorMap(err))
			return
		}

		logger.Error("failed to fetch channel", zap.Error(err))
		s.render.JSON(w, http.StatusInternalServerError, errorMap(err))
		return
	}

	userSlice, err := db.NewFetcher[model.User](s.DB).FetchMany(r.Context(), channel.UserIDs...)
	if err != nil {
		logger.Error("failed to fetch users", zap.Error(err))
		s.render.JSON(w, http.StatusInternalServerError, errorMap(err))
		return
	}

	users := lo.Associate(userSlice, func(user model.User) (string, model.User) { return user.ID, user })
	s.render.JSON(w, http.StatusOK, util.Map{"users": users})
}
