package db

import (
	"gorm.io/gorm"
)

const ComponentName = "github.com/gopherd/components/db"

type Options struct {
	Driver string `json:"driver"`
	DSN    string `json:"dsn"`
}

// Component defines the database component API
type Component interface {
	// Engine returns the database engine
	Engine() *gorm.DB
}
