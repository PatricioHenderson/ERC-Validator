package models

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setUpTestDB() *gorm.DB {
	db, _ := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	db.AutoMigrate(&User{}, &Token{})
	return db
}

func TestUserEmailValidator(t *testing.T) {
	db := setUpTestDB()
	bad := User{Email: "no‑email", Password: "secret"}
	if err := db.Create(&bad).Error; err == nil {
		t.Error("expected error while creating user but was, nil")
	}

	good := User{Email: "foo@example.com", Password: "secret"}
	if err := db.Create(&good).Error; err != nil {
		t.Errorf("No error expected to create the user: %v", err)
	}

}
