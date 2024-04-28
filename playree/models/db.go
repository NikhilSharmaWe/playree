package models

type UserDBModel struct {
	UserID   string `gorm:"column:user_id;primaryKey"`
	Username string `gorm:"column:username"`
}

type PlaylistsDBModel struct {
	UserID     string `gorm:"column:user_id"`
	PlaylistID string `gorm:"column:playlist_id"`
}
