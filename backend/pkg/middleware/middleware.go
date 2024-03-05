package middleware

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"log"
	"math/big"
	"net/http"
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

func QueueMiddleware(requestQueue chan<- *gin.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		// if the request is a write operation, enqueue the request
		// log.Printf("QueueMiddleware triggered for path: %s", c.Request.URL.Path)
		doneChan := make(chan bool, 1) // Channel to signal completion of request processing
		c.Set("doneChan", doneChan)    // Pass the channel along with the context
		requestQueue <- c              // Enqueue the context
	}
}

func RequestProcessor(requestQueue <-chan *gin.Context) {
	for c := range requestQueue {
		// Wait for the request to be processed
		doneChan, exists := c.Get("doneChan")
		if !exists {
			log.Print("Error: doneChan not found in context for request processor +", c.Request.URL.Path)
			continue
		}

		c.Next() // Process the request

		// set timeout time to 5 seconds
		timeout := time.After(5 * time.Second)

		// Wait for the doneChan to receive a signal
		select {
		case <-doneChan.(chan bool):
			log.Print("Request processed +", c.Request.URL.Path)
		case <-timeout:
			log.Print("Request processing timed out +", c.Request.URL.Path)
			close(doneChan.(chan bool))
		}

		// log.Print("Finished processing request +", c.Request.URL.Path)

		// log.Print("Sleeping for 10 seconds")
		// time.Sleep(10 * time.Second) // Simulate processing time
		// log.Print("Woke up")
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
