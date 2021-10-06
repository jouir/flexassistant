package main

import (
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
	if err := db.AutoMigrate(&Pool{}); err != nil {
		return err
	}
	return nil
}
