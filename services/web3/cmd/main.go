package main

import (
	"erc-validator/helpers/db/connection"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("could not load env file")
	}

	db, err := connection.ConnectToDb()
	if err != nil {
		log.Fatalf("error connecting to DB: %v", err)
	}
	// defer db.Close()
	//TODO: Use these only to pass pre-commit. Fix this in future. Or look after a db.close()
	db.Begin()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte("Hello world from api")); err != nil {
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
