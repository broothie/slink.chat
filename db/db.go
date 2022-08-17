package db

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	"github.com/broothie/slink.chat/config"
	"github.com/pkg/errors"
)

type DB struct {
	*firestore.Client
	cfg *config.Config
}

func New(cfg *config.Config) (*DB, error) {
	client, err := firestore.NewClient(context.Background(), cfg.ProjectID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create new firestore client")
	}

	return &DB{Client: client, cfg: cfg}, nil
}

func (db *DB) Collection(path string) *firestore.CollectionRef {
	return db.Client.Collection(fmt.Sprintf("%s.%s", db.cfg.Environment, path))
}
