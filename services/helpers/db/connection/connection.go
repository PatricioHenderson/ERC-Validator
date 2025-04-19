package connection

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func ConnectToDb() (*gorm.DB, error) {
	if os.Getenv("NODE_ENV") != "production" {
		if err := godotenv.Load(); err != nil {
			log.Println("No .env file found, using system environment variables")
		}
	}

	databaseUrl := os.Getenv("DATABASE_URL")
	if databaseUrl == "" {
		errMsg := "missing required environment variable: DATABASE_URL"
		fmt.Println(errMsg)
		return nil, errors.New(errMsg)
	}

	if os.Getenv("NODE_ENV") == "production" {
		path := os.Getenv("POSTGRES_CA_PATH")
		if path == "" {
			errMsg := "missing environment variable for TLS: POSTGRES_CA_PATH"
			fmt.Println(errMsg)
			return nil, errors.New(errMsg)
		}
		databaseUrl = fmt.Sprintf("%s sslmode=verify-full sslrootcert=%s", databaseUrl, path)
	}

	db, err := gorm.Open(postgres.Open(databaseUrl), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	return db, nil
}
