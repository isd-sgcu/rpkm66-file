package database

import (
	"fmt"
	"strconv"

	"github.com/isd-sgcu/rpkm66-file/cfgldr"
	"github.com/isd-sgcu/rpkm66-file/internal/entity/file"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitDatabase(conf *cfgldr.Database) (db *gorm.DB, err error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s", conf.Host, conf.User, conf.Password, conf.Name, strconv.Itoa(conf.Port))

	db, err = gorm.Open(postgres.New(postgres.Config{
		DSN: dsn,
	}), &gorm.Config{})

	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(file.File{})
	if err != nil {
		return nil, err
	}

	return
}
