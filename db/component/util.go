package component

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

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
