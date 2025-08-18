package domain

import "gorm.io/gorm"

type Anger struct {
	gorm.Model
	User string `gorm:"column:user;not null" json:"user"`
}

func (Anger) TableName() string {
	return "tilted"
}
