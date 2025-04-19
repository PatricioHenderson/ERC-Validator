package main

import (
	"erc-validator/api/internal/routes"
	"erc-validator/helpers/db/connection"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("could not load env file")
	}

	if _, err := connection.ConnectToDb(); err != nil {
		log.Fatalf("error connecting to DB: %v", err)
	}
	
	r := routes.InitRoutes()

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Printf("API running on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}