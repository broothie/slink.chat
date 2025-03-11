package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/broothie/slink.chat/config"
	"github.com/broothie/slink.chat/core"
	"github.com/broothie/slink.chat/server"
	_ "github.com/joho/godotenv/autoload"
	"go.uber.org/zap"
)

func main() {
	cfg, err := config.New()
	if err != nil {
		log.Fatalln("failed to get new config", err)
	}

	core, err := core.New(cfg)
	if err != nil {
		log.Fatalln("failed to get new core", err)
	}

	server, err := server.New(core)
	if err != nil {
		core.Logger.Error("failed to get new server", zap.Error(err))
		os.Exit(1)
	}

	core.Logger.Info("server running on port", zap.Any("config", cfg))
	if err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), server.Handler()); err != nil {
		core.Logger.Error("server error", zap.Error(err))
	}
}
