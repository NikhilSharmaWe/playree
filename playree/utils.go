package main

import (
	"encoding/json"
	"fmt"

	"github.com/NikhilSharmaWe/playree/playree/app"
	"github.com/NikhilSharmaWe/playree/playree/models"
	amqp "github.com/rabbitmq/amqp091-go"
)

func setupCreatePlaylistSvcRabbitMQForStartup(application *app.Application) (<-chan amqp.Delivery, error) {
	if err := application.CreatePlaylistResponseClient.CreateBinding(
		"create-playlist-response-"+application.RabbitMQInstanceID,
		"create-playlist-response-"+application.RabbitMQInstanceID,
		"create-playlist",
	); err != nil {
		return nil, err
	}

	createPlaylistRespMSGBus, err := application.CreatePlaylistResponseClient.Consume("create-playlist-response-"+application.RabbitMQInstanceID, "playree-"+application.RabbitMQInstanceID, false)
	if err != nil {
		return nil, err
	}

	return createPlaylistRespMSGBus, nil
}

func handleRabbitMQResponses(application *app.Application, msg amqp.Delivery) (*models.RabbitMQCreatePlaylistResponse, error) {
	response := &models.RabbitMQCreatePlaylistResponse{}

	if err := json.Unmarshal(msg.Body, &response); err != nil {
		return nil, err
	}

	_, ok := application.CreatePlaylistResponseChannel[response.PlayreePlaylistID]
	if !ok {
		return nil, fmt.Errorf("no create playlist process running with playlist id: %s", response.PlayreePlaylistID)
	}

	return response, nil
}
