package component

import "gorm.io/gorm"

func (com *dbComponent) Engine() *gorm.DB {
	return com.engine
}
