package app

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
	return nil
}
