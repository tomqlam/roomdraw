package database

import (
	"database/sql"
	"fmt"
	"net/url"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var DB *sql.DB

func NewDatabase() (*sql.DB, error) {
	godotenv.Load(".env")

	cloudSQLPass := os.Getenv("CLOUD_SQL_PASS")
	cloudSQLIP := os.Getenv("CLOUD_SQL_IP")
	encodedPass := url.QueryEscape(cloudSQLPass)

	// Construct the connection string
	connStr := fmt.Sprintf("postgresql://postgres:%s@%s/test-db", encodedPass, cloudSQLIP)

	// Open the database connection
	db, err := sql.Open("postgres", connStr)
	return db, err
}

func InitDB() error {
	err := godotenv.Load(".env")
	if err != nil {
		return fmt.Errorf("error loading .env file: %v", err)
	}

	cloudSQLPass := os.Getenv("CLOUD_SQL_PASS")
	cloudSQLIP := os.Getenv("CLOUD_SQL_IP")
	encodedPass := url.QueryEscape(cloudSQLPass)

	// Construct the connection string
	connStr := fmt.Sprintf("postgresql://postgres:%s@%s/test-db", encodedPass, cloudSQLIP)

	// Open the database connection
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("error opening database connection: %v", err)
	}

	// Check the connection
	if err = DB.Ping(); err != nil {
		return fmt.Errorf("error pinging database: %v", err)
	}

	return nil
}
