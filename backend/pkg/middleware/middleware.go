package middleware

import (
	"crypto/rsa"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"log"
	"math/big"
	"net/http"
	"roomdraw/backend/pkg/database"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

// Define a type for the request type (read or write)
type RequestType int

type GooglePublicKeysResponse struct {
	Keys []struct {
		Alg string `json:"alg"`
		Kty string `json:"kty"`
		Kid string `json:"kid"`
		N   string `json:"n"`
		Use string `json:"use"`
		E   string `json:"e"`
	} `json:"keys"`
}

const (
	Read RequestType = iota
	Write
)

const googleCertsURL = "https://www.googleapis.com/oauth2/v3/certs"

var (
	keysCache     GooglePublicKeysResponse
	cacheMutex    = &sync.RWMutex{}
	keysCacheTime time.Time
)

func BetaTesterMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		email, exists := c.Get("email")
		if !exists {
			c.Next()
			return
		}
		log.Println("Email:", email)
		// List of authorized beta testers
		betaTesters := []string{
			"tlam@g.hmc.edu",
			"smao@g.hmc.edu",
			"elli@g.hmc.edu",
			"kaguo@g.hmc.edu",
			"aazhang@g.hmc.edu",
			"nnickolov@g.hmc.edu",
			"amcintoshlombardo@g.hmc.edu",
			"wkirkland@g.hmc.edu",
			"admedina@g.hmc.edu",
			"huahuang@g.hmc.edu",
			"jgonzalezsalgado@g.hmc.edu",
			"sophiewang@g.hmc.edu",
			"smichaelson@g.hmc.edu",
			"jkeodara@g.hmc.edu",
			"niluo@g.hmc.edu",
			"stamayo@g.hmc.edu",
			"rpenidelema@g.hmc.edu",
			"alrosenberg@g.hmc.edu",
			"meldeng@g.hmc.edu",
			"edgrodriguez@g.hmc.edu",
			"aniksharma@g.hmc.edu",
			"ashetty@g.hmc.edu",
			"simyang@g.hmc.edu",
			"igodoy@g.hmc.edu",
			"breis@g.hmc.edu",
			"brmendoza@g.hmc.edu",
			"lmansfield@g.hmc.edu",
			"grwilliams@g.hmc.edu",
			"lcromwell@g.hmc.edu",
			"aenayati@g.hmc.edu",
			"jeshuang@g.hmc.edu",
			"perodriguez@g.hmc.edu",
			"calmond@g.hmc.edu",
			"dimehta@g.hmc.edu",
			"mgaribaybrandt@g.hmc.edu",
			"amcdaniel@g.hmc.edu",
			"nisorena@g.hmc.edu",
			"saraliu@g.hmc.edu",
		}

		// Check if the user's email is in the list of beta testers
		authorized := false
		for _, tester := range betaTesters {
			if email == tester {
				authorized = true
				log.Println("Authorized:", email)
				break
			}
		}

		if !authorized {
			log.Println("Not authorized:", email)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized access"})
			return
		}

		c.Next()
	}
}

// FetchGooglePublicKeys fetches and caches Google's public keys for JWT validation.
func FetchGooglePublicKeys() error {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	// Check if keys are still fresh, let's assume keys are refreshed every 24 hours for this example
	if time.Since(keysCacheTime) < 24*time.Hour && len(keysCache.Keys) > 0 {
		return nil // Keys are still fresh
	}

	resp, err := http.Get(googleCertsURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var certs GooglePublicKeysResponse
	if err := json.NewDecoder(resp.Body).Decode(&certs); err != nil {
		log.Println("Response:", resp)
		return err
	}

	keysCache = certs
	keysCacheTime = time.Now()
	return nil
}

// getKeyFunc is a helper function to select the appropriate key for JWT validation.
func getKeyFunc(token *jwt.Token) (interface{}, error) {
	// Ensure the token method conforms to "RS256"
	if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
		return nil, jwt.NewValidationError("unexpected signing method", jwt.ValidationErrorSignatureInvalid)
	}

	kid, ok := token.Header["kid"].(string)
	if !ok {
		return nil, jwt.NewValidationError("kid header not found", jwt.ValidationErrorUnverifiable)
	}

	cacheMutex.RLock()
	defer cacheMutex.RUnlock()

	for _, key := range keysCache.Keys {
		if key.Kid == kid {
			// Decode the modulus
			nBytes, err := base64.RawURLEncoding.DecodeString(key.N)
			if err != nil {
				return nil, err // Add appropriate error handling
			}
			n := new(big.Int).SetBytes(nBytes)

			// Decode the exponent
			eBytes, err := base64.RawURLEncoding.DecodeString(key.E)
			if err != nil {
				return nil, err // Add appropriate error handling
			}
			// The exponent is usually 65537, which is a small number, so we can safely use big.Int here as well
			e := new(big.Int).SetBytes(eBytes).Int64()

			rsaKey := &rsa.PublicKey{
				N: n,
				E: int(e),
			}

			return rsaKey, nil
		}
	}

	return nil, jwt.NewValidationError("public key not found", jwt.ValidationErrorSignatureInvalid)
}

func getGooglePublicKey(token *jwt.Token) (interface{}, error) {
	err := FetchGooglePublicKeys()
	if err != nil {
		log.Println("Error fetching Google public keys:", err)
		return nil, err
	}
	return getKeyFunc(token)
}

// RequestQueue represents a queue for serializing requests
type RequestQueue struct {
	queue chan func()
}

// NewRequestQueue creates a new request queue with a specified concurrency
func NewRequestQueue(concurrency int) *RequestQueue {
	queue := &RequestQueue{
		queue: make(chan func(), 100), // Buffer size of 100 pending requests
	}

	// Start worker goroutines
	for i := 0; i < concurrency; i++ {
		go queue.worker()
	}

	return queue
}

// worker processes jobs from the queue
func (q *RequestQueue) worker() {
	for job := range q.queue {
		job()
	}
}

// QueueMiddleware creates middleware that serializes write operations
func QueueMiddleware(queue *RequestQueue) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create a channel to signal when the request is done
		done := make(chan struct{})

		// Create a timeout channel
		timeout := time.After(30 * time.Second)

		// Queue the request processing
		queue.queue <- func() {
			defer close(done)
			// This executes the next handler in the chain
			c.Next()
		}

		// Wait for completion or timeout
		select {
		case <-done:
			// Request completed normally
			log.Printf("Request processed: %s", c.Request.URL.Path)
		case <-timeout:
			// Request timed out
			log.Printf("Request timed out: %s", c.Request.URL.Path)
			c.AbortWithStatusJSON(http.StatusRequestTimeout, gin.H{"error": "Request processing timed out"})
		}
	}
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// JWTAuthMiddleware checks if the JWT token is present and valid default value to false
func JWTAuthMiddleware(requiresAdmin bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		const BEARER_SCHEMA = "Bearer "
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			log.Println("Authorization header missing")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing"})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, BEARER_SCHEMA)
		token, err := jwt.Parse(tokenString, getGooglePublicKey)
		if err != nil {
			log.Println("Error parsing token:", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		// Check if the token is expired
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorExpired != 0 {
				// Token is expired
				log.Println("Token expired")
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token is expired"})
				return
			} else {
				// Handle other validation errors
				log.Println("Error parsing token:", err)
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
				return
			}
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			// print type of claims
			if email, ok := claims["email"].(string); ok && strings.HasSuffix(email, "@g.hmc.edu") {
				if requiresAdmin { // admins are tlam, smao
					user := strings.Split(email, "@")[0]
					if user != "smao" && user != "tlam" {
						c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized access to admin endpoint"})
						return
					}
				}

				c.Set("email", email)                   // Pass the email to the next middleware or handler
				c.Set("user_full_name", claims["name"]) // Pass the user's full name to the next middleware or handler
				log.Println("Email:", email)
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized access"})
	}
}

// BlacklistCheckMiddleware checks if the user is blacklisted and blocks write operations if they are
func BlacklistCheckMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip this middleware for non-write operations or if not authenticated
		if c.Request.Method == "GET" || c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		// Get the user's email from the JWT token
		email, exists := c.Get("email")
		if !exists {
			c.Next() // If not authenticated, let the auth middleware handle it
			return
		}

		// Check if the user is blacklisted
		var isBlacklisted bool
		err := database.DB.QueryRow("SELECT is_blacklisted FROM user_rate_limits WHERE email = $1", email).Scan(&isBlacklisted)
		if err != nil {
			if err == sql.ErrNoRows {
				// User not in the rate limits table yet, so not blacklisted
				isBlacklisted = false
			} else {
				log.Printf("Error checking blacklist status for %s: %v", email, err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
				c.Abort()
				return
			}
		}

		if isBlacklisted {
			log.Printf("Blocked request from blacklisted user: %s", email)
			c.JSON(http.StatusForbidden, gin.H{
				"error":       "Your account has been temporarily restricted due to unusual activity. Please contact an administrator.",
				"blacklisted": true,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
