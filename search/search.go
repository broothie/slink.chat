package search

import (
	"github.com/broothie/slink.chat/model"
)

type Search interface {
	IndexUser(model.User) error
	SearchUsers(string) ([]model.User, error)
	DeleteUser(string) error

	IndexChannel(model.Channel) error
	SearchChannels(string) ([]model.Channel, error)
	DeleteChannel(string) error
}
