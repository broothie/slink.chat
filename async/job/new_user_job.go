package job

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/broothie/slink.chat/db"
	"github.com/broothie/slink.chat/model"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/pkg/errors"
	"github.com/rs/xid"
	"github.com/samber/lo"
	"go.uber.org/zap"
)

type NewUserJob struct {
	UserID string
}

func (j NewUserJob) Name() string {
	return typeName(j)
}

func (s *Server) NewUserJob(ctx context.Context, payload NewUserJob) error {
	logger := ctxzap.Extract(ctx).With(zap.String("user_id", payload.UserID))

	userFetcher := db.NewFetcher[model.User](s.DB)
	user, err := userFetcher.Fetch(ctx, payload.UserID)
	if err != nil {
		return errors.Wrap(err, "failed to fetch user")
	}

	if err := s.Search.IndexUser(user); err != nil {
		return errors.Wrap(err, "failed to index user")
	}

	logger.Info("indexed user")

	channelFetcher := db.NewFetcher[model.Channel](s.DB)
	worldChat, err := channelFetcher.FetchFirst(ctx, func(query firestore.Query) firestore.Query {
		return query.Where("name", "==", model.ChannelNameWorldChat).OrderBy("created_at", firestore.Asc)
	})
	if err != nil {
		return errors.Wrap(err, "failed to get world chat")
	}

	updates := []firestore.Update{{Path: "user_ids", Value: firestore.ArrayUnion(payload.UserID)}}
	if _, err = s.DB.CollectionFor(worldChat.Type()).Doc(worldChat.ID).Update(ctx, updates); err != nil {
		return errors.Wrap(err, "failed to create world chat subscription")
	}

	logger.Info("added user to World Chat")

	smarterChild, err := userFetcher.FetchFirst(ctx, func(query firestore.Query) firestore.Query {
		return query.Where("screenname", "==", model.ScreennameSmarterChild)
	})
	if err != nil {
		return errors.Wrapf(err, "failed to find %q", model.ScreennameSmarterChild)
	}

	users := []model.User{user, smarterChild}
	sort.Slice(users, func(i, j int) bool { return users[i].ID < users[j].ID })

	now := time.Now()
	smarterChildChat := model.Channel{
		ID:        xid.New().String(),
		CreatedAt: now,
		UpdatedAt: now,
		Name:      strings.Join(lo.Map(users, func(user model.User, _ int) string { return user.Screenname }), ", "),
		UserID:    smarterChild.ID,
		UserIDs:   lo.Map(users, func(user model.User, _ int) string { return user.ID }),
		Private:   true,
	}

	if _, err := s.DB.CollectionFor(model.TypeChannel).Doc(smarterChildChat.ID).Create(ctx, smarterChildChat); err != nil {
		return errors.Wrapf(err, "failed to create %s chat", model.ScreennameSmarterChild)
	}

	logger.Info(fmt.Sprintf("created %s chat", model.ScreennameSmarterChild))
	return nil
}
