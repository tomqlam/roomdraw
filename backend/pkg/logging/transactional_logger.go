// File: pkg/logging/transactional_logger.go
package logging

import (
	"encoding/json"
	"log"
	"roomdraw/backend/pkg/database" // Ensure this path is correct

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lib/pq" // Import for handling potential json marshaling of pq types if needed
)

// LogOperation records a transaction log entry into the database.
// It retrieves user information and request ID from the Gin context.
func LogOperation(
	ctx *gin.Context,
	operationType string,
	entityType string,
	entityID string,
	previousState interface{}, // State before the operation
	newState interface{}, // State after the operation
	details map[string]interface{}, // Additional context-specific details
) error {
	// Get user information from the context (set by JWTAuthMiddleware)
	emailInterface, _ := ctx.Get("email")
	userNameInterface, _ := ctx.Get("user_full_name")

	// Use type assertion, default to empty string if not found or wrong type
	email, _ := emailInterface.(string)
	userName, _ := userNameInterface.(string)

	// If email is missing (e.g., unauthenticated endpoint somehow calling this), log a warning
	if email == "" {
		log.Printf("Warning: Attempted to log operation '%s' without user email in context for endpoint %s", operationType, ctx.Request.URL.Path)
		// Decide if you want to return an error or proceed without user info
		// return errors.New("user email not found in context for logging")
	}

	// Convert states and details to JSONB compatible format (byte slices)
	var prevStateJSON, newStateJSON, detailsJSON []byte
	var err error

	// Use helper to marshal, handling nil gracefully
	prevStateJSON, err = marshalToJSON(previousState)
	if err != nil {
		log.Printf("Error marshaling previousState for log (%s, %s): %v", operationType, entityID, err)
		// Continue logging even if state marshaling fails? Or return error?
		// return fmt.Errorf("failed to marshal previous state: %w", err)
	}

	newStateJSON, err = marshalToJSON(newState)
	if err != nil {
		log.Printf("Error marshaling newState for log (%s, %s): %v", operationType, entityID, err)
		// return fmt.Errorf("failed to marshal new state: %w", err)
	}

	detailsJSON, err = marshalToJSON(details)
	if err != nil {
		log.Printf("Error marshaling details for log (%s, %s): %v", operationType, entityID, err)
		// return fmt.Errorf("failed to marshal details: %w", err)
	}

	// Get or create a request ID from the context (set by TransactionLogMiddleware)
	requestIDInterface, exists := ctx.Get("request_id")
	var requestID uuid.UUID
	if !exists {
		// Should ideally not happen if middleware is applied correctly, but generate one as fallback
		log.Printf("Warning: request_id not found in context for %s, generating new one.", ctx.Request.URL.Path)
		requestID = uuid.New()
		ctx.Set("request_id", requestID) // Set it for potential subsequent logs in the same handler
	} else {
		var ok bool
		requestID, ok = requestIDInterface.(uuid.UUID)
		if !ok {
			log.Printf("Error: request_id in context is not a UUID for %s. Generating new one.", ctx.Request.URL.Path)
			requestID = uuid.New() // Fallback
		}
	}

	// --- Database Insertion ---
	sqlStatement := `
        INSERT INTO transaction_logs
        (operation_type, endpoint, user_email, user_name, entity_type, entity_id,
         previous_state, new_state, details, ip_address, request_id)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	_, err = database.DB.Exec(sqlStatement,
		operationType,
		ctx.Request.URL.Path,
		email,
		userName, // Might be empty if not in JWT
		entityType,
		entityID,
		jsonbOrNull(prevStateJSON), // Use helper to handle nil JSON
		jsonbOrNull(newStateJSON),
		jsonbOrNull(detailsJSON),
		ctx.ClientIP(),
		requestID,
	)

	if err != nil {
		// Log the error but don't fail the original request because of logging failure
		log.Printf("ERROR: Failed to insert transaction log for %s (%s): %v", operationType, entityID, err)
		// Return the error so the calling handler knows logging failed, but the handler should decide whether to proceed.
		return err
	}

	log.Printf("INFO: Logged operation '%s' for entity %s:%s by user %s (RequestID: %s)",
		operationType, entityType, entityID, email, requestID)

	return nil // Log successfully inserted
}

// TransactionLogMiddleware adds a unique request ID to the context for write operations.
func TransactionLogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip adding request ID for read operations (GET, OPTIONS)
		// LogOperation itself won't be called for reads anyway, but this keeps context clean.
		if c.Request.Method == "GET" || c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		// Add a unique request ID to the context for this request
		requestID := uuid.New()
		c.Set("request_id", requestID)
		// log.Printf("DEBUG: Set request_id %s for %s", requestID, c.Request.URL.Path) // Optional: Debug logging

		// Process request
		c.Next()

		// You could potentially add a log here *after* the request is processed
		// to capture the final status code, but LogOperation is designed
		// to be called *within* the handler where the change occurs.
	}
}

// marshalToJSON safely marshals an interface to JSON bytes, returning nil if the input is nil.
func marshalToJSON(data interface{}) ([]byte, error) {
	if data == nil {
		return nil, nil // Return nil bytes and nil error for nil input
	}
	// Handle specific types like pq.StringArray if necessary before general marshaling
	switch v := data.(type) {
	case pq.StringArray:
		// pq types might need special handling if json.Marshal doesn't work directly
		// In many cases, it works fine. If not, convert to []string first.
		// Example: return json.Marshal([]string(v))
		return json.Marshal(v)
	default:
		return json.Marshal(data)
	}
}

// jsonbOrNull returns the JSON byte slice or nil if the slice is empty or nil,
// suitable for inserting into nullable JSONB columns.
func jsonbOrNull(jsonData []byte) interface{} {
	if len(jsonData) == 0 || string(jsonData) == "null" { // Check for empty or explicit "null"
		return nil
	}
	return jsonData
}
