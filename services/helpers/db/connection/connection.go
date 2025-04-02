package connection

import (
	"context"
	"errors"
	"database/sql"
	"fmt"
	"os"
	"time"
	_ "github.com/lib/pq"
)

func ConnectToDb() (*sql.DB, error) {
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

	db, err := sql.Open("postgres", databaseUrl)

	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database (db pointer: %v): %w", db, err)
	}

	return db, nil
}
