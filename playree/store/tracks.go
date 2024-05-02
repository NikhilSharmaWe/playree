package store

import (
	"errors"

	"github.com/NikhilSharmaWe/playree/playree/models"
	"gorm.io/gorm"
)

type TrackStore interface {
	CreateTable() error
	Create(fr models.TrackDBModel) error
	CreateInBatches(tracks []models.TrackDBModel) error
	GetOne(whereQuery string, whereArgs ...interface{}) (*models.TrackDBModel, error)
	GetMany(fields []string, whereQuery string, whereArgs ...interface{}) ([]models.TrackDBModel, error)
	Update(updateMap map[string]any, whereQuery string, whereArgs ...interface{}) error
	Delete(whereQuery string, whereArgs ...interface{}) error
	IsExists(whereQuery string, whereArgs ...interface{}) (bool, error)
	DB() *gorm.DB
}

type trackStore struct {
	db *gorm.DB
}

func NewTrackStore(db *gorm.DB) TrackStore {
	return &trackStore{
		db: db,
	}
}

func (ps *trackStore) table() string {
	return "tracks"
}

func (ps *trackStore) DB() *gorm.DB {
	return ps.db
}

func (ps *trackStore) CreateTable() error {
	return ps.db.Table(ps.table()).AutoMigrate(models.TrackDBModel{})
}

func (ps *trackStore) Create(fr models.TrackDBModel) error {
	return ps.db.Table(ps.table()).Create(fr).Error
}

func (ps *trackStore) CreateInBatches(tracks []models.TrackDBModel) error {
	return ps.db.Table(ps.table()).CreateInBatches(tracks, len(tracks)).Error
}

func (ps *trackStore) GetOne(whereQuery string, whereArgs ...interface{}) (*models.TrackDBModel, error) {
	var track models.TrackDBModel
	if err := ps.db.Table(ps.table()).Where(whereQuery, whereArgs...).First(&track).Error; err != nil {
		return nil, err
	}

	return &track, nil
}

func (ps *trackStore) GetMany(fields []string, whereQuery string, whereArgs ...interface{}) ([]models.TrackDBModel, error) {
	var tracks []models.TrackDBModel

	if err := ps.db.Table(ps.table()).Select(fields).Where(whereQuery, whereArgs...).Find(&tracks).Error; err != nil {
		return nil, err
	}

	return tracks, nil
}

func (ps *trackStore) Update(updateMap map[string]any, whereQuery string, whereArgs ...interface{}) error {
	return ps.db.Table(ps.table()).Where(whereQuery, whereArgs...).Updates(updateMap).Error
}

func (ps *trackStore) Delete(whereQuery string, whereArgs ...interface{}) error {
	return ps.db.Table(ps.table()).Where(whereQuery, whereArgs...).Delete(nil).Error
}

func (ps *trackStore) IsExists(whereQuery string, whereArgs ...interface{}) (bool, error) {

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
