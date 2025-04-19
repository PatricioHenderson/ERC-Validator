package routes

import (
	// "erc-validator/admin/internal/routes/handlers"
	// "erc-validator/api/internal/middleware"
	"github.com/gorilla/mux"
	"erc-validator/admin/internal/routes/handlers"
	"net/http"
)

func InitRoutes() *mux.Router {
	r := mux.NewRouter()

	// public routes
	// mux.HandleFunc("/login", handlers.LogInHandler)
	
	r.HandleFunc("/user/create", handlers.CreateUserHandler).Methods(http.MethodPost)
	//private routes
	// mux.Handle("/admin", middleware.autMiddleware(http.HandlerFunc(handlers.AdminHandler)))

	return r
}