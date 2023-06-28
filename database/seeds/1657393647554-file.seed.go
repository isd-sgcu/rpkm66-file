package seed

import (
	"github.com/bxcodec/faker/v4"
	"github.com/isd-sgcu/rpkm66-file/internal/entity/file"
)

func (s Seed) UserSeed1655751437484() error {
	for i := 0; i < 10; i++ {
		usr := file.File{
			Filename: faker.Word(),
			OwnerID:  faker.UUIDDigit(),
			Tag:      1,
		}
		err := s.db.Create(&usr).Error

		if err != nil {
			return err
		}
	}
	return nil
}
