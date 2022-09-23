package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"cloud.google.com/go/firestore"
	"github.com/algolia/algoliasearch-client-go/v3/algolia/search"
	"github.com/broothie/slink.chat/config"
	pkgdb "github.com/broothie/slink.chat/db"
	"github.com/broothie/slink.chat/model"
	"github.com/broothie/slink.chat/util"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	environment := flag.String("e", "development", "environment to run in")
	flag.Parse()

	if err := os.Setenv("ENVIRONMENT", *environment); err != nil {
		log.Fatalln("failed to set environment", err)
	}

	cfg, err := config.New()
	if err != nil {
		log.Fatalln("failed to get new config", err)
	}

	db, err := pkgdb.New(cfg)
	if err != nil {
		log.Fatalln("failed to get new db", err)
	}

	var operations []search.BatchOperationIndexed

	log.Println("fetching users")
	userFetcher := pkgdb.NewFetcher[model.User](db)
	users, err := userFetcher.Query(context.Background(), func(query firestore.Query) firestore.Query { return query })
	if err != nil {
		log.Fatalln("failed to fetch users", err)
	}

	log.Println("user count", len(users))
	for _, user := range users {
		operations = append(operations, search.BatchOperationIndexed{
			IndexName: fmt.Sprintf("users-%s", *environment),
			BatchOperation: search.BatchOperation{
				Action: search.AddObject,
				Body: util.Map{
					"objectID":   user.ID,
					"screenname": user.Screenname,
				},
			},
		})
	}

	log.Println("fetching channels")
	channelFetcher := pkgdb.NewFetcher[model.Channel](db)
	channels, err := channelFetcher.Query(context.Background(), func(query firestore.Query) firestore.Query { return query })
	if err != nil {
		log.Fatalln("failed to fetch channels", err)
	}

	log.Println("channel count", len(channels))
	for _, channel := range channels {
		if !channel.Private {
			operations = append(operations, search.BatchOperationIndexed{
				IndexName: fmt.Sprintf("channels-%s", *environment),
				BatchOperation: search.BatchOperation{
					Action: search.AddObject,
					Body: util.Map{
						"objectID": channel.ID,
						"name":     channel.Name,
					},
				},
			})
		}
	}

	log.Println("indexing")
	algolia := search.NewClient(cfg.AlgoliaAppID, cfg.AlgoliaAPIKey)
	if _, err := algolia.MultipleBatch(operations); err != nil {
		log.Fatalln("failed to batch index", err)
	}

	log.Println("done")
}
