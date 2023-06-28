package database

import (
	"fmt"
	"strconv"

	"github.com/isd-sgcu/rpkm66-file/cfgldr"
	"github.com/isd-sgcu/rpkm66-file/internal/model/file"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func InitDatabase(conf *cfgldr.Database) (db *gorm.DB, err error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True", conf.User, conf.Password, conf.Host, strconv.Itoa(conf.Port), conf.Name)

	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(file.File{})
	if err != nil {
		return nil, err
	}

	return
}
