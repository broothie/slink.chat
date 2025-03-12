package job

import (
	"context"

	"github.com/broothie/slink.chat/model"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
)

type ResetDatabase struct{}

func (j ResetDatabase) Name() string {
	return typeName(j)
}

func (s *Server) ResetDatabase(ctx context.Context) error {
	ctxzap.Info(ctx, "resetting database")

	for _, f := range []func(context.Context) error{s.deleteMessages, s.deleteChannels, s.deleteUsers} {
		if err := f(ctx); err != nil {
			return err
		}
	}

	ctxzap.Info(ctx, "reset database job done")
	return nil
}

func (s *Server) deleteUsers(ctx context.Context) error {
	ctxzap.Info(ctx, "deleting all users")
	docs, err := s.DB.CollectionFor(model.TypeUser).
		Where("screenname", "!=", model.ScreennameSmarterChild).
		Documents(ctx).
		GetAll()
	if err != nil {
		return err
	}

	for _, doc := range docs {
		if _, err := doc.Ref.Delete(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (s *Server) deleteChannels(ctx context.Context) error {
	ctxzap.Info(ctx, "deleting all channels")
	docs, err := s.DB.CollectionFor(model.TypeChannel).
		Where("name", "!=", model.ChannelNameWorldChat).
		Documents(ctx).
		GetAll()
	if err != nil {
		return err
	}

	for _, doc := range docs {
		if _, err := doc.Ref.Delete(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (s *Server) deleteMessages(ctx context.Context) error {
	ctxzap.Info(ctx, "deleting all messages")
	refs, err := s.DB.CollectionFor(model.TypeMessage).DocumentRefs(ctx).GetAll()
	if err != nil {
		return err
	}

	for _, ref := range refs {
		if _, err := ref.Delete(ctx); err != nil {
			return err
		}
	}

	return nil
}
