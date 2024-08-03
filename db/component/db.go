package component

import (
	"context"

	"github.com/gopherd/core/component"
	"gorm.io/gorm"

	"github.com/gopherd/components/db"
)

var _ db.Component = (*dbComponent)(nil)

func init() {
	component.Register(db.ComponentName, func() component.Component {
		return &dbComponent{}
	})
}

type dbComponent struct {
	component.BaseComponent[db.Options]
	engine *gorm.DB
}

func (com *dbComponent) Init(ctx context.Context, entity component.Entity) error {
	options := com.Options()
	if db, err := open(options.Driver, options.DSN); err != nil {
		return err
	} else {
		com.engine = db
	}
	return nil
}

func (com *dbComponent) Shutdown(ctx context.Context) error {
	if db, err := com.engine.DB(); err == nil {
		return db.Close()
	}
	return nil
}
