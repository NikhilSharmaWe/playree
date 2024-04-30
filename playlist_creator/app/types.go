package app

import "github.com/NikhilSharmaWe/playree/playlist_creator/proto"

// type Track struct {
// 	Name    string `json:"name"`
// 	Artists string `json:"artists"`
// }

type CreatePlaylistRequest struct {
	PlayreePlaylistID string         `json:"playree_playlist_id"`
	Tracks            []*proto.Track `json:"tracks"`
}

type RabbitMQCreatePlaylistResponse struct {
	PlayreePlaylistID string `json:"playree_playlist_id"`
	Success           bool   `json:"success"`
	Error             string `json:"error"`
}
