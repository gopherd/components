package internal

import (
	"context"
	"fmt"

	"github.com/gopherd/core/component"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/gopherd/components/db"
)

func init() {
	component.Register(db.Name, func() component.Component {
		return &DBComponent{}
	})
}

// DBComponent implements the database component.
type DBComponent struct {
	component.BaseComponent[db.Options]
	db *gorm.DB
}

// Init initializes the database component.
func (c *DBComponent) Init(ctx context.Context) error {
	opts := c.Options()
	db, err := openDB(opts.Driver, opts.DSN)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	c.db = db
	return nil
}

// Uninit closes the database connection.
func (c *DBComponent) Uninit(ctx context.Context) error {
	if c.db == nil {
		return nil
	}
	sqlDB, err := c.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}
	return sqlDB.Close()
}

// Ensure dbComponent implements db.Component interface.
var _ db.Component = (*DBComponent)(nil)

// Engine returns the GORM database instance.
func (c *DBComponent) Engine() *gorm.DB {
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
