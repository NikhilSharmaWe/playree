package app

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/NikhilSharmaWe/playree/playlist_creator/proto"
	"github.com/NikhilSharmaWe/rabbitmq"
	ytdl "github.com/NikhilSharmaWe/youtube/downloader"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

type Application struct {
	Addr                 string
	YTService            *youtube.Service
	MinioClient          *minio.Client
	MinioBucketName      string
	ConsumingClient      *rabbitmq.RabbitClient
	PublishingConn       *amqp.Connection
	CreatePlaylistClient proto.CreatePlaylistServiceClient
}

func NewApplication() (*Application, error) {
	addr := os.Getenv("ADDR")

	ytService, err := youtube.NewService(context.Background(), option.WithAPIKey(os.Getenv("YT_API_KEY")))
	if err != nil {
		return nil, err
	}

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
		YTService:            ytService,
		MinioClient:          client,
		MinioBucketName:      minioBucketName,
		ConsumingClient:      consumingClient,
		PublishingConn:       publishingConn,
		CreatePlaylistClient: createPlaylistClient,
	}, nil
}

func (app *Application) getYTVideoIDs(tracks []*proto.Track) ([]string, error) {
	videoIDs := []string{}

	for _, track := range tracks {
		query := fmt.Sprintf("%s : %s", track.Name, track.Artists[1:len(track.Artists)-1])

		call := app.YTService.Search.List([]string{"id"}).
			Q(query).
			MaxResults(1).
			Order("relevance")
		response, err := call.Do()
		if err != nil {
			return nil, err
		}

		for _, item := range response.Items {
			switch item.Id.Kind {
			case "youtube#video":
				videoIDs = append(videoIDs, item.Id.VideoId)
			default:
				return nil, err
			}
		}
	}

	return videoIDs, nil
}

func (app *Application) downloadToAudioLocally(req CreatePlaylistRequest, videoIDs []string) error {
	outputDir := fmt.Sprintf("./local-playlists/%s", req.PlayreePlaylistID)

	downloader := ytdl.GetDownloader(outputDir)

	for i, videoID := range videoIDs {
		filename := fmt.Sprintf("%s_%s_%s.mp3", strconv.Itoa(i), req.Tracks[i].Name, req.Tracks[i].Artists)
		video, _, err := downloader.GetVideoWithFormat(videoID, outputDir)
		if err != nil {
			return err
		}

		if err := downloader.DownloadAudio(context.Background(), fmt.Sprintf("%s/%s", outputDir, filename), video, "", ""); err != nil {
			return err
		}
	}

	return nil
}

func (app *Application) pushToMinio(playreePlaylistID string) error {
	folderPath := fmt.Sprintf("./local-playlists/%s", playreePlaylistID)

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	return filepath.WalkDir(folderPath, func(path string, d fs.DirEntry, err error) error {
		fileInfo, err := os.Stat(path)
		if err != nil {
			return err
		}

		if fileInfo.IsDir() {
			return nil
		}

		absolutefilePath := cwd + "/" + strings.TrimPrefix(path, "./")
		key := strings.TrimPrefix(path, "local-playlists/")

		_, err = app.MinioClient.FPutObject(context.Background(), app.MinioBucketName, key, absolutefilePath, minio.PutObjectOptions{})

		return err
	})
}
