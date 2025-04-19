package handlers

import "net/http"

func LogInHandler(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("Login succeeded"))
}