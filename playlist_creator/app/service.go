package app

import (
	"fmt"
	"os"
)

type CreatePlaylistService interface {
	CreatePlaylist(CreatePlaylistRequest) error
}

type createPlaylistService struct {
	app *Application
}

func NewCreatePlaylistService(app *Application) CreatePlaylistService {
	return &createPlaylistService{
		app: app,
	}
}

func (svc *createPlaylistService) CreatePlaylist(req CreatePlaylistRequest) error {
	videoIDs, err := svc.app.getYTVideoIDs(req.Tracks)
	if err != nil {
		return err
	}

	if err := svc.app.downloadToAudioLocally(req, videoIDs); err != nil {
		return err
	}

	defer os.RemoveAll(fmt.Sprintf("./local-playlists/%s", req.PlayreePlaylistID))

	return svc.app.pushToMinio(req.PlayreePlaylistID)
}
