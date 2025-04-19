// services/admin/internal/routes/handlers/user_handler_test.go
package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"erc-validator/admin/internal/db"
	"erc-validator/admin/internal/models"

	"github.com/gorilla/mux"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupRouter(t *testing.T) *mux.Router {
	t.Helper()

	gormDB, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("no pudo abrir SQLite en memoria: %v", err)
	}

	if err := gormDB.AutoMigrate(&models.User{}, &models.Token{}); err != nil {
		t.Fatalf("migración falló: %v", err)
	}

	db.Conn = gormDB

	r := mux.NewRouter()
	r.HandleFunc("/user/create", CreateUserHandler).Methods(http.MethodPost)
	return r
}

func TestCreateUserHandler_Success(t *testing.T) {
	router := setupRouter(t)

	payload := map[string]string{
		"email":    "foo@example.com",
		"password": "secret123",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/user/create", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("esperaba 201 Created, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		ID    uint   `json:"id"`
		Email string `json:"email"`
		Token string `json:"token"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("respuesta no es JSON válido: %v", err)
	}
	if resp.Email != "foo@example.com" {
		t.Errorf("esperaba email 'foo@example.com', got %q", resp.Email)
	}
	if resp.Token == "" {
		t.Error("esperaba un token no vacío")
	}
}

func TestCreateUserHandler_Duplicate(t *testing.T) {
	router := setupRouter(t)

	payload := map[string]string{"email": "dup@test.com", "password": "pw"}
	body, _ := json.Marshal(payload)

	req1 := httptest.NewRequest(http.MethodPost, "/user/create", bytes.NewReader(body))
	req1.Header.Set("Content-Type", "application/json")
	rec1 := httptest.NewRecorder()
	router.ServeHTTP(rec1, req1)
	if rec1.Code != http.StatusCreated {
		t.Fatalf("setup: no pudo crear user: %d", rec1.Code)
	}

	req2 := httptest.NewRequest(http.MethodPost, "/user/create", bytes.NewReader(body))
	req2.Header.Set("Content-Type", "application/json")
	rec2 := httptest.NewRecorder()
	router.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusBadRequest {
		t.Errorf("esperaba 400 Bad Request por duplicado, got %d", rec2.Code)
	}
}

func TestCreateUserHandler_InvalidJSON(t *testing.T) {
	router := setupRouter(t)
	req := httptest.NewRequest(http.MethodPost, "/user/create", bytes.NewReader([]byte(`{email:}`)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("esperaba 400 por JSON inválido, got %d", rec.Code)
	}
}

func TestCreateUserHandler_MissingFields(t *testing.T) {
	router := setupRouter(t)
	body, _ := json.Marshal(map[string]string{})
	req := httptest.NewRequest(http.MethodPost, "/user/create", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("esperaba 400 por campos faltantes, got %d", rec.Code)
	}
}
