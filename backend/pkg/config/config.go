package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

var (
	// Database configuration
	SQLPass   string
	SQLIP     string
	SQLDBName string
	SQLUser   string
	SQLPort   string
	UseSSL    string

	// Authentication
	RequireAuth bool

	// BunnyNet configuration
	BunnyNetReadAPIKey  string
	BunnyNetWriteAPIKey string
	BunnyNetStorageZone string

	// CDN configuration
	CDNURL string

	// Email configuration
	EmailUsername string
	EmailPassword string
)

// Server configuration
const (
	ServerAddress = ":8080"
)

// LoadConfig loads all environment variables once
func LoadConfig() error {
	env := os.Getenv("ENV")
	if env == "" {
		env = "development"
	}

	envFile := fmt.Sprintf(".env.%s", env)
	err := godotenv.Load(envFile)
	if err != nil {
		return fmt.Errorf("error loading %s file: %v", envFile, err)
	}

	// Database configuration
	SQLPass = os.Getenv("SQL_PASS")
	SQLIP = os.Getenv("SQL_IP")
	SQLDBName = os.Getenv("SQL_DB_NAME")
	SQLUser = os.Getenv("SQL_USER")
	SQLPort = os.Getenv("SQL_PORT")
	UseSSL = os.Getenv("USE_SSL")

	// Authentication
	RequireAuth = (os.Getenv("REQUIRE_AUTH") == "True")

	// BunnyNet configuration
	BunnyNetReadAPIKey = os.Getenv("BUNNYNET_READ_API_KEY")
	BunnyNetWriteAPIKey = os.Getenv("BUNNYNET_WRITE_API_KEY")
	BunnyNetStorageZone = os.Getenv("BUNNYNET_STORAGE_ZONE")

	// CDN configuration
	CDNURL = os.Getenv("CDN_URL")

	// Email configuration
	EmailUsername = os.Getenv("EMAIL_USERNAME")
	EmailPassword = os.Getenv("EMAIL_PASSWORD")

	// Log the environment being used
	log.Printf("Loaded configuration from %s", envFile)

	return nil
}
