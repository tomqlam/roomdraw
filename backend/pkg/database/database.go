package database

import (
	"database/sql"
	"fmt"
	"net/url"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

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
