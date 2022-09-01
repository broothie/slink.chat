package server

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/broothie/slink.chat/db"
	"github.com/broothie/slink.chat/model"
	"github.com/broothie/slink.chat/util"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/pkg/errors"
	"github.com/rs/xid"
	"github.com/samber/lo"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

	if _, err := s.db.CollectionFor(channel.Type()).Doc(channel.ID).Create(r.Context(), channel); err != nil {
		logger.Error("failed to create channel", zap.Error(err))
		s.render.JSON(w, http.StatusBadRequest, errorMap(err))
		return
	}

	if !channel.Private {
		if err := s.search.IndexChannel(channel); err != nil {
			logger.Error("failed to update channel search index", zap.Error(err))
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
	if channel, err := db.NewFetcher[model.Channel](s.db).FetchFirst(r.Context(), func(query *firestore.CollectionRef) firestore.Query {
		return query.Where("user_ids", "==", userIDs).Where("private", "==", true)
	}); err == nil {
		logger.Info("chat already exists")
		s.render.JSON(w, http.StatusOK, util.Map{"channel": channel})
		return
	}

	users, err := db.NewFetcher[model.User](s.db).FetchMany(r.Context(), userIDs...)
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

	if _, err := s.db.CollectionFor(channel.Type()).Doc(channel.ID).Create(r.Context(), channel); err != nil {
		logger.Error("failed to create channel", zap.Error(err))
		s.render.JSON(w, http.StatusBadRequest, errorMap(err))
		return
	}

	s.render.JSON(w, http.StatusCreated, util.Map{"channel": channel})
}

func (s *Server) indexChannels(w http.ResponseWriter, r *http.Request) {
	logger := ctxzap.Extract(r.Context())

	user, _ := model.UserFromContext(r.Context())
	channelSlice, err := db.NewFetcher[model.Channel](s.db).Query(r.Context(), func(query *firestore.CollectionRef) firestore.Query {
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

	channel, err := db.NewFetcher[model.Channel](s.db).Fetch(r.Context(), chi.URLParam(r, "channel_id"))
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

	channels, err := s.search.SearchChannels(query)
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
	if _, err := s.db.CollectionFor(model.TypeChannel).Doc(channelID).Update(r.Context(), updates); err != nil {
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
	if _, err := s.db.CollectionFor(model.TypeChannel).Doc(channelID).Update(r.Context(), updates); err != nil {
		logger.Error("failed to delete subscription", zap.Error(err))
		s.render.JSON(w, http.StatusInternalServerError, errorMap(err))
		return
	}

	s.render.JSON(w, http.StatusOK, util.Map{"channelID": channelID})
}

func (s *Server) indexChannelUsers(w http.ResponseWriter, r *http.Request) {
	logger := ctxzap.Extract(r.Context())

	channel, err := db.NewFetcher[model.Channel](s.db).Fetch(r.Context(), chi.URLParam(r, "channel_id"))
	if err != nil {
		if err == db.NotFound {
			s.render.JSON(w, http.StatusBadRequest, errorMap(err))
			return
		}

		logger.Error("failed to fetch channel", zap.Error(err))
		s.render.JSON(w, http.StatusInternalServerError, errorMap(err))
		return
	}

	userSlice, err := db.NewFetcher[model.User](s.db).FetchMany(r.Context(), channel.UserIDs...)
	if err != nil {
		logger.Error("failed to fetch users", zap.Error(err))
		s.render.JSON(w, http.StatusInternalServerError, errorMap(err))
		return
	}

	users := lo.Associate(userSlice, func(user model.User) (string, model.User) { return user.ID, user })
	s.render.JSON(w, http.StatusOK, util.Map{"users": users})
}

func (s *Server) channelSocket(w http.ResponseWriter, r *http.Request) {
	logger := ctxzap.Extract(r.Context())

	user, _ := model.UserFromContext(r.Context())
	channelID := chi.URLParam(r, "channel_id")
	if _, err := db.NewFetcher[model.Channel](s.db).FetchFirst(r.Context(), func(query *firestore.CollectionRef) firestore.Query {
		return query.Where("user_ids", "array-contains", user.ID)
	}); err != nil {
		if err == db.NotFound {
			logger.Info("user not in channel")
			s.render.JSON(w, http.StatusUnauthorized, errorMap(errors.New("user not in channel")))
			return
		}

		logger.Error("error finding subscription")
		s.render.JSON(w, http.StatusInternalServerError, errorMap(err))
		return
	}

	upgrader := &websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error("failed to upgrade request", zap.Error(err))
		s.render.JSON(w, http.StatusInternalServerError, errorMap(err))
		return
	}

	defer func() {
		if err := conn.Close(); err != nil {
			logger.Error("failed to close connection", zap.Error(err))
		}
	}()

	socketCloseChan := make(chan struct{})
	go func() {
		defer close(socketCloseChan)

		for {
			if _, _, err := conn.NextReader(); err != nil {
				if _, isCloseErr := err.(*websocket.CloseError); !isCloseErr {
					logger.Error("next reader error", zap.Error(err))
				}

				return
			}
		}
	}()

	dbCloseChan := make(chan struct{})
	messages := make(chan model.Message)
	go func() {
		defer close(dbCloseChan)

		logger.Debug("listening for messages")
		snapshots := s.db.
			CollectionFor(model.TypeMessage).
			Where("channel_id", "==", channelID).
			Where("created_at", ">", time.Now()).
			Snapshots(r.Context())
		defer snapshots.Stop()

		for {
			snapshot, err := snapshots.Next()
			if err != nil {
				if status.Code(err) == codes.DeadlineExceeded {
					logger.Debug("db listen timeout", zap.Error(err))
					return
				} else if status.Code(err) == codes.Canceled {
					return
				}

				logger.Error("next snapshot error", zap.Error(err))
				return
			}

			if snapshot == nil {
				continue
			}

			change, found := lo.Find(snapshot.Changes, func(change firestore.DocumentChange) bool {
				return change.Kind == firestore.DocumentAdded
			})

			if !found {
				continue
			}

			var message model.Message
			if err := change.Doc.DataTo(&message); err != nil {
				logger.Error("failed to read message", zap.Error(err))
				continue
			}

			messages <- message
		}
	}()

	logger.Debug("socket opened")
	for {
		select {
		case <-socketCloseChan:
			logger.Info("client closed socket")
			return

		case <-dbCloseChan:
			logger.Info("db closed stream")
			return

		case message := <-messages:
			socket, err := conn.NextWriter(websocket.TextMessage)
			if err != nil {
				logger.Error("failed to get writer", zap.Error(err))
				continue
			}

			if err := json.NewEncoder(socket).Encode(message); err != nil {
				logger.Error("failed to write json to socket", zap.Error(err))
				continue
			}

			if err := socket.Close(); err != nil {
				logger.Error("failed to close socket", zap.Error(err))
			}
		}
	}
}
