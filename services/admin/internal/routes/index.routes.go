package routes

import (
	"net/http"

	"erc-validator/admin/internal/routes/handlers"

	"github.com/gorilla/mux"
)

func InitRoutes() *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/users/login", handlers.LogInUserHandler).Methods(http.MethodPost)
	r.HandleFunc("/users/create", handlers.CreateUserHandler).Methods(http.MethodPost)

	//Private routes
	r.HandleFunc("/users/me", handlers.GetMeUserHandler).Methods(http.MethodGet)
	r.HandleFunc("/users/logout", handlers.LogOutUserHandler).Methods(http.MethodPost)

	return r
}
