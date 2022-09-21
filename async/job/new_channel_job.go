package job

import (
	"context"

	"github.com/broothie/slink.chat/db"
	"github.com/broothie/slink.chat/model"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type NewChannelJob struct {
	ChannelID string
}

func (j NewChannelJob) Name() string {
	return typeName(j)
}

func (s *Server) NewChannelJob(ctx context.Context, payload NewChannelJob) error {
	logger := ctxzap.Extract(ctx).With(zap.String("channel_id", payload.ChannelID))

	channelFetcher := db.NewFetcher[model.Channel](s.DB)
	channel, err := channelFetcher.Fetch(ctx, payload.ChannelID)
	if err != nil {
		return errors.Wrap(err, "failed to fetch channel")
	}

	if err := s.Search.IndexChannel(channel); err != nil {
		return errors.Wrap(err, "failed to index channel")
	}

	logger.Info("indexed channel", zap.String("name", channel.Name))
	return nil
}
