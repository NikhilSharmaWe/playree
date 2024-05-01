package models

import "time"

type UserDBModel struct {
	UserID   string `gorm:"column:user_id;primaryKey"`
	Username string `gorm:"column:username"`
}

type PlaylistsDBModel struct {
	UserID       string `gorm:"column:user_id"`
	PlaylistID   string `gorm:"column:playlist_id"`
	PlaylistName string `gorm:"column:playlist_name"`
}

// type TrackDBModel struct {
// 	PlaylistID string    `gorm:"column:playlist_id"`
// 	TrackKey   string    `gorm:"column:track_key"`
// 	TrackURI   string    `gorm:"column:track_uri"`
// 	InsertedAt time.Time `gorm:"column:inserted_at;default:CURRENT_TIMESTAMP"`
// }

type TrackDBModel struct {
	PlaylistID string    `gorm:"column:playlist_id" json:"playlist_id,omitempty"`
	TrackKey   string    `gorm:"column:track_key;primaryKey" json:"track_key,omitempty"`
	TrackURI   string    `gorm:"column:track_uri"  json:"track_uri,omitempty"`
	InsertedAt time.Time `gorm:"column:inserted_at;default:CURRENT_TIMESTAMP" json:"inserted_at,omitempty"`
}
