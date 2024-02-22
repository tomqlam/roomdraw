package database

import (
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var DB *sql.DB

func InitDB() error {
	err := godotenv.Load(".env")
	if err != nil {
		return fmt.Errorf("error loading .env file: %v", err)
	}

	cloudSQLPass := os.Getenv("CLOUD_SQL_PASS")
	cloudSQLIP := os.Getenv("CLOUD_SQL_IP")
	cloudSQLDBName := os.Getenv("CLOUD_SQL_DB_NAME")
	cloudSQLUser := os.Getenv("CLOUD_SQL_USER")
	encodedPass := url.PathEscape(strings.TrimSpace(cloudSQLPass))

	// Construct the connection string
	connStr := fmt.Sprintf("postgresql://%s:%s@%s/%s", cloudSQLUser, encodedPass, cloudSQLIP, cloudSQLDBName)

	log.Println("connStr", connStr)

	// Open the database connection
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("error opening database connection: %v", err)
	}

	// Configure connection pool
	DB.SetMaxOpenConns(25)                 // Example: 25 open connections
	DB.SetMaxIdleConns(10)                 // Example: 10 idle connections
	DB.SetConnMaxLifetime(5 * time.Minute) // Example: 5 minutes

	// Check the connection
	if err = DB.Ping(); err != nil {
		return fmt.Errorf("error pinging database: %v", err)
	}

	return nil
}
