package handlers

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"roomdraw/backend/pkg/config"
	"roomdraw/backend/pkg/database"
	"roomdraw/backend/pkg/models"
	"strings"

	"git.sr.ht/~jamesponddotco/bunnystorage-go"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

func SetSuiteDesign(c *gin.Context) {
	suiteUUID := c.Param("suiteuuid")

	tx, err := database.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}

	// Ensure the transaction is either committed or rolled back
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	// ensure the suite exists
	var suiteExists bool
	err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM suites WHERE suite_uuid = $1)", suiteUUID).Scan(&suiteExists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check if suite exists"})
		return
	}

	if !suiteExists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Suite not found"})
		return
	}

	// query for the current suite design
	var currentSuiteDesign string
	err = tx.QueryRow("SELECT suite_design FROM suites WHERE suite_uuid = $1", suiteUUID).Scan(&currentSuiteDesign)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get current suite design"})
		return
	}

	// extract file name from URL
	currentSuiteDesign = currentSuiteDesign[strings.LastIndex(currentSuiteDesign, "/")+1:]

	cfg := &bunnystorage.Config{
		StorageZone: config.BunnyNetStorageZone,
		Key:         config.BunnyNetWriteAPIKey,
		ReadOnlyKey: config.BunnyNetReadAPIKey,
		Endpoint:    bunnystorage.EndpointLosAngeles,
	}

	client, err := bunnystorage.NewClient(cfg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create BunnyStorage client"})
		return
	}

	fileHeader, err := c.FormFile("suite_design")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No suite design file uploaded"})
		return
	}

	// Open the file
	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open suite design file"})
		return
	}
	defer file.Close()

	// Read the first 512 bytes to determine the content type
	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read the file"})
		return
	}

	// Reset the read pointer back to the start of the file
	_, err = file.Seek(0, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reset file read pointer"})
		return
	}

	contentType := http.DetectContentType(buffer)
	// Check if the content type is one of the allowed types
	if contentType != "image/svg+xml" && contentType != "image/png" && contentType != "image/jpeg" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File must be an SVG, PNG, or JPEG image"})
		return
	}

	// create random filename
	imageFilename := uuid.New().String()

	if contentType == "image/png" {
		imageFilename += ".png"
	} else if contentType == "image/jpeg" {
		imageFilename += ".jpg"
	} else {
		imageFilename += ".svg"
	}

	// upload the suite design to BunnyStorage
	uploadRes, err := client.Upload(c, "suite_designs", imageFilename, "", file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload suite design"})
		return
	}

	if uploadRes.Status != http.StatusCreated {
		log.Println("uploadRes", uploadRes)
		log.Println("Failed to upload suite design to BunnyStorage:", uploadRes.Status)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload suite design"})
		return
	}

	imageUrl := config.CDNURL + "/suite_designs/" + imageFilename

	log.Println(suiteUUID)

	// Add a new link to the CDN to the suite design
	_, err = tx.Exec("UPDATE suites SET suite_design = $1 WHERE suite_uuid = $2", imageUrl, suiteUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update suite design"})
		return
	}

	log.Println("Uploaded suite design to BunnyStorage:", imageUrl)

	// commit the transaction
	err = tx.Commit()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Suite design updated"})
}

func DeleteSuiteDesign(c *gin.Context) {
	suiteUUID := c.Param("suiteuuid")
	tx, err := database.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}

	// Ensure the transaction is either committed or rolled back
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	// ensure the suite exists
	var suiteExists bool
	err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM suites WHERE suite_uuid = $1)", suiteUUID).Scan(&suiteExists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check if suite exists"})
		return
	}

	if !suiteExists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Suite not found"})
		return
	}

	// remove the suite design from the suite default ''
	_, err = tx.Exec("UPDATE suites SET suite_design = '' WHERE suite_uuid = $1", suiteUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove suite design"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Suite design deleted"})
}

func UpdateSuiteGenderPreference(c *gin.Context) {
	tx, err := database.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}

	// Ensure the transaction is either committed or rolled back
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	// Get all suites that can be gender preferenced
	suiteRows, err := tx.Query("SELECT suite_uuid FROM suites WHERE can_be_gender_preferenced = true")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get gender-preferenceable suites"})
		return
	}

	// Process each suite and put in into a list
	var suiteUUIDs []uuid.UUID
	for suiteRows.Next() {
		var suiteUUID uuid.UUID
		err = suiteRows.Scan(&suiteUUID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan suite data"})
			return
		}
		suiteUUIDs = append(suiteUUIDs, suiteUUID)
	}

	suiteRows.Close()

	// Use the helper function to update gender preferences for each suite
	for _, suiteUUID := range suiteUUIDs {
		err = UpdateSuiteGenderPreferencesBySuiteUUID(tx, suiteUUID)
		if err != nil {
			log.Printf("Failed to update gender preferences for suite %s: %v", suiteUUID, err)
			continue
		}
	}
	c.JSON(http.StatusOK, gin.H{"message": "Suite gender preferences updated"})
}

// UpdateSuiteGenderPreferencesBySuiteUUID is a helper function that updates a suite's gender preferences
// based on its occupants. This should be called after any changes to room occupants.
func UpdateSuiteGenderPreferencesBySuiteUUID(tx *sql.Tx, suiteUUID uuid.UUID) error {
	// Check if the suite can be gender preferenced
	var canBeGenderPreferenced bool
	err := tx.QueryRow("SELECT can_be_gender_preferenced FROM suites WHERE suite_uuid = $1", suiteUUID).Scan(&canBeGenderPreferenced)
	if err != nil {
		log.Printf("Failed to check if suite %s can be gender preferenced: %v", suiteUUID, err)
		return err
	}

	// If the suite can't be gender preferenced, no need to proceed
	if !canBeGenderPreferenced {
		return nil
	}

	// Get the dorm ID for the suite
	var dormId int
	err = tx.QueryRow("SELECT dorm FROM suites WHERE suite_uuid = $1", suiteUUID).Scan(&dormId)
	if err != nil {
		log.Printf("Failed to get dorm ID for suite %s: %v", suiteUUID, err)
		return err
	}

	// Get all rooms in the suite
	var roomUUIDs models.UUIDArray
	err = tx.QueryRow("SELECT rooms FROM suites WHERE suite_uuid = $1", suiteUUID).Scan(&roomUUIDs)
	if err != nil {
		log.Printf("Failed to get rooms for suite %s: %v", suiteUUID, err)
		return err
	}

	// Get all users in the suite - directly join with the rooms table to ensure we only get actual room occupants
	var users []models.UserRaw
	for _, roomUUID := range roomUUIDs {
		// Get occupants directly from the rooms table for this room
		var occupantIds models.IntArray
		err = tx.QueryRow("SELECT occupants FROM rooms WHERE room_uuid = $1", roomUUID).Scan(&occupantIds)
		if err != nil {
			log.Printf("Failed to get occupants for room %s: %v", roomUUID, err)
			continue
		}

		// Skip rooms with no occupants
		if len(occupantIds) == 0 {
			continue
		}

		// Get user data for each occupant
		for _, occupantId := range occupantIds {
			// Verify user is actually in this room
			var userRoomUUID uuid.UUID
			err = tx.QueryRow("SELECT room_uuid FROM users WHERE id = $1", occupantId).Scan(&userRoomUUID)
			if err != nil {
				log.Printf("Failed to check room for user ID %d: %v", occupantId, err)
				continue
			}

			// Skip if user isn't actually in this room anymore
			if userRoomUUID != roomUUID {
				log.Printf("User %d is not in room %s (in room %s instead), skipping", occupantId, roomUUID, userRoomUUID)
				continue
			}

			var user models.UserRaw
			err = tx.QueryRow(`
				SELECT id, year, first_name, last_name, email, draw_number, preplaced, in_dorm, 
				sgroup_uuid, participated, participation_time, room_uuid, reslife_role, 
				notifications_enabled, notification_created_at, notification_updated_at, gender_preferences
				FROM users WHERE id = $1
			`, occupantId).Scan(
				&user.Id, &user.Year, &user.FirstName, &user.LastName, &user.Email,
				&user.DrawNumber, &user.Preplaced, &user.InDorm, &user.SGroupUUID,
				&user.Participated, &user.PartitipationTime, &user.RoomUUID, &user.ReslifeRole,
				&user.NotificationsEnabled, &user.NotificationCreatedAt, &user.NotificationUpdatedAt,
				&user.GenderPreferences,
			)
			if err != nil {
				log.Printf("Failed to get user data for ID %d: %v", occupantId, err)
				continue
			}
			users = append(users, user)
		}
	}

	// Log the actual user list we're using for gender preference calculation
	userNames := make([]string, 0, len(users))
	for _, u := range users {
		userNames = append(userNames, fmt.Sprintf("%s %s (ID: %d)", u.FirstName, u.LastName, u.Id))
	}
	log.Printf("Calculating gender preferences for suite %s based on users: %s", suiteUUID, strings.Join(userNames, ", "))

	// Determine the gender preferences for this suite
	genderPreferences, found := GetSuiteGenderPreference(users, dormId)
	if found {
		// Update the suite's gender preferences
		_, err = tx.Exec("UPDATE suites SET gender_preferences = $1 WHERE suite_uuid = $2",
			pq.Array(genderPreferences), suiteUUID)
		if err != nil {
			log.Printf("Failed to update gender preferences for suite %s: %v", suiteUUID, err)
			return err
		}
		log.Printf("Updated gender preferences for suite %s to %v", suiteUUID, genderPreferences)
		return nil
	} else {
		// Check if there are any preplaced users with gender preferences
		var hasPreplacedUsersWithPreferences bool
		for _, user := range users {
			if user.Preplaced && len(user.GenderPreferences) > 0 {
				hasPreplacedUsersWithPreferences = true
				break
			}
		}

		// If there are preplaced users with preferences but no valid intersection was found,
		// return a specific error
		if hasPreplacedUsersWithPreferences {
			err := errors.New("no valid intersection of gender preferences found between preplaced users")
			log.Printf("Error in suite %s: %v", suiteUUID, err)
			return err
		}

		// Clear the suite gender preferences since there are no valid preferences
		_, err = tx.Exec("UPDATE suites SET gender_preferences = $1 WHERE suite_uuid = $2",
			pq.Array([]string{}), suiteUUID)
		if err != nil {
			log.Printf("Failed to clear gender preferences for suite %s: %v", suiteUUID, err)
			return err
		}
		log.Printf("Cleared gender preferences for suite %s as no users have preferences", suiteUUID)

		// Otherwise, it's fine to have no preferences
		return nil
	}
}
