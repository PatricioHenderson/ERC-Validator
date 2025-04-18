package main

import (
	"erc-validator/helpers/db/connection"
	"log"
	// "net/http"
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
	// defer db.Close()
	//TODO: Use these only to pass pre-commit. Fix this in future. Or look after a db.close()
	db.Begin()

	// r := routes.InitRoutes()

	// adminDB.Conn = db

	port := os.Getenv("PORT")
	if port == "" {
		port = "3004"
	}

	log.Printf("ADMIN running on port %s", port)
	// log.Fatal(http.ListenAndServe(":"+port, r))
}
