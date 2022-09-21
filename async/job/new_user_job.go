package job

import (
	"context"

	"cloud.google.com/go/firestore"
	"github.com/broothie/slink.chat/db"
	"github.com/broothie/slink.chat/model"
	"github.com/pkg/errors"
)

type NewUserJob struct {
	UserID string
}

func (j NewUserJob) Name() string {
	return typeName(j)
}

func (s *Server) NewUserJob(ctx context.Context, payload NewUserJob) error {
	userFetcher := db.NewFetcher[model.User](s.DB)
	user, err := userFetcher.Fetch(ctx, payload.UserID)
	if err != nil {
		return errors.Wrap(err, "failed to fetch user")
	}

	if err := s.Search.IndexUser(user); err != nil {
		return errors.Wrap(err, "failed to index user")
	}

	channelFetcher := db.NewFetcher[model.Channel](s.DB)
	worldChat, err := channelFetcher.FetchFirst(ctx, func(query *firestore.CollectionRef) firestore.Query {
		return query.Where("name", "==", model.WorldChatName).OrderBy("created_at", firestore.Asc)
	})
	if err != nil {
		return errors.Wrap(err, "failed to get world chat")
	}

	updates := []firestore.Update{{Path: "user_ids", Value: firestore.ArrayUnion(payload.UserID)}}
	if _, err = s.DB.CollectionFor(worldChat.Type()).Doc(worldChat.ID).Update(ctx, updates); err != nil {
		return errors.Wrap(err, "failed to create world chat subscription")
	}

	return nil
}
