package file

import (
	"github.com/isd-sgcu/rpkm66-file/internal/entity/file"
	"gorm.io/gorm"
)

type repositoryImpl struct {
	db *gorm.DB
}

func (r *repositoryImpl) FindByOwnerID(id string, result *file.File) error {
	return r.db.First(&result, "owner_id = ?", id).Error
}

func (r *repositoryImpl) CreateOrUpdate(result *file.File) error {
	if r.db.Where("owner_id = ?", result.OwnerID).Updates(&result).RowsAffected == 0 {
		return r.db.Create(&result).Error
	}
	return nil
}

func (r *repositoryImpl) Delete(id string) error {
	return r.db.First(id).Delete(&file.File{}).Error
}

func NewRepositoryImpl(db *gorm.DB) *repositoryImpl {
	return &repositoryImpl{
		db: db,
	}
}
