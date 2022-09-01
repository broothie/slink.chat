package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/broothie/slink.chat/config"
	"github.com/broothie/slink.chat/db"
	"github.com/broothie/slink.chat/model"
	"github.com/gorilla/securecookie"
	_ "github.com/joho/godotenv/autoload"
	"github.com/pkg/errors"
)

const (
	batchSize = 500

	//  layout: 2006-01-02T15:04:05Z07:00
	// example: 2022-09-01 01:43:13.703598
	dbLayout = "2006-01-02 15:04:05.000000"
)

type User struct {
	ID             int    `json:"id"`
	Screenname     string `json:"screenname"`
	PasswordDigest string `json:"password_digest"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}

type Channel struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	OwnerID   int    `json:"owner_id"`
	Private   bool   `json:"private"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type Subscription struct {
	UserID    int `json:"user_id"`
	ChannelID int `json:"channel_id"`
}

type Message struct {
	ID        int    `json:"id"`
	AuthorID  int    `json:"author_id"`
	ChannelID int    `json:"channel_id"`
	Timestamp string `json:"timestamp"`
	Body      string `json:"body"`
}

type Create struct {
	dr   *firestore.DocumentRef
	data any
}

func main() {
	environment := flag.String("e", "", "environment to run in")
	flag.Parse()

	if *environment == "" {
		log.Fatalln("missing environment")
	}

	if err := os.Setenv("ENVIRONMENT", *environment); err != nil {
		log.Fatalln("failed to set environment", err)
	}

	cfg, err := config.New()
	if err != nil {
		log.Fatalln("failed to get new config", err)
	}

	db, err := db.New(cfg)
	if err != nil {
		log.Fatalln("failed to get new db", err)
	}

	// batch := db.Batch()

	now := time.Now()
	smarterChild := model.User{
		UserID:     model.TypeUser.NewID(),
		Type:       model.TypeUser,
		CreatedAt:  now,
		UpdatedAt:  now,
		Screenname: "SmarterChild",
	}

	if err := smarterChild.UpdatePassword(string(securecookie.GenerateRandomKey(32))); err != nil {
		log.Fatalln("failed to generate password", err)
	}

	// Create World Chat
	worldChat := model.Channel{
		ChannelID: model.TypeChannel.NewID(),
		Type:      model.TypeChannel,
		CreatedAt: now,
		UpdatedAt: now,
		UserID:    smarterChild.UserID,
		Name:      model.WorldChatName,
	}

	var creates []Create
	creates = append(creates, Create{
		dr:   db.Collection().Doc(smarterChild.UserID),
		data: smarterChild,
	})

	creates = append(creates, Create{
		dr:   db.Collection().Doc(worldChat.ChannelID),
		data: worldChat,
	})

	var oldUsers []User
	if err := readJSONFile(".local/migration/users.slink.json", &oldUsers); err != nil {
		log.Fatalln("failed to decode users file", err)
	}

	userLookup := make(map[int]*model.User)
	for _, oldUser := range oldUsers {
		createdAt, err := time.Parse(dbLayout, oldUser.CreatedAt)
		if err != nil {
			log.Println(err)
			continue
		}

		updatedAt, err := time.Parse(dbLayout, oldUser.UpdatedAt)
		if err != nil {
			log.Println(err)
			continue
		}

		newUser := model.User{
			UserID:         model.TypeUser.NewID(),
			Type:           model.TypeUser,
			CreatedAt:      createdAt,
			UpdatedAt:      updatedAt,
			Screenname:     oldUser.Screenname,
			PasswordDigest: []byte(oldUser.PasswordDigest),
		}

		userLookup[oldUser.ID] = &newUser
		log.Println("user", oldUser.ID, newUser.UserID)
	}

	var oldChannels []Channel
	if err := readJSONFile(".local/migration/channels.slink.json", &oldChannels); err != nil {
		log.Fatalln("failed to decode channels file", err)
	}

	channelLookup := make(map[int]*model.Channel)
	for _, oldChannel := range oldChannels {
		userID := ""
		user := userLookup[oldChannel.OwnerID]
		if user != nil {
			userID = user.UserID
		}

		createdAt, err := time.Parse(dbLayout, oldChannel.CreatedAt)
		if err != nil {
			log.Println(err)
			continue
		}

		updatedAt, err := time.Parse(dbLayout, oldChannel.UpdatedAt)
		if err != nil {
			log.Println(err)
			continue
		}

		newChannel := model.Channel{
			ChannelID: model.TypeChannel.NewID(),
			Type:      model.TypeChannel,
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
			Name:      oldChannel.Name,
			UserID:    userID,
			UserIDs:   nil,
			Private:   oldChannel.Private,
		}

		channelLookup[oldChannel.ID] = &newChannel
		log.Println("channel", oldChannel.ID, newChannel.ChannelID)
	}

	var oldSubscriptions []Subscription
	if err := readJSONFile(".local/migration/subscriptions.slink.json", &oldSubscriptions); err != nil {
		log.Fatalln("failed to decode subscriptions file", err)
	}

	for _, subscription := range oldSubscriptions {
		channel := channelLookup[subscription.ChannelID]
		if channel == nil {
			continue
		}

		user := userLookup[subscription.UserID]
		if user == nil {
			continue
		}

		channel.UserIDs = append(channel.UserIDs, user.UserID)
	}

	var oldMessages []Message
	if err := readJSONFile(".local/migration/messages.slink.json", &oldMessages); err != nil {
		log.Fatalln("failed to decode messages file", err)
	}

	for _, oldMessage := range oldMessages {
		user := userLookup[oldMessage.AuthorID]
		if user == nil {
			continue
		}

		channel := channelLookup[oldMessage.ChannelID]
		if channel == nil {
			continue
		}

		timestamp, err := time.Parse(dbLayout, oldMessage.Timestamp)
		if err != nil {
			log.Println(err)
			continue
		}

		newMessage := model.Message{
			MessageID: model.TypeMessage.NewID(),
			Type:      model.TypeMessage,
			CreatedAt: timestamp,
			UpdatedAt: timestamp,
			UserID:    user.UserID,
			ChannelID: channel.ChannelID,
			Body:      oldMessage.Body,
		}

		creates = append(creates, Create{
			dr:   db.Collection().Doc(newMessage.MessageID),
			data: newMessage,
		})

		log.Println("channel", oldMessage.ID, newMessage.MessageID)
	}

	for _, user := range userLookup {
		creates = append(creates, Create{
			dr:   db.Collection().Doc(user.UserID),
			data: user,
		})
	}

	for _, channel := range channelLookup {
		creates = append(creates, Create{
			dr:   db.Collection().Doc(channel.ChannelID),
			data: channel,
		})
	}

	log.Println("committing!")
	for i := 0; i < len(creates); i += batchSize {
		start := i
		end := i + batchSize
		if end > len(creates) {
			end = len(creates)
		}

		batch := db.Batch()
		for _, create := range creates[start:end] {
			batch.Create(create.dr, create.data)
		}

		log.Println("committing", start, end)
		if _, err := batch.Commit(context.Background()); err != nil {
			log.Fatalln("failed to commit batch", start, end, err)
		}
	}
}

func readJSONFile(filename string, v any) error {
	file, err := os.Open(filename)
	if err != nil {
		return errors.Wrap(err, "failed to open json file")
	}

	if err := json.NewDecoder(file).Decode(v); err != nil {
		return errors.Wrap(err, "failed to decode json file")
	}

	return file.Close()
}
