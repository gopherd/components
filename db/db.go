package db

import (
	"gorm.io/gorm"
)

// Name represents the name of the component.
const Name = "github.com/gopherd/components/db"

// Options represents the options of the component.
type Options struct {
	Driver string `json:"driver"`
	DSN    string `json:"dsn"`
}

// Component defines the database component API
type Component interface {
	// Engine returns the database engine
	Engine() *gorm.DB
}
