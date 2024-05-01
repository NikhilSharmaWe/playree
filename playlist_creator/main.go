package main

import (
	"context"
	"log"

	"github.com/NikhilSharmaWe/playree/playlist_creator/app"
	"github.com/joho/godotenv"
	"golang.org/x/sync/errgroup"
)

func init() {
	if err := godotenv.Load("vars.env"); err != nil {
		log.Fatal(err)
	}
}

func main() {
	application, err := app.NewApplication()
	if err != nil {
		log.Fatal(err)
	}

	defer application.ConsumingClient.Close()

	createPlaylistRequestMSGBus, err := setupRabbitMQForStartup(application)
	if err != nil {
		log.Fatal(err)
	}

	g, _ := errgroup.WithContext(context.Background())
	g.SetLimit(50)

	go func() {
		for message := range createPlaylistRequestMSGBus {
			msg := message
			g.Go(func() error {
				if err := handleCreatePlaylistRequests(application, msg); err != nil {
					log.Println("ERROR: ", err)
				}

				return nil
			})
		}
	}()

	log.Fatal(application.MakeCreatePlaylistServerAndRun())
}
