package dbapi

import (
	"gorm.io/gorm"
)

// Component defines the database component API
type Component interface {
	// Engine returns the database engine
	Engine() *gorm.DB
}
