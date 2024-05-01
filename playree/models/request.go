package models

type Track struct {
	Name    string `json:"name,omitempty"`
	Artists string `json:"artists,omitempty"`
}

type CreatePlaylistRequest struct {
	PlayreePlaylistID string   `json:"playree_playlist_id,omitempty"`
	Tracks            []*Track `json:"tracks,omitempty"`
}
