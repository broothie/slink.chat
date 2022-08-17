package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/broothie/slink.chat/config"
	"github.com/broothie/slink.chat/server"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	fmt.Println("GCP_PROJECT", os.Getenv("GCP_PROJECT"))

	cfg, err := config.New()
	if err != nil {
		fmt.Println("failed to get new config", err)
		os.Exit(1)
	}

	srv, err := server.New(cfg)
	if err != nil {
		fmt.Println("failed to get new server", err)
		os.Exit(1)
	}

	fmt.Println("server running on port", cfg.Port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), srv.Handler()); err != nil {
		fmt.Println("server error", err)
		os.Exit(1)
	}
}
