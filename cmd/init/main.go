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
	"github.com/broothie/slink.chat/search"
	"github.com/gorilla/securecookie"
	_ "github.com/joho/godotenv/autoload"
	"github.com/rs/xid"
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

	src := search.NewAlgolia(cfg)

	now := time.Now()
	smarterChild := model.User{
		ID:         xid.New().String(),
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
		ID:        xid.New().String(),
		CreatedAt: now,
		UpdatedAt: now,
		UserID:    smarterChild.ID,
		UserIDs:   []string{smarterChild.ID},
		Name:      model.WorldChatName,
	}

	batch := db.Batch()
	batch.Create(db.CollectionFor(smarterChild.Type()).Doc(smarterChild.ID), smarterChild)
	batch.Create(db.CollectionFor(worldChat.Type()).Doc(worldChat.ID), worldChat)
	if _, err := batch.Commit(context.Background()); err != nil {
		fmt.Println("failed to init db defaults", err)
		os.Exit(1)
	}

	if err := src.IndexUser(smarterChild); err != nil {
		fmt.Println("failed to index smarterchild", err)
		os.Exit(1)
	}

	if err := src.IndexChannel(worldChat); err != nil {
		fmt.Println("failed to index world chat", err)
		os.Exit(1)
	}
}
