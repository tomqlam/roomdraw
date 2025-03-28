package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"roomdraw/backend/pkg/database"
	"time"

	"github.com/gin-gonic/gin"
	"roomdraw/backend/pkg/models"
)

// BlacklistedUser represents a user who has been blacklisted
type BlacklistedUser struct {
	Email             string    `json:"email"`
	ClearRoomCount    int       `json:"clearRoomCount"`
	ClearRoomDate     string    `json:"clearRoomDate"`
	BlacklistedAt     time.Time `json:"blacklistedAt"`
	BlacklistedReason string    `json:"reason"`
}

// GetBlacklistedUsers returns a list of all blacklisted users
func GetBlacklistedUsers(c *gin.Context) {
	rows, err := database.DB.Query(`
		SELECT email, clear_room_count, clear_room_date, blacklisted_at, blacklisted_reason 
		FROM user_rate_limits 
		WHERE is_blacklisted = true
		ORDER BY blacklisted_at DESC
	`)
	if err != nil {
		log.Printf("Error querying blacklisted users: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve blacklisted users"})
		return
	}
	defer rows.Close()

	users := []BlacklistedUser{} // Initialize with empty array instead of nil
	for rows.Next() {
		var user BlacklistedUser
		var blacklistedAt sql.NullTime
		if err := rows.Scan(
			&user.Email,
			&user.ClearRoomCount,
			&user.ClearRoomDate,
			&blacklistedAt,
			&user.BlacklistedReason,
		); err != nil {
			log.Printf("Error scanning blacklisted user: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan blacklisted users"})
			return
		}
		// Convert NullTime to Time if it's valid
		if blacklistedAt.Valid {
			user.BlacklistedAt = blacklistedAt.Time
		} else {
			user.BlacklistedAt = time.Time{} // Zero value for time
		}
		users = append(users, user)
	}

	c.JSON(http.StatusOK, users)
}

// RemoveUserBlacklist removes a user from the blacklist
func RemoveUserBlacklist(c *gin.Context) {
	email := c.Param("email")

	_, err := database.DB.Exec(`
		UPDATE user_rate_limits
		SET is_blacklisted = false, 
		    clear_room_count = 0, 
		    clear_room_date = CURRENT_TIMESTAMP
		WHERE email = $1
	`, email)

	if err != nil {
		log.Printf("Error removing user from blacklist: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove user from blacklist"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User removed from blacklist", "email": email})
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
		SELECT email, clear_room_count, clear_room_date, is_blacklisted, blacklisted_at, blacklisted_reason
		FROM user_rate_limits
		WHERE email = $1
	`, emailStr).Scan(
		&userLimit.Email,
		&userLimit.ClearRoomCount,
		&userLimit.ClearRoomDate,
		&userLimit.IsBlacklisted,
		&userLimit.BlacklistedAt,
		&userLimit.BlacklistedReason,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			// If no record exists, user hasn't cleared any rooms yet
			userLimit.Email = emailStr
			userLimit.ClearRoomCount = 0
			userLimit.ClearRoomDate.Valid = true
			userLimit.ClearRoomDate.Time = todayDate
			userLimit.IsBlacklisted = false
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
	log.Printf("For user %s: clearCount=%d, recordDate=%s, isBlacklisted=%v", 
		emailStr, userLimit.ClearRoomCount, recordDateStr, userLimit.IsBlacklisted)

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
		"isBlacklisted":   userLimit.IsBlacklisted,
	})
}
