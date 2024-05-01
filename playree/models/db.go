package models

type UserDBModel struct {
	UserID   string `gorm:"column:user_id;primaryKey"`
	Username string `gorm:"column:username"`
}

type PlaylistsDBModel struct {
	UserID       string `gorm:"column:user_id"`
	PlaylistID   string `gorm:"column:playlist_id"`
	PlaylistName string `gorm:"playlist_name"`
}

type TrackDBModel struct {
	PlaylistID string `gorm:"column:playlist_id"`
	TrackKey   string `gorm:"track_key"`
	TrackURI   string `gorm:"track_uri"`
}
