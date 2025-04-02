package main

import (
	"log"
	"net/http"
	"os"
	"erc-validator/helpers/db/connection"
)

func main() {
	db, err := connection.ConnectToDb()
	if err != nil {
		log.Fatalf("error connecting to DB: %v", err)
	}
	defer db.Close()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte("Hello world from web3")); err != nil {
			log.Printf("error writing response: %v", err)
		}
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("API running on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
