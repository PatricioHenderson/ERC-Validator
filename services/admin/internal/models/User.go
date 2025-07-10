package models

import (
	"erc-validator/admin/internal/db"
	"fmt"
	"net/mail"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model

	Email    string  `gorm:"not null;unique_index"`
	Password string  `gorm:"not null"`
	Token    []Token `gorm:"many2many:user_tokens"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if _, err := mail.ParseAddress(u.Email); err != nil {
		return fmt.Errorf("invalid email: %w", err)
	}
	return nil
}

func IsPasswordValid(email string, password string) bool {
	var user User
	result := db.Conn.Where("email = ?", email).First(&user)
	if result.Error != nil {
		return false
	}
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	return err == nil
}
