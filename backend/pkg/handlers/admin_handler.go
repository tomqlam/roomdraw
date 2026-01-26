package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"roomdraw/backend/pkg/database"
	"time"

	"roomdraw/backend/pkg/models"

	"github.com/gin-gonic/gin"
)

// BlocklistedUser represents a user who has been blocklisted
type BlocklistedUser struct {
	Email             string    `json:"email"`
	ClearRoomCount    int       `json:"clearRoomCount"`
	ClearRoomDate     string    `json:"clearRoomDate"`
	BlocklistedAt     time.Time `json:"blocklistedAt"`
	BlocklistedReason string    `json:"reason"`
}

// GetBlocklistedUsers returns a list of all blocklisted users
func GetBlocklistedUsers(c *gin.Context) {
	rows, err := database.DB.Query(`
		SELECT email, clear_room_count, clear_room_date, blocklisted_at, blocklisted_reason
		FROM user_rate_limits
		WHERE is_blocklisted = true
		ORDER BY blocklisted_at DESC
	`)
	if err != nil {
		log.Printf("Error querying blocklisted users: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve blocklisted users"})
		return
	}
	defer rows.Close()

	users := []BlocklistedUser{} // Initialize with empty array instead of nil
	for rows.Next() {
		var user BlocklistedUser
		var blocklistedAt sql.NullTime
		if err := rows.Scan(
			&user.Email,
			&user.ClearRoomCount,
			&user.ClearRoomDate,
			&blocklistedAt,
			&user.BlocklistedReason,
		); err != nil {
			log.Printf("Error scanning blocklisted user: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan blocklisted users"})
			return
		}
		// Convert NullTime to Time if it's valid
		if blocklistedAt.Valid {
			user.BlocklistedAt = blocklistedAt.Time
		} else {
			user.BlocklistedAt = time.Time{} // Zero value for time
		}
		users = append(users, user)
	}

	c.JSON(http.StatusOK, users)
}

// RemoveUserBlocklist removes a user from the blocklist
func RemoveUserBlocklist(c *gin.Context) {
	email := c.Param("email")

	// Start a transaction
	tx, err := database.DB.Begin()
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}

	// Ensure the transaction is either committed or rolled back
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	_, err = tx.Exec(`
		UPDATE user_rate_limits
		SET is_blocklisted = false,
		    clear_room_count = 0,
		    clear_room_date = CURRENT_TIMESTAMP
		WHERE email = $1
	`, email)

	if err != nil {
		log.Printf("Error removing user from blocklist: %v", err)
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove user from blocklist"})
		return
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		log.Printf("Error committing transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User removed from blocklist", "email": email})
}

// GetUserClearRoomStats gets a user's clear room usage stats
func GetUserClearRoomStats(c *gin.Context) {
	email, exists := c.Get("email")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User email not found"})
		return
	}
	emailStr := email.(string)

	// Get Pacific time
	loc, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		log.Printf("Error loading Pacific timezone: %v", err)
		// Fall back to UTC if unable to load Pacific time
		loc = time.UTC
	}
	pacificNow := time.Now().In(loc)
	today := pacificNow.Format("2006-01-02") // YYYY-MM-DD format
	todayDate, _ := time.Parse("2006-01-02", today)

	const MAX_DAILY_CLEARS = 10

	// Define a UserRateLimit instance
	var userLimit models.UserRateLimit

	// Get the user's clear room stats
	err = database.DB.QueryRow(`
		SELECT email, clear_room_count, clear_room_date, is_blocklisted, blocklisted_at, blocklisted_reason
		FROM user_rate_limits
		WHERE email = $1
	`, emailStr).Scan(
		&userLimit.Email,
		&userLimit.ClearRoomCount,
		&userLimit.ClearRoomDate,
		&userLimit.IsBlocklisted,
		&userLimit.BlocklistedAt,
		&userLimit.BlocklistedReason,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			// If no record exists, user hasn't cleared any rooms yet
			userLimit.Email = emailStr
			userLimit.ClearRoomCount = 0
			userLimit.ClearRoomDate.Valid = true
			userLimit.ClearRoomDate.Time = todayDate
			userLimit.IsBlocklisted = false
		} else {
			log.Printf("Error retrieving clear room stats: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve rate limit information"})
			return
		}
	}

	// Determine record date and format it
	recordDateStr := today // Default to today
	if userLimit.ClearRoomDate.Valid {
		recordDateStr = userLimit.ClearRoomDate.Time.Format("2006-01-02")
	}

	// Log the values for debugging
	log.Printf("For user %s: clearCount=%d, recordDate=%s, isBlocklisted=%v",
		emailStr, userLimit.ClearRoomCount, recordDateStr, userLimit.IsBlocklisted)

	// Check if we need to reset based on date
	if recordDateStr != today {
		// This would be reset on next operation, but for display purposes we'll show it as reset
		userLimit.ClearRoomCount = 0
	}

	// Calculate time until midnight Pacific Time reset
	midnight := time.Date(pacificNow.Year(), pacificNow.Month(), pacificNow.Day(), 0, 0, 0, 0, loc)
	if pacificNow.After(midnight) {
		midnight = midnight.Add(24 * time.Hour)
	}
	minutesUntilReset := int(midnight.Sub(pacificNow).Minutes())

	c.JSON(http.StatusOK, gin.H{
		"clearRoomCount":  userLimit.ClearRoomCount,
		"maxDailyClears":  MAX_DAILY_CLEARS,
		"remainingClears": MAX_DAILY_CLEARS - userLimit.ClearRoomCount,
		"resetsInMinutes": minutesUntilReset,
		"pacificDate":     today,
		"isBlocklisted":   userLimit.IsBlocklisted,
	})
}
