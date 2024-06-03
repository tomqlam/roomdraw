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

	SQLPass := os.Getenv("SQL_PASS")
	SQLIP := os.Getenv("SQL_IP")
	SQLDBName := os.Getenv("SQL_DB_NAME")
	SQLUser := os.Getenv("SQL_USER")
	useSSL := os.Getenv("USE_SSL")
	// replace every space with %20
	encodedPass := url.QueryEscape(SQLPass)
	// replace + with %20
	encodedPass = strings.Replace(encodedPass, "+", "%20", -1)

	// Construct the connection string
	connStr := fmt.Sprintf("postgresql://%s:%s@%s/%s?sslmode=%s", SQLUser, encodedPass, SQLIP, SQLDBName, useSSL)

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
