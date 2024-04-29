package models

type RabbitMQCreatePlaylistResponse struct {
	PlayreePlaylistID string `json:"playree_playlist_id"`
	Success           bool   `json:"success"`
	Error             string `json:"error"`
}
