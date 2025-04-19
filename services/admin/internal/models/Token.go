package models

import (
	"gorm.io/gorm"
)

type Token struct {
	gorm.Model

	Value string `gorm:"not null;unique_index"`
	User  []User `gorm:"many2many:user_tokens"`
}
