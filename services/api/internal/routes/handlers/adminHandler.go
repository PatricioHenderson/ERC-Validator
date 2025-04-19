package handlers

import "net/http"

func AdminHandler(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("Admin"))
}