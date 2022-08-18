package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/broothie/slink.chat/config"
	"github.com/broothie/slink.chat/db"
	"github.com/broothie/slink.chat/model"
	"github.com/gorilla/securecookie"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	environment := flag.String("e", "development", "environment to run in")
	flag.Parse()

	if err := os.Setenv("ENVIRONMENT", *environment); err != nil {
		fmt.Println("failed to set environment", err)
		os.Exit(1)
	}

	cfg, err := config.New()
	if err != nil {
		fmt.Println("failed to get new config", err)
		os.Exit(1)
	}

	db, err := db.New(cfg)
	if err != nil {
		fmt.Println("failed to get new db", err)
		os.Exit(1)
	}

	now := time.Now()
	smarterChild := model.User{
		UserID:     model.TypeUser.NewID(),
		Type:       model.TypeUser,
		CreatedAt:  now,
		UpdatedAt:  now,
		Screenname: "SmarterChild",
	}

	if err := smarterChild.UpdatePassword(string(securecookie.GenerateRandomKey(32))); err != nil {
		fmt.Println("failed to generate password", err)
		os.Exit(1)
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

	batch := db.Batch()
	batch.Create(db.Collection().Doc(smarterChild.UserID), smarterChild)
	batch.Create(db.Collection().Doc(worldChat.ChannelID), worldChat)
	if _, err := batch.Commit(context.Background()); err != nil {
		fmt.Println("failed to init db defaults", err)
		os.Exit(1)
	}

	if _, err := db.Collection().Doc(worldChat.ChannelID).Create(context.Background(), worldChat); err != nil {
		fmt.Println("failed to create World Chat", err)
		os.Exit(1)
	}
}
