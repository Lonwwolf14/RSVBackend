package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/pressly/goose/v3"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

func main() {
	// Fetch DATABASE_URL from environment
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Println("Error: DATABASE_URL environment variable is not set or empty")
		log.Println("Current environment variables:", os.Environ()) // Debug output
		log.Fatal("DATABASE_URL is required to connect to the database")
	}

	// Open database connection
	db, err := sql.Open("libsql", dbURL)
	if err != nil {
		log.Fatalf("Failed to open database connection: %v", err)
	}
	defer db.Close()

	// Verify the connection is alive
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Enable foreign keys (specific to SQLite/LibSQL)
	if _, err := db.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		log.Fatalf("Failed to enable foreign keys: %v", err)
	}

	// Set dialect for goose (SQLite3 works for LibSQL)
	goose.SetDialect("sqlite3")

	// Apply migrations from sql/schema directory
	if err := goose.Up(db, "sql/schema"); err != nil {
		log.Fatalf("Failed to apply migrations: %v", err)
	}

	log.Println("Migrations applied successfully!")
}
