package handlers

import (
	"encoding/json"
	"erc-validator/admin/internal/db"
	"erc-validator/admin/internal/models"
	"errors"
	"fmt"
	"net/http"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// func Login {}

func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("invalid JSON: %v", err), http.StatusBadRequest)
		return
	}
	if req.Email == "" || req.Password == "" {
		http.Error(w, "email and password are required", http.StatusBadRequest)
		return
	}

	var existing models.User
	err := db.Conn.Where("email = ?", req.Email).First(&existing).Error
	switch {
	case err == nil:
		http.Error(w, "user already exists", http.StatusBadRequest)
		return
	case errors.Is(err, gorm.ErrRecordNotFound):
	default:
		http.Error(w, fmt.Sprintf("database error: %v", err), http.StatusInternalServerError)
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, fmt.Sprintf("error hashing password: %v", err), http.StatusInternalServerError)
		return
	}

	user := models.User{Email: req.Email, Password: string(hashed)}
	tokenValue, err := GenerateToken()
	if err != nil {
		http.Error(w, fmt.Sprintf("error generating token: %v", err), http.StatusInternalServerError)
		return
	}
	token := models.Token{Value: tokenValue, User: []models.User{user}}

	tx := db.Conn.Begin()
	defer tx.Rollback()

	if err := tx.Create(&user).Error; err != nil {
		http.Error(w, fmt.Sprintf("error creating user: %v", err), http.StatusInternalServerError)
		return
	}
	if err := tx.Create(&token).Error; err != nil {
		http.Error(w, fmt.Sprintf("error creating token: %v", err), http.StatusInternalServerError)
		return
	}
	tx.Commit()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	resp := struct {
		ID    uint   `json:"id"`
		Email string `json:"email"`
		Token string `json:"token"`
	}{user.ID, user.Email, token.Value}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		fmt.Printf("error encoding response: %v\n", err)
	}
}
