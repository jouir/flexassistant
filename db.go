package main

import (
	"time"

	log "github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// NewDatabase creates a SQLite database object from a file name
func NewDatabase(filename string) (*gorm.DB, error) {
	return gorm.Open(sqlite.Open(filename), &gorm.Config{})
}

// CreateDatabaseObjects creates database relations
func CreateDatabaseObjects(db *gorm.DB) error {
	if err := db.AutoMigrate(&Miner{}); err != nil {
		return err
	}
	if err := db.AutoMigrate(&Worker{}); err != nil {
		return err
	}
	if err := db.AutoMigrate(&Pool{}); err != nil {
		return err
	}
	return nil
}

// EnsureDatabaseRetention removes stale objects from database
func EnsureDatabaseRetention(db *gorm.DB) error {
	log.Debugf("Deleting inactive workers")
	lastWeek := time.Now().AddDate(0, 0, -7)
	var worker *Worker
	trx := db.Unscoped().Where("last_seen < ?", lastWeek).Delete(&worker)
	if trx.Error != nil {
		return trx.Error
	}
	return nil
}
