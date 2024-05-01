package models

type RabbitMQCreatePlaylistResponse struct {
	PlayreePlaylistID string `json:"playree_playlist_id,omitempty"`
	PlaylistName      string `json:"playlist_name,omitempty"`
	Success           bool   `json:"success,omitempty"`
	Error             string `json:"error,omitempty"`
}
