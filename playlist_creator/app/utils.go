package app

import (
	"os"

	"github.com/NikhilSharmaWe/playree/playlist_creator/proto"
	"github.com/NikhilSharmaWe/rabbitmq"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Application struct {
	Addr                 string
	MinioClient          *minio.Client
	MinioBucketName      string
	ConsumingClient      *rabbitmq.RabbitClient
	PublishingConn       *amqp.Connection
	CreatePlaylistClient proto.CreatePlaylistServiceClient
}

func NewApplication() (*Application, error) {
	addr := os.Getenv("ADDR")
	minioServerAddr := os.Getenv("MINIO_SERVER_ADDR")
	minioAccessKey := os.Getenv("MINIO_ACCESS_KEY")
	minioSecretKey := os.Getenv("MINIO_SECRET_KEY")
	minioBucketName := os.Getenv("MINIO_BUCKET_NAME")

	client, err := minio.New(minioServerAddr, &minio.Options{
		Creds: credentials.NewStaticV4(minioAccessKey, minioSecretKey, ""),
	})
	if err != nil {
		return nil, err
	}

	rabbitMQUser := os.Getenv("RABBITMQ_USER")
	rabbitMQPassword := os.Getenv("RABBITMQ_PASSWORD")
	rabbitMQVhost := os.Getenv("RABBITMQ_VHOST")
	rabbitMQAddr := os.Getenv("RABBITMQ_ADDR")

	// each concurrent task should be done with new channel
	// different connections should be used for publishing and consuming
	consumingConn, err := rabbitmq.ConnectRabbitMQ(rabbitMQUser, rabbitMQPassword, rabbitMQAddr, rabbitMQVhost)
	if err != nil {
		return nil, err
	}

	consumingClient, err := rabbitmq.NewRabbitMQClient(consumingConn)
	if err != nil {
		return nil, err
	}

	publishingConn, err := rabbitmq.ConnectRabbitMQ(rabbitMQUser, rabbitMQPassword, rabbitMQAddr, rabbitMQVhost)
	if err != nil {
		return nil, err
	}

	_, err = rabbitmq.CreateNewQueueReturnClient(consumingConn, "create-playlist-request", true, true)
	if err != nil {
		return nil, err
	}

	createPlaylistClient, err := NewCreatePlaylistClient(addr)
	if err != nil {
		return nil, err
	}

	return &Application{
		Addr:                 addr,
		MinioClient:          client,
		MinioBucketName:      minioBucketName,
		ConsumingClient:      consumingClient,
		PublishingConn:       publishingConn,
		CreatePlaylistClient: createPlaylistClient,
	}, nil
}
