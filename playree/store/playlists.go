package store

import (
	"errors"

	"github.com/NikhilSharmaWe/playree/playree/models"
	"gorm.io/gorm"
)

type PlaylistStore interface {
	CreateTable() error
	Create(fr models.PlaylistsDBModel) error
	GetOne(whereQuery string, whereArgs ...interface{}) (*models.PlaylistsDBModel, error)
	GetManyWithFields(fields []string, whereQuery string, whereArgs ...interface{}) ([]models.PlaylistsDBModel, error)
	Update(updateMap map[string]any, whereQuery string, whereArgs ...interface{}) error
	Delete(whereQuery string, whereArgs ...interface{}) error
	IsExists(whereQuery string, whereArgs ...interface{}) (bool, error)
	DB() *gorm.DB
}

type playlistStore struct {
	db *gorm.DB
}

func NewPlaylistStore(db *gorm.DB) PlaylistStore {
	return &playlistStore{
		db: db,
	}
}

func (ps *playlistStore) table() string {
	return "playlists"
}

func (ps *playlistStore) DB() *gorm.DB {
	return ps.db
}

func (ps *playlistStore) CreateTable() error {
	return ps.db.Table(ps.table()).AutoMigrate(models.PlaylistsDBModel{})
}

func (ps *playlistStore) Create(fr models.PlaylistsDBModel) error {
	return ps.db.Table(ps.table()).Create(fr).Error
}

func (ps *playlistStore) GetOne(whereQuery string, whereArgs ...interface{}) (*models.PlaylistsDBModel, error) {
	var inventory models.PlaylistsDBModel
	if err := ps.db.Table(ps.table()).Where(whereQuery, whereArgs...).First(&inventory).Error; err != nil {
		return nil, err
	}

	return &inventory, nil
}

func (ps *playlistStore) GetManyWithFields(fields []string, whereQuery string, whereArgs ...interface{}) ([]models.PlaylistsDBModel, error) {
	var playlists []models.PlaylistsDBModel

	if err := ps.db.Table(ps.table()).Select(fields).Where(whereQuery, whereArgs...).Find(&playlists).Error; err != nil {
		return nil, err
	}

	return playlists, nil
}

func (ps *playlistStore) Update(updateMap map[string]any, whereQuery string, whereArgs ...interface{}) error {
	return ps.db.Table(ps.table()).Where(whereQuery, whereArgs...).Updates(updateMap).Error
}

func (ps *playlistStore) Delete(whereQuery string, whereArgs ...interface{}) error {
	return ps.db.Table(ps.table()).Where(whereQuery, whereArgs...).Delete(nil).Error
}

func (ps *playlistStore) IsExists(whereQuery string, whereArgs ...interface{}) (bool, error) {

	type Res struct {
		IsExists bool
	}

	var res Res

	if err := ps.db.Table(ps.table()).Select("1 = 1 AS is_exists").Where(whereQuery, whereArgs...).Find(&res).Error; err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}

		return false, err
	}

	return res.IsExists, nil
}
