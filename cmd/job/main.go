package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/broothie/slink.chat/async/job"
	"github.com/broothie/slink.chat/config"
	"github.com/broothie/slink.chat/core"
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

	server := job.NewServer(core)
	core.Logger.Info("running job server", zap.Any("config", cfg))
	if err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), server.Handler()); err != nil {
		core.Logger.Error("server error", zap.Error(err))
	}
}
