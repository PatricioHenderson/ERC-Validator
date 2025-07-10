package routes

import (
	"erc-validator/api/internal/middleware"
	"net/http"

	"github.com/gorilla/mux"
)

func InitRoutes() *mux.Router {
	r := mux.NewRouter()

	// public routes
	public := r.NewRoute().Subrouter()
	public.HandleFunc("/admin/users/login", ProxyHandler).Methods(http.MethodPost)
	public.HandleFunc("/admin/users/create", ProxyHandler).Methods(http.MethodPost)

	//private routes
	private := r.NewRoute().Subrouter()
	private.Use(middleware.Auth)
	private.PathPrefix("/").HandlerFunc(ProxyHandler)

	return r
}
