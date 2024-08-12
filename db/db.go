// Package db provides a database component implementation using GORM.
package db

import (
	"context"
	"fmt"

	"github.com/gopherd/core/component"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/gopherd/components/db/dbapi"
)

// Name is the unique identifier for the database component.
const Name = "github.com/gopherd/components/db"

// Options defines the configuration options for the db component.
type Options struct {
	Driver string // Database driver name
	DSN    string // Data Source Name for database connection
}

// Ensure dbComponent implements dbapi.Component interface.
var _ dbapi.Component = (*dbComponent)(nil)

func init() {
	component.Register(Name, func() component.Component {
		return &dbComponent{}
	})
}

// dbComponent implements the database component.
type dbComponent struct {
	component.BaseComponent[Options]
	db *gorm.DB
}

// Init initializes the database component.
func (c *dbComponent) Init(ctx context.Context) error {
	opts := c.Options()
	db, err := openDB(opts.Driver, opts.DSN)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	c.db = db
	return nil
}

// Uninit closes the database connection.
func (c *dbComponent) Uninit(ctx context.Context) error {
	if c.db == nil {
		return nil
	}
	sqlDB, err := c.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}
	return sqlDB.Close()
}

// Engine returns the GORM database instance.
func (c *dbComponent) Engine() *gorm.DB {
	return c.db
}

// openDB creates a new database connection based on the driver and DSN.
func openDB(driverName, dsn string) (*gorm.DB, error) {
	var dialector gorm.Dialector
	switch driverName {
	case "mysql":
		dialector = mysql.Open(dsn)
	case "postgres":
		dialector = postgres.Open(dsn)
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", driverName)
	}
	return gorm.Open(dialector, &gorm.Config{})
}
