package job

import (
	"context"

	"github.com/broothie/slink.chat/model"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/pkg/errors"
	"google.golang.org/api/iterator"
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
	docs := s.DB.CollectionFor(model.TypeUser).
		Where("screenname", "!=", model.ScreennameSmarterChild).
		Documents(ctx)

	for {
		doc, err := docs.Next()
		if errors.Is(err, iterator.Done) {
			break
		} else if err != nil {
			return errors.Wrap(err, "iterating over message docs")
		}

		if _, err := doc.Ref.Delete(ctx); err != nil {
			return errors.Wrap(err, "deleting message doc")
		}
	}

	return nil
}

func (s *Server) deleteChannels(ctx context.Context) error {
	ctxzap.Info(ctx, "deleting all channels")
	docs := s.DB.CollectionFor(model.TypeChannel).
		Where("name", "!=", model.ChannelNameWorldChat).
		Documents(ctx)

	for {
		doc, err := docs.Next()
		if errors.Is(err, iterator.Done) {
			break
		} else if err != nil {
			return errors.Wrap(err, "iterating over channel docs")
		}

		if _, err := doc.Ref.Delete(ctx); err != nil {
			return errors.Wrap(err, "deleting channel doc")
		}
	}

	return nil
}

func (s *Server) deleteMessages(ctx context.Context) error {
	ctxzap.Info(ctx, "deleting all messages")
	refs := s.DB.CollectionFor(model.TypeMessage).DocumentRefs(ctx)

	for {
		ref, err := refs.Next()
		if errors.Is(err, iterator.Done) {
			break
		} else if err != nil {
			return errors.Wrap(err, "iterating over message refs")
		}

		if _, err := ref.Delete(ctx); err != nil {
			return errors.Wrap(err, "deleting message ref")
		}
	}

	return nil
}
