package main

import (
	adminDB "erc-validator/admin/internal/db"
	"erc-validator/admin/internal/models"
	"erc-validator/admin/internal/routes"
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

	db, err := connection.ConnectToDb()
	if err != nil {
		log.Fatalf("error connecting to DB: %v", err)
	}

	err = db.AutoMigrate(&models.User{}, &models.Token{})
	if err != nil {
		log.Fatalf("error executing auto migration: %v", err)
	}

	r := routes.InitRoutes()

	adminDB.Conn = db

	port := os.Getenv("PORT")
	if port == "" {
		port = "3004"
	}

	log.Printf("ADMIN running on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
