package search

import (
	"fmt"

	"github.com/algolia/algoliasearch-client-go/v3/algolia/search"
	"github.com/broothie/slink.chat/config"
	"github.com/broothie/slink.chat/model"
	"github.com/broothie/slink.chat/util"
	"github.com/pkg/errors"
	"github.com/samber/lo"
)

type Algolia struct {
	cfg     *config.Config
	algolia *search.Client
}

func NewAlgolia(cfg *config.Config) *Algolia {
	return &Algolia{
		cfg:     cfg,
		algolia: search.NewClient(cfg.AlgoliaAppID, cfg.AlgoliaAPIKey),
	}
}

func (a *Algolia) IndexUser(user model.User) error {
	if _, err := a.usersIndex().SaveObject(util.Map{"objectID": user.ID, "screenname": user.Screenname}); err != nil {
		return errors.Wrap(err, "failed to update index")
	}

	return nil
}

func (a *Algolia) SearchUsers(query string) ([]model.User, error) {
	result, err := a.usersIndex().Search(query)
	if err != nil {
		return nil, errors.Wrap(err, "failed to search users index")
	}

	users := lo.Map(result.Hits, func(hit map[string]any, _ int) model.User {
		return model.User{
			ID:         hit["objectID"].(string),
			Screenname: hit["screenname"].(string),
		}
	})

	return users, nil
}

func (a *Algolia) DeleteUser(userID string) error {
	_, err := a.usersIndex().DeleteObject(userID)
	return errors.Wrap(err, "deleting user")
}

func (a *Algolia) IndexChannel(channel model.Channel) error {
	if _, err := a.channelsIndex().SaveObject(util.Map{"objectID": channel.ID, "name": channel.Name}); err != nil {
		return errors.Wrap(err, "failed to fetch channels")
	}

	return nil
}

func (a *Algolia) SearchChannels(query string) ([]model.Channel, error) {
	result, err := a.channelsIndex().Search(query)
	if err != nil {
		return nil, errors.Wrap(err, "failed to search channels index")
	}

	channels := lo.Map(result.Hits, func(hit map[string]any, _ int) model.Channel {
		return model.Channel{
			ID:   hit["objectID"].(string),
			Name: hit["name"].(string),
		}
	})

	return channels, nil
}

func (a *Algolia) DeleteChannel(channelID string) error {
	_, err := a.channelsIndex().DeleteObject(channelID)
	return errors.Wrap(err, "deleting channel")
}

func (a *Algolia) usersIndex() *search.Index {
	return a.algolia.InitIndex(fmt.Sprintf("users-%s", a.cfg.Environment))
}

func (a *Algolia) channelsIndex() *search.Index {
	return a.algolia.InitIndex(fmt.Sprintf("channels-%s", a.cfg.Environment))
}
