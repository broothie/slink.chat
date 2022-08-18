package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/broothie/slink.chat/config"
	"github.com/broothie/slink.chat/db"
	"github.com/broothie/slink.chat/model"
	"github.com/gorilla/securecookie"
	_ "github.com/joho/godotenv/autoload"
	"github.com/rs/xid"
)

func main() {
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
		ID:         xid.New().String(),
		CreatedAt:  now,
		UpdatedAt:  now,
		Screenname: "SmarterChild",
	}

	if err := smarterChild.UpdatePassword(string(securecookie.GenerateRandomKey(32))); err != nil {
		fmt.Println("failed to generate password", err)
		os.Exit(1)
	}

	if _, err := db.Collection("users").Doc(smarterChild.ID).Create(context.Background(), smarterChild); err != nil {
		fmt.Println("failed to create SmarterChild", err)
		os.Exit(1)
	}

	// Create World Chat
	worldChat := model.Channel{
		ID:        xid.New().String(),
		CreatedAt: now,
		UpdatedAt: now,
		UserID:    "asdf",
		Name:      model.WorldChatName,
	}

	if _, err := db.Collection("channels").Doc(worldChat.ID).Create(context.Background(), worldChat); err != nil {
		fmt.Println("failed to create World Chat", err)
		os.Exit(1)
	}
}
