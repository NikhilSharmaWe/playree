package main

import (
	"context"
	"log"
	"os"

	"github.com/NikhilSharmaWe/playree/playree/app"
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

	e := application.Router()

	createPlaylistRespMSGBus, err := setupCreatePlaylistSvcRabbitMQForStartup(application)
	if err != nil {
		log.Fatal(err)
	}

	createPlaylistG, _ := errgroup.WithContext(context.Background())
	createPlaylistG.SetLimit(50)

	go func() {
		for message := range createPlaylistRespMSGBus {
			msg := message

			createPlaylistG.Go(func() error {
				response, err := handleRabbitMQResponses(application, msg)
				if err != nil {
					log.Println("ERRROR: HANDLING CREATE PLAYLIST RESPONSES: ", err)
				} else {
					application.CreatePlaylistResponseChannel[response.PlayreePlaylistID] <- *response
				}
				return nil
			})
		}
	}()

	log.Fatal(e.Start(os.Getenv("ADDR")))
}
