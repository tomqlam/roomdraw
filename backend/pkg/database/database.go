package database

import (
	"database/sql"
	"fmt"
	"net/url"
	"strings"
	"time"

	"roomdraw/backend/pkg/config"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func InitDB() error {
	// replace every space with %20
	encodedPass := url.QueryEscape(config.SQLPass)
	// replace + with %20
	encodedPass = strings.Replace(encodedPass, "+", "%20", -1)

	// Construct the connection string
	connStr := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=%s",
		config.SQLUser, encodedPass, config.SQLIP, config.SQLPort, config.SQLDBName, config.UseSSL)

	fmt.Println("Connection string:", connStr)

	// Open the database connection
	var err error
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
