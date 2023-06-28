package file

import (
	entity "github.com/isd-sgcu/rpkm66-file/internal/entity/file"
	"github.com/isd-sgcu/rpkm66-file/internal/repository/file"
	"gorm.io/gorm"
)

func NewRepository(db *gorm.DB) Repository {
	return file.NewRepositoryImpl(db)
}

type Repository interface {
	FindByOwnerID(string, *entity.File) error
	CreateOrUpdate(*entity.File) error
	Delete(string) error
}
