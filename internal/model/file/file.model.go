package file

import (
	"github.com/isd-sgcu/rpkm66-file/internal/model"
)

type File struct {
	model.Base
	Filename string `json:"filename" gorm:"index"`
	OwnerID  string `json:"owner_id" gorm:"index:,unique"`
	Tag      int    `json:"tag"`
}
