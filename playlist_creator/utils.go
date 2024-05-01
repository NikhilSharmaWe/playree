package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/NikhilSharmaWe/playree/playlist_creator/app"
	"github.com/NikhilSharmaWe/playree/playlist_creator/proto"
	"github.com/NikhilSharmaWe/rabbitmq"
	amqp "github.com/rabbitmq/amqp091-go"
)

func setupRabbitMQForStartup(app *app.Application) (<-chan amqp.Delivery, error) {
	if err := app.ConsumingClient.CreateBinding(
		"create-playlist-request",
		"create-playlist-request",
		"create-playlist",
	); err != nil {
		return nil, err
	}

	createPlaylistRequestMSGBus, err := app.ConsumingClient.Consume("create-playlist-request", "create-playlist-service", false)
	if err != nil {
		return nil, err
	}

	return createPlaylistRequestMSGBus, nil
}

func handleCreatePlaylistRequests(application *app.Application, msg amqp.Delivery) error {
	// for consuming and publishing separate connections should be used
	// and for concurrent tasks new channels should be used therefore I am creating new clients here
	publishingClient, err := rabbitmq.NewRabbitMQClient(application.PublishingConn)
	if err != nil {
		return err
	}

	req := proto.CreatePlaylistRequest{}

	if err := json.Unmarshal(msg.Body, &req); err != nil {
		return err
	}

	ctx, cancelFunc := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancelFunc()

	var response *app.RabbitMQCreatePlaylistResponse

	resp, err := application.CreatePlaylistClient.CreatePlaylist(ctx, &req)
	if err != nil {
		log.Println("ERROR: CREATE PLAYLIST: ", err)
		response = &app.RabbitMQCreatePlaylistResponse{
			PlayreePlaylistID: req.PlayreePlaylistId,
			Success:           false,
			Error:             fmt.Sprint("CREATE PLAYLIST SERVICE: ", err.Error()),
		}
	} else {
		response = &app.RabbitMQCreatePlaylistResponse{
			PlayreePlaylistID: resp.PlayreePlaylistId,
			Success:           true,
		}
	}

	body, err := json.Marshal(*response)
	if err != nil {
		return err
	}

	return publishingClient.Send(context.Background(), "create-playlist", msg.ReplyTo, amqp.Publishing{
		ContentType:  "application/json",
		Body:         body,
		DeliveryMode: amqp.Persistent,
	})
}
