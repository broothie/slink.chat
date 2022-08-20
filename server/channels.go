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
		ChannelID: model.TypeChannel.NewID(),
		Type:      model.TypeChannel,
		CreatedAt: now,
		UpdatedAt: now,
		UserID:    user.UserID,
		Name:      params.Name,
		Private:   params.Private,
	}

	subscription := model.Subscription{
		SubscriptionID: model.TypeSubscription.NewID(),
		Type:           model.TypeSubscription,
		CreatedAt:      now,
		UpdatedAt:      now,
		UserID:         user.UserID,
		ChannelID:      channel.ChannelID,
	}

	batch := s.db.Batch()
	batch.Create(s.db.Collection().Doc(channel.ChannelID), channel)
	batch.Create(s.db.Collection().Doc(subscription.SubscriptionID), subscription)
	if _, err := batch.Commit(r.Context()); err != nil {
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

	users, err := db.NewFetcher[model.User](s.db).FetchMany(r.Context(), userIDs...)
	if err != nil {
		logger.Error("failed to fetch users", zap.Error(err))
		s.render.JSON(w, http.StatusInternalServerError, errorMap(err))
		return
	}

	user, _ := model.UserFromContext(r.Context())
	if lo.NoneBy(users, func(u model.User) bool { return u.UserID == user.UserID }) {
		users = append(users, user)
	}

	sort.Slice(users, func(i, j int) bool { return users[i].UserID < users[j].UserID })
	name := strings.Join(lo.Map(users, func(user model.User, _ int) string { return user.Screenname }), ", ")
	if channel, err := db.NewFetcher[model.Channel](s.db).FetchFirst(r.Context(), func(query firestore.Query) firestore.Query {
		return query.Where("name", "==", name)
	}); err == nil {
		logger.Info("chat already exists")
		s.render.JSON(w, http.StatusOK, util.Map{"channel": channel})
		return
	}

	now := time.Now()
	channel := model.Channel{
		ChannelID: model.TypeChannel.NewID(),
		Type:      model.TypeChannel,
		CreatedAt: now,
		UpdatedAt: now,
		UserID:    user.UserID,
		Name:      name,
		Private:   true,
	}

	batch := s.db.Batch()
	batch.Create(s.db.Collection().Doc(channel.ChannelID), channel)
	for _, user := range users {
		subscription := model.Subscription{
			SubscriptionID: model.TypeSubscription.NewID(),
			Type:           model.TypeSubscription,
			CreatedAt:      now,
			UpdatedAt:      now,
			UserID:         user.UserID,
			ChannelID:      channel.ChannelID,
		}

		batch.Create(s.db.Collection().Doc(subscription.SubscriptionID), subscription)
	}

	if _, err := batch.Commit(r.Context()); err != nil {
		logger.Error("failed to create channel", zap.Error(err))
		s.render.JSON(w, http.StatusBadRequest, errorMap(err))
		return
	}

	s.render.JSON(w, http.StatusCreated, util.Map{"channel": channel})
}

func (s *Server) indexChannels(w http.ResponseWriter, r *http.Request) {
	logger := ctxzap.Extract(r.Context())

	user, _ := model.UserFromContext(r.Context())
	subscriptions, err := db.NewFetcher[model.Subscription](s.db).Query(r.Context(), func(query firestore.Query) firestore.Query {
		return query.Where("user_id", "==", user.UserID)
	})
	if err != nil {
		logger.Error("failed to get subscriptions", zap.Error(err))
		s.render.JSON(w, http.StatusInternalServerError, errorMap(err))
		return
	}

	channelIDs := lo.Map(subscriptions, func(subscription model.Subscription, _ int) string {
		return subscription.ChannelID
	})

	channelSlice, err := db.NewFetcher[model.Channel](s.db).FetchMany(r.Context(), channelIDs...)
	if err != nil {
		logger.Error("failed to get channels", zap.Error(err))
		s.render.JSON(w, http.StatusInternalServerError, errorMap(err))
		return
	}

	channels := lo.Associate(channelSlice, func(channel model.Channel) (string, model.Channel) {
		return channel.ChannelID, channel
	})

	s.render.JSON(w, http.StatusOK, util.Map{"channels": channels})
}

func (s *Server) showChannel(w http.ResponseWriter, r *http.Request) {
	logger := ctxzap.Extract(r.Context())

	docs := s.db.Collection().Where("channel_id", "==", chi.URLParam(r, "channel_id")).Documents(r.Context())
	defer docs.Stop()
	snapshots, err := docs.GetAll()
	if err != nil {
		logger.Error("failed to get channel data", zap.Error(err))
		s.render.JSON(w, http.StatusInternalServerError, errorMap(err))
		return
	}

	var channel model.Channel
	messages := make(map[string]model.Message)
	userIDs := make([]string, 0, len(snapshots))
	for _, snapshot := range snapshots {
		typ, err := snapshot.DataAt("type")
		if err != nil {
			logger.Error("failed to get data type", zap.Error(err), zap.String("id", snapshot.Ref.ID))
			continue
		}

		switch model.Type(typ.(string)) {
		case model.TypeChannel:
			if err := snapshot.DataTo(&channel); err != nil {
				logger.Error("failed to get channel", zap.Error(err), zap.String("id", snapshot.Ref.ID))
				continue
			}

		case model.TypeMessage:
			var message model.Message
			if err := snapshot.DataTo(&message); err != nil {
				logger.Error("failed to get message", zap.Error(err), zap.String("id", snapshot.Ref.ID))
				continue
			}

			messages[message.MessageID] = message

		case model.TypeSubscription:
			userID, err := snapshot.DataAt("user_id")
			if err != nil {
				logger.Error("failed to get user id from subscription", zap.Error(err), zap.String("id", snapshot.Ref.ID))
				continue
			}

			userIDs = append(userIDs, userID.(string))
		}
	}

	usersSlice, err := db.NewFetcher[model.User](s.db).FetchMany(r.Context(), userIDs...)
	if err != nil {
		logger.Error("failed to get users", zap.Error(err))
		s.render.JSON(w, http.StatusInternalServerError, errorMap(err))
		return
	}

	users := lo.Associate(usersSlice, func(user model.User) (string, model.User) {
		return user.UserID, user
	})

	s.render.JSON(w, http.StatusOK, util.Map{"channel": channel, "messages": messages, "users": users})
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

func (s *Server) channelSocket(w http.ResponseWriter, r *http.Request) {
	logger := ctxzap.Extract(r.Context())

	user, _ := model.UserFromContext(r.Context())
	channelID := chi.URLParam(r, "channel_id")
	if _, err := db.NewFetcher[model.Subscription](s.db).FetchFirst(r.Context(), func(query firestore.Query) firestore.Query {
		return query.Where("user_id", "==", user.UserID).Where("channel_id", "==", channelID)
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
			Collection().
			Where("type", "==", model.TypeMessage).
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
