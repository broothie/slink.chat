package job

import (
	"context"

	"github.com/broothie/slink.chat/model"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
)

type ResetDB struct{}

func (j ResetDB) Name() string {
	return typeName(j)
}

func (s *Server) ResetDB(ctx context.Context) error {
	logger := ctxzap.Extract(ctx)
	logger.Info("deleting all messages")

	refs, err := s.DB.CollectionFor(model.TypeMessage).DocumentRefs(ctx).GetAll()
	if err != nil {
		return err
	}

	for _, ref := range refs {
		if _, err := ref.Delete(ctx); err != nil {
			return err
		}
	}

	logger.Info("hourly job done")
	return nil
}
