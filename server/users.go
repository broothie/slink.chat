package server

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/broothie/slink.chat/model"
	"github.com/broothie/slink.chat/util"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/pkg/errors"
	"github.com/rs/xid"
	"go.uber.org/zap"
	"google.golang.org/api/iterator"
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
	snapshots := s.db.
		Collection("users").
		Where("screenname", "==", params.Screenname).
		Limit(1).
		Documents(r.Context())
	defer snapshots.Stop()
	if _, err := snapshots.Next(); err == nil {
		logger.Info("screenname is taken")
		s.render.JSON(w, http.StatusBadRequest, errorMap(errors.New("screenname is taken")))
		return
	} else if err != iterator.Done {
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

	if _, err := s.db.Collection("users").Doc(user.ID).Create(r.Context(), user); err != nil {
		logger.Error("failed to create user", zap.Error(err))
		s.render.JSON(w, http.StatusInternalServerError, errorMap(err))
		return
	}

	if err := s.createWorldChatSubscription(r.Context(), user.ID); err != nil {
		logger.Error("failed to create world chat subscription", zap.Error(err))
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

func (s *Server) showUser(w http.ResponseWriter, r *http.Request) {
	user, _ := model.UserFromContext(r.Context())
	s.render.JSON(w, http.StatusOK, util.Map{"user": user})
}

func (s *Server) createWorldChatSubscription(ctx context.Context, userID string) error {
	docs := s.db.
		Collection("channels").
		Where("name", "==", model.WorldChatName).
		Limit(1).
		Documents(ctx)
	defer docs.Stop()

	doc, err := docs.Next()
	if err != nil {
		return errors.Wrap(err, "failed to get world chat channel")
	}

	now := time.Now()
	worldChatSubscription := model.Subscription{
		ID:        xid.New().String(),
		CreatedAt: now,
		UpdatedAt: now,
		ChannelID: doc.Ref.ID,
	}

	_, err = s.db.
		Collection("users").
		Doc(userID).
		Collection("subscriptions").
		Doc(worldChatSubscription.ID).
		Create(ctx, worldChatSubscription)
	if err != nil {
		return errors.Wrap(err, "failed to create world chat subscription")
	}

	return nil
}
