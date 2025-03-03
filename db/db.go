package db

import (
	"github.com/takedown-observer/backend/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type DB struct {
	*gorm.DB
}

func New(filepath string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(filepath), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(&models.Account{})
	if err != nil {
		return nil, err
	}

	return db, nil
}
