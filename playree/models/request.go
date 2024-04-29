package models

type Track struct {
	Name    string `json:"name"`
	Artists string `json:"artists"`
}

type CreatePlaylistRequest struct {
	PlayreePlaylistID string   `json:"playree_playlist_id"`
	Tracks            []*Track `json:"tracks"`
}
