package db

import (
	"gorm.io/gorm"
)

const ComponentName = "db"

type Options struct {
	Driver string `json:"driver"`
	DSN    string `json:"dsn"`
}

// Component defines the database component API
type Component interface {
	// Engine returns the database engine
	Engine() *gorm.DB
}
