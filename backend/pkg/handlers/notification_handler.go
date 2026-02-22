package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"roomdraw/backend/pkg/database"
	"roomdraw/backend/pkg/models"
	"roomdraw/backend/pkg/services"
	"time"

	"github.com/gin-gonic/gin"
)

var emailService *services.EmailService

// InitializeEmailService initializes the email service with loaded configuration
func InitializeEmailService() {
	log.Println("Initializing email service from handlers...")
	emailService = services.NewEmailService()
}

func SetNotificationPreference(c *gin.Context) {
	var pref struct {
		Enabled bool `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&pref); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Use the authenticated email from JWT context instead of client-supplied email
	email, exists := c.Get("email")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}
	userEmail, ok := email.(string)
	if !ok || userEmail == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authentication token"})
		return
	}

	tx, err := database.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	now := time.Now()
	_, err = tx.Exec(`
		UPDATE users
		SET notifications_enabled = $1,
			notification_updated_at = $2
		WHERE email = $3`,
		pref.Enabled, now, userEmail)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update notification preferences"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification preferences updated"})
}

func GetNotificationPreference(c *gin.Context) {
	userEmail := c.Query("email")
	if userEmail == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email parameter is required"})
		return
	}

	var pref struct {
		Email   string `json:"email"`
		Enabled bool   `json:"enabled"`
	}
	err := database.DB.QueryRow(
		"SELECT email, notifications_enabled FROM users WHERE email = $1",
		userEmail,
	).Scan(&pref.Email, &pref.Enabled)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Notification preferences not found"})
		return
	}

	c.JSON(http.StatusOK, pref)
}

func SendBumpNotification(userID int, roomID string, dormName string) {
	var user models.UserRaw
	var email sql.NullString
	err := database.DB.QueryRow(
		"SELECT id, first_name, last_name, email, notifications_enabled FROM users WHERE id = $1",
		userID,
	).Scan(&user.Id, &user.FirstName, &user.LastName, &email, &user.NotificationsEnabled)

	if err != nil {
		log.Printf("Failed to fetch user: %v", err)
		return // User not found
	}

	// Handle NULL email
	if !email.Valid || email.String == "" {
		log.Printf("User %d has no email, skipping notification", userID)
		return
	}
	user.Email = email.String

	log.Println("User: ", user)

	log.Println("Checking if user has opted in ...")

	if !user.NotificationsEnabled {
		log.Println("User hasn't opted in or preferences not found")
		return // User hasn't opted in or preferences not found
	}

	err = emailService.SendBumpNotification(user, roomID, dormName)
	if err != nil {
		log.Printf("Failed to send bump notification: %v", err)
	}
}
