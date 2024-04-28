package store

import (
	"errors"

	"github.com/NikhilSharmaWe/playree/playree/models"
	"gorm.io/gorm"
)

type UserStore interface {
	CreateTable() error
	Create(fr models.UserDBModel) error
	GetOne(whereQuery string, whereArgs ...interface{}) (*models.UserDBModel, error)
	Update(updateMap map[string]any, whereQuery string, whereArgs ...interface{}) error
	Delete(whereQuery string, whereArgs ...interface{}) error
	IsExists(whereQuery string, whereArgs ...interface{}) (bool, error)
	DB() *gorm.DB
}

type userStore struct {
	db *gorm.DB
}

func NewUserStore(db *gorm.DB) UserStore {

	return &userStore{
		db: db,
	}
}

func (us *userStore) table() string {
	return "users"
}

func (us *userStore) DB() *gorm.DB {
	return us.db
}

func (us *userStore) CreateTable() error {
	return us.db.Table(us.table()).AutoMigrate(models.UserDBModel{})
}

func (us *userStore) Create(fr models.UserDBModel) error {
	return us.db.Table(us.table()).Create(fr).Error
}

func (us *userStore) GetOne(whereQuery string, whereArgs ...interface{}) (*models.UserDBModel, error) {
	var inventory models.UserDBModel
	if err := us.db.Table(us.table()).Where(whereQuery, whereArgs...).First(&inventory).Error; err != nil {
		return nil, err
	}

	return &inventory, nil
}

func (us *userStore) Update(updateMap map[string]any, whereQuery string, whereArgs ...interface{}) error {
	return us.db.Table(us.table()).Where(whereQuery, whereArgs...).Updates(updateMap).Error
}

func (us *userStore) Delete(whereQuery string, whereArgs ...interface{}) error {
	return us.db.Table(us.table()).Where(whereQuery, whereArgs...).Delete(nil).Error
}

func (us *userStore) IsExists(whereQuery string, whereArgs ...interface{}) (bool, error) {

	type Res struct {
		IsExists bool
	}

	var res Res

	if err := us.db.Table(us.table()).Select("1 = 1 AS is_exists").Where(whereQuery, whereArgs...).Find(&res).Error; err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}

		return false, err
	}

	return res.IsExists, nil
}
