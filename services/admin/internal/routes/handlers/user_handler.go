package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"erc-validator/admin/internal/db"
	"erc-validator/admin/internal/models"
	"erc-validator/api/helpers/auth"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

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
	result := db.Conn.Where("email = ?", req.Email).First(&existing)
	if result.RowsAffected > 0 {
		http.Error(w, "user already exists", http.StatusBadRequest)
		return
	}
	if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		http.Error(w, fmt.Sprintf("database error: %v", result.Error), http.StatusInternalServerError)
		return
	}

	// Usuario no existe → continuar normalmente

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, fmt.Sprintf("error hashing password: %v", err), http.StatusInternalServerError)
		return
	}

	user := models.User{Email: req.Email, Password: string(hashed)}
	// Generate a JWT token for the new user
	tokenValue, err := auth.CreateToken(req.Email)
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

func GetMeUserHandler(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(w, "Missing or invalid Authorization header", http.StatusUnauthorized)
		return
	}
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	claims, err := auth.VerifyToken(tokenString)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	email, ok := claims["email"].(string)
	if !ok {
		http.Error(w, "Invalid token payload", http.StatusUnauthorized)
		return
	}

	var user models.User
	if err := db.Conn.Where("email = ?", email).First(&user).Error; err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func LogInUserHandler(w http.ResponseWriter, r *http.Request) {
	var credentials struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}
	if !models.IsPasswordValid(credentials.Email, credentials.Password) {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}
	tokenString, err := auth.CreateToken(credentials.Email)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
}

func LogOutUserHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
	})
	w.WriteHeader(http.StatusOK)
	http.Redirect(w, r, "/home", http.StatusSeeOther)
}

func ChangePasswordHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value("user_id").(string)

	var user models.User
	if err := db.Conn.First(&user, userID).Error; err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword)); err != nil {
		http.Error(w, "Invalid current password", http.StatusUnauthorized)
		return
	}

	hashedNew, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error hashing new password", http.StatusInternalServerError)
		return
	}

	if err := db.Conn.
		Model(&models.User{}).
		Where("id = ?", userID).
		Update("password", string(hashedNew)).Error; err != nil {
		http.Error(w, "Could not update password", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Password changed successfully",
	})
}
