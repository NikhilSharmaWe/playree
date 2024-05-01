package app

import (
	"fmt"
	"net/http"
	"os"

	"github.com/NikhilSharmaWe/playree/playree/models"
	"github.com/NikhilSharmaWe/playree/playree/store"
	"github.com/NikhilSharmaWe/rabbitmq"
	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"github.com/gorilla/websocket"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	amqp "github.com/rabbitmq/amqp091-go"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
)

type Application struct {
	CookieStore *sessions.CookieStore

	SpotifyClientID     string
	SpotifyClientSecret string
	SpotifyRedirectPath string
	Authenticator       *spotifyauth.Authenticator

	MinioClient     *minio.Client
	MinioBucketName string

	UserStore     store.UserStore
	PlaylistStore store.PlaylistStore
	TrackStore    store.TrackStore
	TokenStore    store.TokenStore

	CreatePlaylistResponseClient  *rabbitmq.RabbitClient
	CreatePlaylistResponseChannel map[string]chan models.RabbitMQCreatePlaylistResponse
	PublishingConn                *amqp.Connection
	RabbitMQInstanceID            string

	Upgrader websocket.Upgrader
}

func NewApplication() (*Application, error) {
	db := createSQLDB()

	rc := createRedisClient()

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

	instanceID := uuid.NewString()

	consumingConnection, err := rabbitmq.ConnectRabbitMQ(rabbitMQUser, rabbitMQPassword, rabbitMQAddr, rabbitMQVhost)
	if err != nil {
		return nil, err
	}

	publishingConnection, err := rabbitmq.ConnectRabbitMQ(rabbitMQUser, rabbitMQPassword, rabbitMQAddr, rabbitMQVhost)
	if err != nil {
		return nil, err
	}

	_, err = rabbitmq.CreateNewQueueReturnClient(publishingConnection, "create-playlist-request", true, true)
	if err != nil {
		return nil, err
	}

	createPlaylistResponseClient, err := rabbitmq.CreateNewQueueReturnClient(consumingConnection, "create-playlist-response-"+instanceID, true, true)
	if err != nil {
		return nil, err
	}

	return &Application{
		CookieStore: sessions.NewCookieStore([]byte(os.Getenv("SECRET"))),

		SpotifyClientID:     os.Getenv("CLIENT_ID"),
		SpotifyClientSecret: os.Getenv("CLIENT_SECRET"),
		SpotifyRedirectPath: os.Getenv("REDIRECT_PATH"),
		Authenticator: spotifyauth.New(
			spotifyauth.WithRedirectURL(fmt.Sprintf("http://%s%s", os.Getenv("ADDR"), os.Getenv("REDIRECT_PATH"))),
			spotifyauth.WithScopes(spotifyauth.ScopePlaylistReadPrivate),
			spotifyauth.WithClientID(os.Getenv("CLIENT_ID")),
			spotifyauth.WithClientSecret(os.Getenv("CLIENT_SECRET")),
		),

		MinioClient:     client,
		MinioBucketName: minioBucketName,

		UserStore:     store.NewUserStore(db),
		PlaylistStore: store.NewPlaylistStore(db),
		TrackStore:    store.NewTrackStore(db),
		TokenStore:    store.NewTokenStore(rc, "oauth_tokens"),

		CreatePlaylistResponseClient:  createPlaylistResponseClient,
		CreatePlaylistResponseChannel: make(map[string]chan models.RabbitMQCreatePlaylistResponse),
		PublishingConn:                publishingConnection,

		RabbitMQInstanceID: instanceID,

		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}, nil
}
