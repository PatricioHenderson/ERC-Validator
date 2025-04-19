package models

import (
	"fmt"
	"net/mail"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model

	Email	string	`gorm:"not null;unique_index"`
	Password	string `gorm:"not null"`
	Token 	[]Token	`gorm:"many2many:user_tokens"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if _, err:= mail.ParseAddress(u.Email); err != nil {
		return fmt.Errorf("invalid email: %w", err)
	}
	return nil
}