package search

import (
	"context"
	"strings"

	"cloud.google.com/go/firestore"
	pkgdb "github.com/broothie/slink.chat/db"
	"github.com/broothie/slink.chat/model"
	"github.com/pkg/errors"
	"github.com/samber/lo"
)

type DB struct {
	db *pkgdb.DB
}

func NewDB(db *pkgdb.DB) *DB {
	return &DB{db: db}
}

func (db *DB) IndexUser(model.User) error {
	return nil
}

func (db *DB) SearchUsers(query string) ([]model.User, error) {
	users, err := pkgdb.NewFetcher[model.User](db.db).Query(context.Background(), func(query firestore.Query) firestore.Query { return query.Where("screenname", ">", "") })
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch users")
	}

	query = strings.ToLower(query)
	return lo.Filter(users, func(user model.User, _ int) bool { return strings.HasPrefix(strings.ToLower(user.Screenname), query) }), nil
}

func (db *DB) DeleteUser(userID string) error {
	_, err := db.db.Doc(userID).Delete(context.Background())
	return errors.Wrap(err, "deleting user from db")
}

func (db *DB) IndexChannel(user model.Channel) error {
	return nil
}

func (db *DB) SearchChannels(query string) ([]model.Channel, error) {
	channels, err := pkgdb.NewFetcher[model.Channel](db.db).Query(context.Background(), func(query firestore.Query) firestore.Query { return query.Where("screenname", ">", "") })
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch channels")
	}

	query = strings.ToLower(query)
	return lo.Filter(channels, func(channel model.Channel, _ int) bool {
		return strings.HasPrefix(strings.ToLower(channel.Name), query)
	}), nil
}

func (db *DB) DeleteChannel(channelID string) error {
	_, err := db.db.Doc(channelID).Delete(context.Background())
	return errors.Wrap(err, "deleting channel from db")
}
