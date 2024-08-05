package db

import (
	"context"
	"fmt"

	dbapi "github.com/gopherd/components/db/api"
	"github.com/gopherd/core/component"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Name represents the name of the component.
const Name = "github.com/gopherd/components/db"

// Options represents the options of the component.
type Options struct {
	Driver string
	DSN    string
}

var _ dbapi.Component = (*dbComponent)(nil)

func init() {
	component.Register(Name, func() component.Component {
		return &dbComponent{}
	})
}

type dbComponent struct {
	component.BaseComponent[Options]
	engine *gorm.DB
}

func (com *dbComponent) Init(ctx context.Context) error {
	options := com.Options()
	if db, err := open(options.Driver, options.DSN); err != nil {
		return err
	} else {
		com.engine = db
	}
	return nil
}

func (com *dbComponent) Uninit(ctx context.Context) error {
	if db, err := com.engine.DB(); err == nil {
		return db.Close()
	}
	return nil
}

func (com *dbComponent) Engine() *gorm.DB {
	return com.engine
}

func open(driverName string, dsn string) (*gorm.DB, error) {
	var dialector gorm.Dialector
	switch driverName {
	case "mysql":
		dialector = mysql.Open(dsn)
	case "postgres":
		dialector = postgres.Open(dsn)
	default:
	}
	if dialector == nil {
		return nil, fmt.Errorf("unsupported database driver: %s", driverName)
	}
	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
}
