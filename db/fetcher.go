package db

import (
	"context"

	"cloud.google.com/go/firestore"
	"github.com/broothie/slink.chat/model"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"go.uber.org/zap"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var NotFound = errors.New("not found")

type QueryFunc func(query firestore.Query) firestore.Query

type Fetcher[M model.Model] struct {
	db *DB
}

func NewFetcher[M model.Model](db *DB) Fetcher[M] {
	return Fetcher[M]{db: db}
}

func (f Fetcher[Model]) Fetch(ctx context.Context, id string) (Model, error) {
	snapshot, err := f.db.Collection().Doc(id).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return f.Zero(), NotFound
		}

		return f.Zero(), errors.Wrapf(err, "failed to get %q", f.Type())
	}

	model := f.Zero()
	if err := snapshot.DataTo(&model); err != nil {
		return f.Zero(), errors.Wrapf(err, "failed to read %q data", f.Type())
	}

	return model, nil
}

func (f Fetcher[Model]) FetchMany(ctx context.Context, ids ...string) ([]Model, error) {
	logger := ctxzap.Extract(ctx)

	refs := lo.Map(ids, func(id string, _ int) *firestore.DocumentRef { return f.db.Collection().Doc(id) })
	snapshots, err := f.db.GetAll(ctx, refs)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to fetch snapshots for %q", f.Type())
	}

	models := make([]Model, 0, len(snapshots))
	for i, snapshot := range snapshots {
		var model Model
		if err := snapshot.DataTo(&model); err != nil {
			logger.Error("failed to read snapshot", zap.Error(err), zap.Int("i", i), zap.String("type", string(f.Type())))
			continue
		}

		models = append(models, model)
	}

	return models, nil
}

func (f Fetcher[Model]) FetchFirst(ctx context.Context, queryFunc QueryFunc) (Model, error) {
	docs := queryFunc(f.db.Collection().Where("type", "==", f.Type())).Documents(ctx)
	defer docs.Stop()

	snapshot, err := docs.Next()
	if err != nil {
		if err == iterator.Done {
			return f.Zero(), NotFound
		}

		return f.Zero(), errors.Wrapf(err, "failed to fetch snapshot for %q", f.Type())
	}

	var model Model
	if err := snapshot.DataTo(&model); err != nil {
		return f.Zero(), errors.Wrapf(err, "failed to read snapshot for %q", f.Type())
	}

	return model, nil
}

func (f Fetcher[Model]) Query(ctx context.Context, queryFunc QueryFunc) ([]Model, error) {
	logger := ctxzap.Extract(ctx)

	docs := queryFunc(f.db.Collection().Where("type", "==", f.Type())).Documents(ctx)
	defer docs.Stop()

	snapshots, err := docs.GetAll()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to query snapshots for %q", f.Type())
	}

	models := make([]Model, 0, len(snapshots))
	for i, snapshot := range snapshots {
		var model Model
		if err := snapshot.DataTo(&model); err != nil {
			logger.Error("failed to read snapshot", zap.Error(err), zap.Int("i", i), zap.String("type", string(f.Type())))
			continue
		}

		models = append(models, model)
	}

	return models, nil
}

func (f Fetcher[Model]) Zero() Model {
	var zero Model
	return zero
}

func (f Fetcher[Model]) Type() model.Type {
	var zero Model
	return zero.ModelType()
}
