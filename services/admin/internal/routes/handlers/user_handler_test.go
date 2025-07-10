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
		t.Fatalf("couldn't open sqlite: %v", err)
	}
	if err := gormDB.AutoMigrate(&models.User{}, &models.Token{}); err != nil {
		t.Fatalf("migration failed: %v", err)
	}
	db.Conn = gormDB

	r := mux.NewRouter()
	r.HandleFunc("/users/create", CreateUserHandler).Methods(http.MethodPost)
	r.HandleFunc("/users/login", LogInUserHandler).Methods(http.MethodPost)
	return r
}

func TestCreateUserHandler_Success(t *testing.T) {
	router := setupRouter(t)

	payload := map[string]string{
		"email":    "foo@example.com",
		"password": "secret123",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/users/create", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		ID    uint   `json:"id"`
		Email string `json:"email"`
		Token string `json:"token"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("JSON not valid: %v", err)
	}
	if resp.Email != "foo@example.com" {
		t.Errorf("expected email 'foo@example.com', got %q", resp.Email)
	}
}

func TestCreateUserHandler_Duplicate(t *testing.T) {
	router := setupRouter(t)

	payload := map[string]string{"email": "dup@test.com", "password": "pw"}
	body, _ := json.Marshal(payload)

	req1 := httptest.NewRequest(http.MethodPost, "/users/create", bytes.NewReader(body))
	req1.Header.Set("Content-Type", "application/json")
	rec1 := httptest.NewRecorder()
	router.ServeHTTP(rec1, req1)
	if rec1.Code != http.StatusCreated {
		t.Fatalf("setup: user not created: %d", rec1.Code)
	}

	req2 := httptest.NewRequest(http.MethodPost, "/users/create", bytes.NewReader(body))
	req2.Header.Set("Content-Type", "application/json")
	rec2 := httptest.NewRecorder()
	router.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusBadRequest {
		t.Errorf("esperaba 400 Bad Request por duplicado, got %d", rec2.Code)
	}
}

func TestCreateUserHandler_InvalidJSON(t *testing.T) {
	router := setupRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/users/create", bytes.NewReader([]byte(`{email:}`)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("400 expected, got %d", rec.Code)
	}
}

func TestCreateUserHandler_MissingFields(t *testing.T) {
	router := setupRouter(t)

	body, _ := json.Marshal(map[string]string{})
	req := httptest.NewRequest(http.MethodPost, "/users/create", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("400 expected, got %d", rec.Code)
	}
}

func TestLoginUserHandler_Success(t *testing.T) {
	router := setupRouter(t)

	createPayload := map[string]string{"email": "foo@example.com", "password": "secret123"}
	createBody, _ := json.Marshal(createPayload)
	createReq := httptest.NewRequest(http.MethodPost, "/users/create", bytes.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createRec := httptest.NewRecorder()
	router.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("setup: user not created: %d", createRec.Code)
	}

	loginPayload := map[string]string{"email": "foo@example.com", "password": "secret123"}
	loginBody, _ := json.Marshal(loginPayload)
	loginReq := httptest.NewRequest(http.MethodPost, "/users/login", bytes.NewReader(loginBody))
	loginReq.Header.Set("Content-Type", "application/json")
	loginRec := httptest.NewRecorder()
	router.ServeHTTP(loginRec, loginReq)

	if loginRec.Code != http.StatusSeeOther {
		t.Fatalf("expected 303 See Other, got %d", loginRec.Code)
	}
	if loc := loginRec.Header().Get("Location"); loc != "/home/user" {
		t.Fatalf("expected redirect to /home/user, got %q", loc)
	}
}
