package database

import (
	"database/sql"
	"fmt"
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
	useSSL := os.Getenv("USE_SSL")
	// replace every space with %20
	encodedPass := url.QueryEscape(cloudSQLPass)
	// replace + with %20
	encodedPass = strings.Replace(encodedPass, "+", "%20", -1)

	// Construct the connection string
	connStr := fmt.Sprintf("postgresql://%s:%s@%s/%s?sslmode=%s", cloudSQLUser, encodedPass, cloudSQLIP, cloudSQLDBName, useSSL)

	// log.Println("connStr", connStr)

	// Open the database connection
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("error opening database connection: %v", err)
	}

	// Configure connection pool
	DB.SetMaxOpenConns(100)
	DB.SetMaxIdleConns(100)
	DB.SetConnMaxLifetime(5 * time.Minute)

	// Check the connection
	if err = DB.Ping(); err != nil {
		return fmt.Errorf("error pinging database: %v", err)
	}

	return nil
}
