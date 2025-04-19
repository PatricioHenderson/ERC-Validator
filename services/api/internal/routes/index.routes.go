package routes

import (
	// "erc-validator/api/internal/middleware"
	"erc-validator/api/internal/routes/handlers"
	"net/http"
)

func InitRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	// public routes
	mux.HandleFunc("/login", handlers.LogInHandler)

	//private routes
	// mux.Handle("/admin", middleware.autMiddleware(http.HandlerFunc(handlers.AdminHandler)))
	return mux
}