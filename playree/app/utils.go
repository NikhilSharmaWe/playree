package app

import (
	"context"
	"html/template"
	"io"
	"time"

	"github.com/NikhilSharmaWe/playree/playree/models"
	"github.com/NikhilSharmaWe/playree/playree/store"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/minio/minio-go/v7"
	"github.com/zmb3/spotify/v2"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

type TemplateRegistry struct {
	templates *template.Template
}

// Implement e.Renderer interface
func (t *TemplateRegistry) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}
func (app *Application) updateTokenFromClientIfNeeded(token *oauth2.Token, client *spotify.Client, userID string) error {
	updatedToken, err := client.Token()
	if err != nil {
		return err
	}

	if updatedToken.AccessToken != token.AccessToken {
		if err := app.TokenStore.Update(context.Background(), userID, updatedToken); err != nil {
			return err
		}
	}

	return nil
}

func getNameAndTracksFromPlaylist(client *spotify.Client, playlistID string) ([]*models.Track, string, error) {
	data := []*models.Track{}

	playlist, err := client.GetPlaylist(context.Background(), spotify.ID(playlistID))
	if err != nil {
		return nil, "", err
	}

	for _, track := range playlist.Tracks.Tracks {
		var artists string

		for _, artist := range track.Track.SimpleTrack.Artists {
			artists = artists + "$" + artist.Name + "@"
		}
		artists = artists[:len(artists)-1] + "$"

		data = append(data, &models.Track{
			Name:    track.Track.Name,
			Artists: artists,
		})
	}

	return data, playlist.Name, nil
}

func (app *Application) handleAfterPlaylistCreated(playlist *models.PlaylistsDBModel) error {
	db := app.TrackStore.DB()
	return db.Transaction(func(tx *gorm.DB) error {
		playlistStore := store.NewPlaylistStore(db)
		tokenStore := store.NewTrackStore(db)

		if err := playlistStore.Create(*playlist); err != nil {
			return err
		}

		data, err := app.generatePresignedURIsForPlaylistTracks(playlist.PlaylistID)
		if err != nil {
			return err
		}

		tracks := []models.TrackDBModel{}
		for key, uri := range data {
			tracks = append(tracks, models.TrackDBModel{
				PlaylistID: playlist.PlaylistID,
				TrackKey:   key,
				TrackURI:   uri,
			})
		}

		return tokenStore.CreateInBatches(tracks)
	})
}

func (app *Application) generatePresignedURIsForPlaylistTracks(playreePlaylistID string) (map[string]string, error) {
	data := make(map[string]string)

	keys, err := app.getListOfAllFiles(playreePlaylistID)
	if err != nil {
		return nil, err
	}

	for _, key := range keys {
		url, err := app.MinioClient.PresignedGetObject(context.Background(), "playree-playlists", key, 7*24*time.Hour, nil)
		if err != nil {
			return nil, err
		}

		data[key] = url.String()
	}

	return data, nil
}

func (app *Application) shouldUpdatePresignedURIs(playreePlaylistID string) (bool, error) {
	now := time.Now()
	fiveDaysAgo := now.AddDate(0, 0, -5)

	exists, err := app.TrackStore.IsExists("playlist_id = ? AND inserted_at <= ?", playreePlaylistID, fiveDaysAgo)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (app *Application) getListOfAllFiles(playreePlaylistID string) ([]string, error) {
	objKeys := []string{}

	objectCh := app.MinioClient.ListObjects(context.Background(), app.MinioBucketName, minio.ListObjectsOptions{
		Prefix:    playreePlaylistID,
		Recursive: true,
	})

	for object := range objectCh {
		if object.Err != nil {
			return objKeys, object.Err
		}

		objKeys = append(objKeys, object.Key)
	}

	return objKeys, nil
}

func sendMessageToFrontend(conn *websocket.Conn, msg string) {
	conn.WriteMessage(1, []byte(msg))
}

func sendFailStatusToFrontend(conn *websocket.Conn) {
	conn.WriteMessage(1, []byte("Error: creating process failed"))
}
