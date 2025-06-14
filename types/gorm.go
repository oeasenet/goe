package types

import (
	"gorm.io/gorm"
)

type DBModel struct {
	ID        string         `gorm:"primarykey"`
	CreatedAt int64          `gorm:"autoCreateTime:milli"`
	UpdatedAt int64          `gorm:"autoUpdateTime:milli"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
