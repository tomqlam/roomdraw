package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"roomdraw/backend/pkg/database"
	"roomdraw/backend/pkg/logging"
	"roomdraw/backend/pkg/models"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func GetRoomsHandler(c *gin.Context) {
	// Start a transaction
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

	// Example SQL query
	rows, err := tx.Query("SELECT room_uuid, dorm, dorm_name, room_id, suite_uuid, max_occupancy, current_occupancy, occupants, pull_priority, sgroup_uuid, has_frosh, frosh_room_type FROM rooms")
	if err != nil {

		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed"})
		return
	}

	var rooms []models.RoomRaw
	for rows.Next() {
		var d models.RoomRaw
		if err := rows.Scan(&d.RoomUUID, &d.Dorm, &d.DormName, &d.RoomID, &d.SuiteUUID, &d.MaxOccupancy, &d.CurrentOccupancy, &d.Occupants, &d.PullPriority, &d.SGroupUUID, &d.HasFrosh, &d.FroshRoomType); err != nil {
			// Handle scan error
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database scan failed"})
			return
		}
		rooms = append(rooms, d)
	}

	c.JSON(http.StatusOK, rooms)
}

func GetSimpleFormattedDorm(c *gin.Context) {
	dormNameParam := c.Param("dormName")

	// Start a transaction
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

	rows, err := tx.Query("SELECT room_uuid, dorm, dorm_name, room_id, suite_uuid, max_occupancy, current_occupancy, occupants, pull_priority, has_frosh, frosh_room_type FROM rooms WHERE UPPER(dorm_name) = UPPER($1)", dormNameParam)
	if err != nil {

		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed on rooms"})
		return
	}

	var rooms []models.RoomRaw
	for rows.Next() {
		var d models.RoomRaw
		if err := rows.Scan(&d.RoomUUID, &d.Dorm, &d.DormName, &d.RoomID, &d.SuiteUUID, &d.MaxOccupancy, &d.CurrentOccupancy, &d.Occupants, &d.PullPriority, &d.HasFrosh, &d.FroshRoomType); err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database scan failed on rooms"})
			return
		}
		rooms = append(rooms, d)
	}

	rows, err = tx.Query("SELECT suite_uuid, dorm, dorm_name, floor, room_count, rooms, alternative_pull, suite_design, can_lock_pull, reslife_room, gender_preferences FROM suites WHERE UPPER(dorm_name) = UPPER($1)", dormNameParam)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed on suites"})
		return
	}

	var suites []models.SuiteRaw
	for rows.Next() {
		var s models.SuiteRaw
		if err := rows.Scan(&s.SuiteUUID, &s.Dorm, &s.DormName, &s.Floor, &s.RoomCount, &s.Rooms, &s.AlternativePull, &s.SuiteDesign, &s.CanLockPull, &s.ReslifeRoom, &s.GenderPreferences); err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database scan failed on suites"})
			return
		}
		suites = append(suites, s)
	}

	// Check if any error occurred during queries or scans
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database operation failed"})
		return
	}

	var suiteUUIDToFloorMap = make(map[uuid.UUID]int)
	for _, s := range suites {
		suiteUUIDToFloorMap[s.SuiteUUID] = s.Floor
	}

	var suiteToRoomMap = make(map[string][]models.RoomSimple)
	for _, r := range rooms {
		suiteUUIDString := r.SuiteUUID.String()
		room := models.RoomSimple{
			RoomNumber:    r.RoomID,
			PullPriority:  r.PullPriority,
			MaxOccupancy:  r.MaxOccupancy,
			RoomUUID:      r.RoomUUID,
			HasFrosh:      r.HasFrosh,
			FroshRoomType: r.FroshRoomType,
			Dorm:          r.Dorm,
		}

		if len(r.Occupants) >= 1 {
			room.Occupant1 = r.Occupants[0]
		}
		if len(r.Occupants) >= 2 {
			room.Occupant2 = r.Occupants[1]
		}
		if len(r.Occupants) >= 3 {
			room.Occupant3 = r.Occupants[2]
		}
		if len(r.Occupants) >= 4 {
			room.Occupant4 = r.Occupants[3]
		}

		suiteToRoomMap[suiteUUIDString] = append(suiteToRoomMap[suiteUUIDString], room)
	}

	var floorMap = make(map[int][]models.SuiteSimple)

	for _, s := range suites {
		suiteUUIDString := s.SuiteUUID.String()
		floor := suiteUUIDToFloorMap[s.SuiteUUID]
		suite := models.SuiteSimple{
			Rooms:             suiteToRoomMap[suiteUUIDString],
			SuiteDesign:       s.SuiteDesign,
			SuiteUUID:         s.SuiteUUID,
			AlternativePull:   s.AlternativePull,
			CanLockPull:       s.CanLockPull,
			GenderPreferences: s.GenderPreferences,
		}

		floorMap[floor] = append(floorMap[floor], suite)
	}

	// now put everything in the dorm struct
	var dorm models.DormSimple

	dorm.DormName = cases.Title(language.English).String(dormNameParam)
	dorm.Description = "This is a description of " + dorm.DormName + "."

	// for all floors (using the map keys)
	for i, floor := range floorMap {
		var floorSimple models.FloorSimple
		floorSimple.Suites = floor
		floorSimple.FloorNumber = i
		dorm.Floors = append(dorm.Floors, floorSimple)
	}

	c.JSON(http.StatusOK, dorm)
}

func GetSimplerFormattedDorm(c *gin.Context) {
	dormNameParam := c.Param("dormName")

	// Start a transaction
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

	rows, err := tx.Query("SELECT room_uuid, dorm, dorm_name, room_id, suite_uuid, max_occupancy, current_occupancy, occupants, pull_priority FROM rooms WHERE UPPER(dorm_name) = UPPER($1)", dormNameParam)
	if err != nil {

		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed on rooms"})
		return
	}

	var rooms []models.RoomRaw
	for rows.Next() {
		var d models.RoomRaw
		if err := rows.Scan(&d.RoomUUID, &d.Dorm, &d.DormName, &d.RoomID, &d.SuiteUUID, &d.MaxOccupancy, &d.CurrentOccupancy, &d.Occupants, &d.PullPriority); err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database scan failed on rooms"})
			return
		}
		rooms = append(rooms, d)
	}

	rows, err = tx.Query("SELECT suite_uuid, dorm, dorm_name, floor, room_count, rooms, alternative_pull, suite_design FROM suites WHERE UPPER(dorm_name) = UPPER($1)", dormNameParam)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed on suites"})
		return
	}

	var suites []models.SuiteRaw
	for rows.Next() {
		var s models.SuiteRaw
		if err := rows.Scan(&s.SuiteUUID, &s.Dorm, &s.DormName, &s.Floor, &s.RoomCount, &s.Rooms, &s.AlternativePull, &s.SuiteDesign); err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database scan failed on suites"})
			return
		}
		suites = append(suites, s)
	}

	// Check if any error occurred during queries or scans
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database operation failed"})
		return
	}

	var suiteUUIDToFloorMap = make(map[uuid.UUID]int)
	for _, s := range suites {
		suiteUUIDToFloorMap[s.SuiteUUID] = s.Floor
	}

	var suiteToRoomMap = make(map[string][]models.RoomSimpler)
	for _, r := range rooms {
		suiteUUIDString := r.SuiteUUID.String()
		room := models.RoomSimpler{
			RoomNumber:    r.RoomID,
			MaxOccupancy:  r.MaxOccupancy,
			FroshRoomType: r.FroshRoomType,
		}

		suiteToRoomMap[suiteUUIDString] = append(suiteToRoomMap[suiteUUIDString], room)
	}

	var floorMap = make(map[int][]models.SuiteSimpler)

	for _, s := range suites {
		suiteUUIDString := s.SuiteUUID.String()
		floor := suiteUUIDToFloorMap[s.SuiteUUID]
		suite := models.SuiteSimpler{
			Rooms:           suiteToRoomMap[suiteUUIDString],
			AlternativePull: s.AlternativePull,
		}

		floorMap[floor] = append(floorMap[floor], suite)
	}

	// now put everything in the dorm struct
	var dorm models.DormSimpler

	// for all floors (using the map keys)
	for _, floor := range floorMap {
		var floorSimple models.FloorSimpler
		floorSimple.Suites = floor
		dorm.Floors = append(dorm.Floors, floorSimple)
	}

	c.JSON(http.StatusOK, dorm)
}

func getRoomStateRaw(roomUUID string) (*models.RoomRaw, error) {
	var room models.RoomRaw
	// Include all columns that represent the state you want to log
	err := database.DB.QueryRow(`
        SELECT room_uuid, dorm, dorm_name, room_id, suite_uuid, max_occupancy,
               current_occupancy, occupants, pull_priority, sgroup_uuid, has_frosh, frosh_room_type
        FROM rooms
        WHERE room_uuid = $1`, roomUUID).Scan(
		&room.RoomUUID, &room.Dorm, &room.DormName, &room.RoomID, &room.SuiteUUID,
		&room.MaxOccupancy, &room.CurrentOccupancy, &room.Occupants, &room.PullPriority,
		&room.SGroupUUID, &room.HasFrosh, &room.FroshRoomType,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Room not found is a valid state (doesn't exist)
		}
		return nil, fmt.Errorf("failed to query room state for %s: %w", roomUUID, err)
	}
	return &room, nil
}

func compareRoomStatesForLogging(state1, state2 *models.RoomRaw) bool {
	if state1 == nil || state2 == nil {
		return state1 == state2 // Both nil is considered same, one nil isn't
	}
	// Compare relevant fields: occupants, pull_priority, sgroup_uuid, etc.
	// Example:
	// occupantsSame := compareIntArrays(state1.Occupants.Elements, state2.Occupants.Elements) // Need helper for array compare
	// prioritySame := state1.PullPriority == state2.PullPriority // Assumes PullPriority is comparable or implement deep compare
	// sgroupSame := state1.SGroupUUID == state2.SGroupUUID
	// return occupantsSame && prioritySame && sgroupSame // Add more fields as needed

	// For now, assume any difference means changed state for simplicity
	state1Json, _ := json.Marshal(state1)
	state2Json, _ := json.Marshal(state2)
	return string(state1Json) == string(state2Json)

}

func ToggleInDorm(c *gin.Context) {
	roomUUIDParam := c.Param("roomuuid")
	_, err := uuid.Parse(roomUUIDParam) // Validate UUID format early
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid room UUID format"})
		return
	}

	// --- Transactional Logging: Get Previous State ---
	previousRoomState, err := getRoomStateRaw(roomUUIDParam)
	if err != nil {
		log.Printf("Error fetching previous room state for TOGGLE_IN_DORM %s: %v", roomUUIDParam, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve current room state"})
		return
	}
	if previousRoomState == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}

	// Start a transaction
	tx, err := database.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}

	// Defer rollback and handle commit/error for logging
	var commitErr error
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			// Re-panic if needed, or handle
			log.Printf("PANIC during TOGGLE_IN_DORM for %s: %v", roomUUIDParam, r)
			// Ensure response indicates server error if not already sent
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error due to panic"})
			return
		}
		if err != nil { // Error occurred within the handler logic before commit
			log.Printf("Rolling back transaction for TOGGLE_IN_DORM %s due to error: %v", roomUUIDParam, err)
			tx.Rollback()
			// We won't log if an error caused rollback before commit
			return
		}
		// No error before commit, attempt commit
		commitErr = tx.Commit()
		if commitErr != nil {
			log.Printf("Failed to commit transaction for TOGGLE_IN_DORM %s: %v", roomUUIDParam, commitErr)
			// Don't log if commit failed
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database transaction commit error"})
			return
		}

		// --- Transactional Logging: Get New State & Log (ONLY IF COMMIT SUCCEEDED) ---
		newRoomState, fetchErr := getRoomStateRaw(roomUUIDParam)
		if fetchErr != nil {
			log.Printf("Error fetching new room state for TOGGLE_IN_DORM %s after commit: %v", roomUUIDParam, fetchErr)
			// Log the operation anyway, but new state might be nil/incomplete
		}

		logDetails := map[string]interface{}{
			"previous_pull_priority": previousRoomState.PullPriority, // Log the specific part that changed
			// Add any other relevant details if needed
		}

		loggingErr := logging.LogOperation(
			c,                     // Pass the Gin context
			"TOGGLE_IN_DORM",      // Operation Type
			models.EntityTypeRoom, // Entity Type
			roomUUIDParam,         // Entity ID
			previousRoomState,     // State Before (fetched before tx)
			newRoomState,          // State After (fetched after successful commit)
			logDetails,            // Additional Details
		)
		if loggingErr != nil {
			// Log the logging error, but the main operation succeeded.
			log.Printf("WARNING: Failed to log TOGGLE_IN_DORM operation for %s: %v", roomUUIDParam, loggingErr)
		}

		// Send success response *after* logging attempt
		c.JSON(http.StatusOK, gin.H{"message": "Successfully toggled in dorm status for room " + roomUUIDParam})

	}() // End of defer func

	// get the current room's into
	var currentRoomInfo models.RoomRaw
	err = tx.QueryRow("SELECT room_uuid, dorm, dorm_name, room_id, suite_uuid, max_occupancy, current_occupancy, occupants, pull_priority, sgroup_uuid, has_frosh, frosh_room_type FROM rooms WHERE room_uuid = $1", roomUUIDParam).Scan(
		&currentRoomInfo.RoomUUID,
		&currentRoomInfo.Dorm,
		&currentRoomInfo.DormName,
		&currentRoomInfo.RoomID,
		&currentRoomInfo.SuiteUUID,
		&currentRoomInfo.MaxOccupancy,
		&currentRoomInfo.CurrentOccupancy,
		&currentRoomInfo.Occupants,
		&currentRoomInfo.PullPriority,
		&currentRoomInfo.SGroupUUID,
		&currentRoomInfo.HasFrosh,
		&currentRoomInfo.FroshRoomType,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query room info from rooms table"})
		return
	}

	// check if the Year property of the PullPriority is 4 and the hasInDorm property is True
	if currentRoomInfo.PullPriority.Year != 4 || !currentRoomInfo.PullPriority.HasInDorm {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot toggle in dorm for a user that is not a senior with in dorm"})
		err = errors.New("cannot toggle in dorm for a user that is not a senior with in dorm")
		return
	}

	// check if they are part of a suite group
	if currentRoomInfo.SGroupUUID != uuid.Nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot toggle in dorm for a user who has been pulled or is pulling another user"})
		err = errors.New("cannot toggle in dorm for a user who has been pulled or is pulling another user")
		return
	}

	newPullPriority := currentRoomInfo.PullPriority

	if !currentRoomInfo.PullPriority.Inherited.Valid { // this means that the in dorm status is not disabled
		newPullPriority.Inherited.Valid = true
		newPullPriority.Inherited.DrawNumber = currentRoomInfo.PullPriority.DrawNumber
		newPullPriority.Inherited.Year = currentRoomInfo.PullPriority.Year
		newPullPriority.Inherited.HasInDorm = !currentRoomInfo.PullPriority.HasInDorm
	} else { // this means that in dorm status is disabled and we should restore it
		newPullPriority.Inherited.Valid = false
		newPullPriority.Inherited.DrawNumber = 0
		newPullPriority.Inherited.Year = 0
		newPullPriority.Inherited.HasInDorm = false
	}

	newPullPriorityJSON, err := json.Marshal(newPullPriority)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal new pull priority"})
		return
	}

	_, err = tx.Exec("UPDATE rooms SET pull_priority = $1 WHERE room_uuid = $2", newPullPriorityJSON, roomUUIDParam)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update pull_priority in rooms table"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully toggled in dorm for room " + roomUUIDParam})
}

func UpdateRoomOccupants(c *gin.Context) {
	var request models.OccupantUpdateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		log.Printf("JSON unmarshal error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var err error
	switch request.PullType {
	case 1: // self pull
		err = SelfPull(c, request)
	case 2: // normal pull
		err = NormalPull(c, request)
	case 3: // lock pull
		err = LockPull(c, request)
	case 4: // alternative pull
		err = AlternativePull(c, request)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pull type"})
		return
	}

	if err != nil {
		log.Println(err)
		// c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

func SelfPull(c *gin.Context, request models.OccupantUpdateRequest) error {
	// the room uuid is in the url
	roomUUIDParam := c.Param("roomuuid")
	_, err := uuid.Parse(roomUUIDParam) // Validate UUID format early
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid room UUID format"})
		return errors.New("invalid room UUID format")
	}

	// --- Transactional Logging: Get Previous State ---
	previousRoomState, err := getRoomStateRaw(roomUUIDParam)
	if err != nil {
		log.Printf("Error fetching previous room state for SELF_PULL %s: %v", roomUUIDParam, err)
		return errors.New("failed to retrieve current room state")
	}
	if previousRoomState == nil {
		return errors.New("room not found")
	}

	userFullName, exists := c.Get("user_full_name")
	if !exists {
		log.Print("Error: user_full_name not found in context")
		userFullName = "unknown user"
	}

	userEmail, exists := c.Get("email")
	if !exists {
		log.Print("Error: email not found in context")
		userEmail = "unknown user email"
	}

	log.Println(userFullName.(string) + " is attempting a self pull for room " + roomUUIDParam)

	proposedOccupants := request.ProposedOccupants

	// verify that the proposed occupants are unique
	proposedOccupantsMap := make(map[int]bool)
	for _, occupant := range proposedOccupants {
		if proposedOccupantsMap[occupant] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Duplicate user was specified in the occupants list"})
			return errors.New("duplicate user was specified in the occupants list")
		}
		proposedOccupantsMap[occupant] = true
	}

	// Create notification queue
	notificationQueue := models.NewBumpNotificationQueue()

	// Start a transaction
	tx, err := database.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return err
	}

	// Defer rollback/commit and logging logic
	var commitErr error
	var genderUpdateErr error // To store non-fatal gender update errors
	defer func() {
		// Handle panic first
		if r := recover(); r != nil {
			tx.Rollback()
			log.Printf("PANIC during SELF_PULL for %s by %s: %v", roomUUIDParam, userEmail, r)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error due to panic"})
			return
		}

		// Handle errors that occurred *before* commit attempt
		if err != nil {
			log.Printf("Rolling back transaction for SELF_PULL %s by %s due to error: %v", roomUUIDParam, userEmail, err)
			tx.Rollback()
			// Don't log or send notifications if core logic failed
			// Response should have been sent where the error occurred
			return
		}

		// Attempt to commit
		commitErr = tx.Commit()
		if commitErr != nil {
			log.Printf("Failed to commit transaction for SELF_PULL %s by %s: %v", roomUUIDParam, userEmail, commitErr)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database transaction commit error"})
			return
		}

		// --- COMMIT SUCCEEDED ---
		log.Printf("Successfully committed SELF_PULL for room %s by %s", roomUUIDParam, userEmail)

		// Send Bump Notifications (only after successful commit)
		for _, notification := range notificationQueue.Notifications {
			log.Printf("Queueing bump notification for user %d from room %s (%s)", notification.UserID, notification.RoomID, notification.DormName)
			// Call the actual send function (make sure it's non-blocking or handled async)
			go SendBumpNotification(notification.UserID, notification.RoomID, notification.DormName)
		}

		// --- Transactional Logging: Get New State & Log ---
		newRoomState, fetchErr := getRoomStateRaw(roomUUIDParam)
		if fetchErr != nil {
			log.Printf("Error fetching new room state for SELF_PULL %s after commit: %v", roomUUIDParam, fetchErr)
			// Log the operation anyway, new state might be nil/incomplete in the log record
		}

		// Prepare details for logging
		bumpedOccupantIDs := make([]int, 0, len(notificationQueue.Notifications))
		for _, n := range notificationQueue.Notifications {
			bumpedOccupantIDs = append(bumpedOccupantIDs, n.UserID)
		}

		logDetails := map[string]interface{}{
			"proposed_occupants": proposedOccupants,
			// Use previousRoomState for accurate pre-change data
			"previous_occupants":   previousRoomState.Occupants,
			"previous_sgroup_uuid": previousRoomState.SGroupUUID, // Could be nil/invalid
			"bumped_occupant_ids":  bumpedOccupantIDs,
			// Add status of secondary operations like gender update
			"gender_update_error": nil, // Default to nil
		}
		if genderUpdateErr != nil {
			logDetails["gender_update_error"] = genderUpdateErr.Error()
		}

		// Call LogOperation
		loggingErr := logging.LogOperation(
			c,                     // Pass the Gin context
			"SELF_PULL",           // Operation Type
			models.EntityTypeRoom, // Entity Type
			roomUUIDParam,         // Entity ID
			previousRoomState,     // State Before (fetched before tx)
			newRoomState,          // State After (fetched after successful commit)
			logDetails,            // Additional Details
		)
		if loggingErr != nil {
			// Log the logging error, but the main operation succeeded.
			log.Printf("WARNING: Failed to log SELF_PULL operation for %s: %v", roomUUIDParam, loggingErr)
		}

		// Send success response *after* logging attempt
		c.JSON(http.StatusOK, gin.H{"message": "Successfully updated occupants"})

	}() // End of defer func

	var currentRoomInfo models.RoomRaw
	err = tx.QueryRow("SELECT room_uuid, dorm, dorm_name, room_id, suite_uuid, max_occupancy, current_occupancy, occupants, pull_priority, has_frosh FROM rooms WHERE room_uuid = $1", roomUUIDParam).Scan(
		&currentRoomInfo.RoomUUID,
		&currentRoomInfo.Dorm,
		&currentRoomInfo.DormName,
		&currentRoomInfo.RoomID,
		&currentRoomInfo.SuiteUUID,
		&currentRoomInfo.MaxOccupancy,
		&currentRoomInfo.CurrentOccupancy,
		&currentRoomInfo.Occupants,
		&currentRoomInfo.PullPriority,
		&currentRoomInfo.HasFrosh,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query room info from rooms table"})
		return err
	}

	// log room uuid
	log.Println(currentRoomInfo.RoomUUID)

	// make sure the room does not have frosh
	if currentRoomInfo.HasFrosh {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot pull into a room with frosh"})
		err = errors.New("room has frosh")
		return err
	}

	if currentRoomInfo.PullPriority.IsPreplaced {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot pull into a preplaced room"})
		err = errors.New("room is preplaced")
		return err
	}

	if len(proposedOccupants) == 0 {
		email := c.MustGet("email").(string)
		err = clearRoom(currentRoomInfo.RoomUUID, tx, notificationQueue, email)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear room"})
			tx.Rollback()
		}
		return err
	}

	// check that the proposed occupants are not more than the max occupancy
	if len(proposedOccupants) > currentRoomInfo.MaxOccupancy {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Proposed occupants exceeds max occupancy"})
		err = errors.New("proposed occupants exceeds max occupancy")
		return err
	}

	// ensure that room is full before self pulling
	if len(proposedOccupants) < currentRoomInfo.MaxOccupancy {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Room is not full"})
		err = errors.New("room is not full")
		return err
	}

	var occupantsAlreadyInRoom models.IntArray
	rows, err := tx.Query("SELECT id FROM users WHERE id = ANY($1) AND room_uuid IS NOT NULL", pq.Array(proposedOccupants))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query room_uuid from users table"})
		return err
	}

	for rows.Next() {
		var occupant int
		if err := rows.Scan(&occupant); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan room_uuid from users table"})
			return err
		}

		occupantsAlreadyInRoom = append(occupantsAlreadyInRoom, occupant)
	}

	if len(occupantsAlreadyInRoom) > 0 {
		err = errors.New("one or more of the proposed occupants is already in a room")
		c.JSON(http.StatusBadRequest, gin.H{"error": "One or more of the proposed occupants is already in a room", "occupants": occupantsAlreadyInRoom})
		return err
	}

	var proposedPullPriority models.PullPriority
	log.Println(request.PullType)

	log.Println("Self pull")
	var occupantsInfo []models.UserRaw
	rows, err = tx.Query("SELECT id, draw_number, year, in_dorm, participated, preplaced FROM users WHERE id = ANY($1)", pq.Array(proposedOccupants))
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed on users for pull priority"})
		tx.Rollback()
		return err
	}
	for rows.Next() {
		var u models.UserRaw
		if err := rows.Scan(&u.Id, &u.DrawNumber, &u.Year, &u.InDorm, &u.Participated, &u.Preplaced); err != nil {
			// Handle scan error
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database scan failed on users for pull priority"})
			tx.Rollback()
			return err
		}

		// if any of the proposed occupants are preplaced, return an error
		if u.Preplaced {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot pull with a preplaced user"})
			err = errors.New("cannot pull with a preplaced user")
			tx.Rollback()
			return err
		}

		occupantsInfo = append(occupantsInfo, u)
	}

	// for all users who currently have not participated, set their participated field to true and partitipation time to now
	_, err = tx.Exec("UPDATE users SET participated = true, participation_time = NOW() WHERE id = ANY($1) AND participated = false", pq.Array(proposedOccupants))
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update participated field in users table"})
		tx.Rollback()
		return err
	}

	// loop through all occupants and check if they have in dorm
	// if at least one does not have in dorm, change each of the pull priorities of each occupant to not have in dorm
	for _, occupant := range occupantsInfo {
		if occupant.InDorm != currentRoomInfo.Dorm {
			log.Println("Forfeited in dorm to pull non-in dorm user")
			for i := range occupantsInfo {
				occupantsInfo[i].InDorm = 0
			}
			break
		}
	}

	sortedOccupants := sortUsersByPriority(occupantsInfo, currentRoomInfo.Dorm)

	proposedPullPriority = generateUserPriority(sortedOccupants[0], currentRoomInfo.Dorm)

	proposedPullPriority.Valid = true
	proposedPullPriority.PullType = 1

	if currentRoomInfo.PullPriority.PullType == 3 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot bump a lock pulled room"})
		err = errors.New("cannot bump a lock pulled room")
		tx.Rollback()
		return err
	}

	if !comparePullPriority(proposedPullPriority, currentRoomInfo.PullPriority) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Proposed occupants do not have higher priority than current occupants"})
		err = errors.New("proposed occupants do not have higher priority than current occupants")
		tx.Rollback()
		return err
	}

	// disband the suite group if there is one
	if currentRoomInfo.SGroupUUID != uuid.Nil {
		_, err := disbandSuiteGroup(currentRoomInfo.SGroupUUID, tx)
		if err != nil {
			// use err in the response
			c.JSON(http.StatusInternalServerError, err.Error())
			return err
		}
	}

	if currentRoomInfo.CurrentOccupancy > 0 {
		// remove the current occupants from the room
		email := c.MustGet("email").(string)
		err = clearRoom(currentRoomInfo.RoomUUID, tx, notificationQueue, email)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove the current occupants of the room"})
			tx.Rollback()
		}
	}

	// update the occupants in the database and the current_occupancy
	_, err = tx.Exec("UPDATE rooms SET occupants = $1, current_occupancy = $2 WHERE room_uuid = $3", pq.Array(proposedOccupants), len(proposedOccupants), roomUUIDParam)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return err
	}

	// for each occupant, update the room_uuid field in the users table
	for _, proposedOccupant := range proposedOccupants {
		_, err = tx.Exec("UPDATE users SET room_uuid = $1 WHERE id = $2", roomUUIDParam, proposedOccupant)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update room_uuid in users table"})
			return err
		}
	}

	// update the pull_priority field in the rooms table
	proposedPullPriorityJSON, err := json.Marshal(proposedPullPriority)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal proposed pull priority"})
		return err
	}

	_, err = tx.Exec("UPDATE rooms SET pull_priority = $1 WHERE room_uuid = $2", proposedPullPriorityJSON, roomUUIDParam)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update pull_priority in rooms table"})
		return err
	}

	// Update gender preferences for the suite
	err = UpdateSuiteGenderPreferencesBySuiteUUID(tx, currentRoomInfo.SuiteUUID)
	if err != nil {
		// If there's a gender preference conflict, fail the transaction
		if strings.Contains(err.Error(), "no valid intersection of gender preferences") {
			c.JSON(http.StatusConflict, gin.H{"error": "Cannot pull users with incompatible gender preferences. The users must have at least one gender preference in common."})
			return err
		}

		// For other errors, log a warning but continue
		log.Printf("Warning: Failed to update gender preferences for suite %s: %v", currentRoomInfo.SuiteUUID, err)
	}

	return nil
}

func NormalPull(c *gin.Context, request models.OccupantUpdateRequest) error {
	// the room uuid is in the url
	roomUUIDParam := c.Param("roomuuid")

	userFullName, exists := c.Get("user_full_name")
	if !exists {
		log.Print("Error: user_full_name not found in context")
		userFullName = "unknown user"
	}

	userEmail, exists := c.Get("email")
	if !exists {
		log.Print("Error: email not found in context")
		userEmail = "unknown user email"
	}

	proposedOccupants := request.ProposedOccupants

	// Convert proposedOccupants from []int to []string
	proposedOccupantStrings := make([]string, len(proposedOccupants))
	for i, occupant := range proposedOccupants {
		proposedOccupantStrings[i] = strconv.Itoa(occupant)
	}

	log.Println(userFullName.(string) + " is attempting a normal pull for room " + roomUUIDParam + " with occupants " + strings.Join(proposedOccupantStrings, ", "))

	// verify that the proposed occupants are unique
	proposedOccupantsMap := make(map[int]bool)
	for _, occupant := range proposedOccupants {
		if proposedOccupantsMap[occupant] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Duplicate user was specified in the occupants list"})
			return errors.New("duplicate user was specified in the occupants list")
		}
		proposedOccupantsMap[occupant] = true
	}

	// --- Transactional Logging: Get Previous State ---
	previousRoomState, err := getRoomStateRaw(roomUUIDParam)
	if err != nil {
		log.Printf("Error fetching previous room state for NORMAL_PULL %s: %v", roomUUIDParam, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve current room state"})
		return err
	}
	if previousRoomState == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Target room not found"})
		return sql.ErrNoRows
	}
	// Fetch previous state of pull leader room IF its state might change significantly (e.g., sgroup_uuid, inherited priority)
	previousPullLeaderRoomState, err := getRoomStateRaw(request.PullLeaderRoom.String())
	if err != nil {
		log.Printf("Error fetching previous pull leader room state for NORMAL_PULL %s: %v", request.PullLeaderRoom.String(), err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve pull leader room state"})
		return err
	}
	if previousPullLeaderRoomState == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pull leader room not found"})
		return sql.ErrNoRows
	}

	// Create notification queue
	notificationQueue := models.NewBumpNotificationQueue()

	// Start a transaction
	tx, err := database.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return err
	}

	var commitErr error
	var genderUpdateErr error       // To store non-fatal gender update errors
	var createdSGroupUUID uuid.UUID // To store newly created group ID for logging
	var joinedSGroupUUID uuid.UUID  // To store the ID of the group joined

	// Ensure the transaction is either committed or rolled back
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Printf("PANIC during NORMAL_PULL for %s by %s: %v", roomUUIDParam, userEmail, r)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error due to panic"})
			return
		}
		if err != nil {
			log.Printf("Rolling back transaction for NORMAL_PULL %s by %s due to error: %v", roomUUIDParam, userEmail, err)
			tx.Rollback()
			return // Error response should have been sent
		}
		commitErr = tx.Commit()
		if commitErr != nil {
			log.Printf("Failed to commit transaction for NORMAL_PULL %s by %s: %v", roomUUIDParam, userEmail, commitErr)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database transaction commit error"})
			return
		}

		// --- COMMIT SUCCEEDED ---
		log.Printf("Successfully committed NORMAL_PULL for room %s by %s", roomUUIDParam, userEmail)

		// Send Bump Notifications
		for _, notification := range notificationQueue.Notifications {
			go SendBumpNotification(notification.UserID, notification.RoomID, notification.DormName)
		}

		// --- Transactional Logging: Get New State & Log ---
		newRoomState, fetchErr := getRoomStateRaw(roomUUIDParam)
		if fetchErr != nil {
			log.Printf("Error fetching new room state for NORMAL_PULL %s after commit: %v", roomUUIDParam, fetchErr)
		}

		// Fetch new state of pull leader room
		newPullLeaderRoomState, fetchLeaderErr := getRoomStateRaw(request.PullLeaderRoom.String())
		if fetchLeaderErr != nil {
			log.Printf("Error fetching new pull leader room state for NORMAL_PULL %s after commit: %v", request.PullLeaderRoom.String(), fetchLeaderErr)
		}

		// Prepare details
		bumpedOccupantIDs := make([]int, 0, len(notificationQueue.Notifications))
		for _, n := range notificationQueue.Notifications {
			bumpedOccupantIDs = append(bumpedOccupantIDs, n.UserID)
		}

		logDetails := map[string]interface{}{
			"proposed_occupants":          proposedOccupants,
			"pull_leader_room":            request.PullLeaderRoom,
			"previous_occupants":          previousRoomState.Occupants,
			"previous_sgroup_uuid":        previousRoomState.SGroupUUID,           // Target room's original group
			"previous_leader_sgroup_uuid": previousPullLeaderRoomState.SGroupUUID, // Leader's original group
			"bumped_occupant_ids":         bumpedOccupantIDs,
			"created_sgroup_uuid":         nil, // Default
			"joined_sgroup_uuid":          nil, // Default
			"gender_update_error":         nil, // Default
		}
		if createdSGroupUUID != uuid.Nil {
			logDetails["created_sgroup_uuid"] = createdSGroupUUID
		}
		if joinedSGroupUUID != uuid.Nil {
			logDetails["joined_sgroup_uuid"] = joinedSGroupUUID
		}
		if genderUpdateErr != nil {
			logDetails["gender_update_error"] = genderUpdateErr.Error()
		}

		// Log primary operation on the target room
		loggingErr := logging.LogOperation(
			c,                     // Pass the Gin context
			"NORMAL_PULL",         // Operation Type
			models.EntityTypeRoom, // Entity Type
			roomUUIDParam,         // Entity ID
			previousRoomState,     // State Before target room
			newRoomState,          // State After target room
			logDetails,            // Additional Details
		)
		if loggingErr != nil {
			log.Printf("WARNING: Failed to log NORMAL_PULL operation for target room %s: %v", roomUUIDParam, loggingErr)
		}

		// Optionally log the change to the pull leader room if significant state changed
		// Determine if leader state changed enough to warrant a separate log entry
		leaderStateChanged := !compareRoomStatesForLogging(previousPullLeaderRoomState, newPullLeaderRoomState) // Implement compareRoomStatesForLogging

		if leaderStateChanged {
			leaderLogDetails := map[string]interface{}{
				"pull_action_target_room": roomUUIDParam,
				"related_occupants":       proposedOccupants,
				"reason":                  "Participant in NORMAL_PULL",
				"previous_sgroup_uuid":    previousPullLeaderRoomState.SGroupUUID,
				"new_sgroup_uuid":         newPullLeaderRoomState.SGroupUUID, // Assuming sgroup is the main change
			}
			leaderLoggingErr := logging.LogOperation(
				c,
				"UPDATE_ROOM_STATE", // Or a more specific type like "UPDATE_PULL_LEADER_STATE"
				models.EntityTypeRoom,
				request.PullLeaderRoom.String(), // Leader Room ID
				previousPullLeaderRoomState,
				newPullLeaderRoomState,
				leaderLogDetails,
			)
			if leaderLoggingErr != nil {
				log.Printf("WARNING: Failed to log state change for pull leader room %s during NORMAL_PULL: %v", request.PullLeaderRoom.String(), leaderLoggingErr)
			}
		}

		// Send success response *after* logging attempt
		c.JSON(http.StatusOK, gin.H{"message": "Successfully updated occupants"})

	}() // End of defer func

	drinkwardTripleInDormPullExceptions := []string{
		"123C",
		"124C",
		"221C",
		"222C",
		"223C",
		"224C",
		"321C",
		"322C",
		"323C",
		"324C",
	}

	var currentRoomInfo models.RoomRaw
	err = tx.QueryRow("SELECT room_uuid, dorm, dorm_name, room_id, suite_uuid, max_occupancy, current_occupancy, occupants, pull_priority, has_frosh FROM rooms WHERE room_uuid = $1", roomUUIDParam).Scan(
		&currentRoomInfo.RoomUUID,
		&currentRoomInfo.Dorm,
		&currentRoomInfo.DormName,
		&currentRoomInfo.RoomID,
		&currentRoomInfo.SuiteUUID,
		&currentRoomInfo.MaxOccupancy,
		&currentRoomInfo.CurrentOccupancy,
		&currentRoomInfo.Occupants,
		&currentRoomInfo.PullPriority,
		&currentRoomInfo.HasFrosh,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query room info from rooms table"})
		return err
	}

	// log room uuid
	log.Println(currentRoomInfo.RoomUUID)

	// make sure the room does not have frosh
	if currentRoomInfo.HasFrosh {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Room has frosh"})
		err = errors.New("room has frosh")
		return err
	}

	if currentRoomInfo.PullPriority.IsPreplaced {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot pull into a preplaced room"})
		err = errors.New("room is preplaced")
		return err
	}

	// check that the proposed occupants are not more than the max occupancy
	if len(proposedOccupants) > currentRoomInfo.MaxOccupancy {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Proposed occupants exceeds max occupancy"})
		err = errors.New("proposed occupants exceeds max occupancy")
		return err
	}

	var occupantsAlreadyInRoom models.IntArray
	rows, err := tx.Query("SELECT id FROM users WHERE id = ANY($1) AND room_uuid IS NOT NULL", pq.Array(proposedOccupants))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query room_uuid from users table"})
		return err
	}

	for rows.Next() {
		var occupant int
		if err := rows.Scan(&occupant); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan room_uuid from users table"})
			return err
		}

		occupantsAlreadyInRoom = append(occupantsAlreadyInRoom, occupant)
	}

	if len(occupantsAlreadyInRoom) > 0 {
		err = errors.New("one or more of the proposed occupants is already in a room")
		c.JSON(http.StatusBadRequest, gin.H{"error": "One or more of the proposed occupants is already in a room", "occupants": occupantsAlreadyInRoom})
		return err
	}

	var proposedPullPriority models.PullPriority
	var pullLeaderPriority models.PullPriority
	var pullLeaderSuiteGroupUUID uuid.UUID

	if len(proposedOccupants) == 0 {
		// error because normal pull requires at least one occupant
		c.JSON(http.StatusBadRequest, gin.H{"error": "Normal pull requires at least one occupant"})
		err = errors.New("normal pull requires at least one occupant")
		tx.Rollback()
		return err
	}

	if currentRoomInfo.RoomUUID == request.PullLeaderRoom {
		// error because the pull leader is already in the room
		c.JSON(http.StatusBadRequest, gin.H{"error": "Pull leader is already in the room"})
		err = errors.New("pull leader is already in the room")
		tx.Rollback()
		return err
	}

	if currentRoomInfo.MaxOccupancy > 1 {
		// error because normal pull is not allowed for rooms with max occupancy > 1
		c.JSON(http.StatusBadRequest, gin.H{"error": "You may only initiate a normal pull for singles"})
		err = errors.New("normal pull is not allowed for rooms with max occupancy > 1")
		tx.Rollback()
		return err
	}
	pullLeaderRoomUUID := request.PullLeaderRoom

	var occupantsInfo []models.UserRaw
	rows, err = tx.Query("SELECT id, draw_number, year, in_dorm, participated, preplaced FROM users WHERE id = ANY($1)", pq.Array(proposedOccupants))
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed on users for pull priority"})
		tx.Rollback()
		return err
	}

	for rows.Next() {
		var u models.UserRaw
		if err := rows.Scan(&u.Id, &u.DrawNumber, &u.Year, &u.InDorm, &u.Participated, &u.Preplaced); err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database scan failed on users for pull priority"})
			tx.Rollback()
			return err
		}

		// if any of the proposed occupants are preplaced, return an error
		if u.Preplaced {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot pull with a preplaced user"})
			err = errors.New("cannot pull with a preplaced user")
			tx.Rollback()
			return err
		}

		occupantsInfo = append(occupantsInfo, u)
	}

	// check if this is a case of an in dorm pull leader pulling three people into a drinkward triple, this will be used in the future as an exception to the rule
	// that in dorm pulls can only pull other in dorm users
	isDrinkwardTripleException := false
	pullLeaderHasInDorm := false
	if pullLeaderPriority.Inherited.Valid {
		pullLeaderHasInDorm = pullLeaderPriority.Inherited.HasInDorm
	} else {
		pullLeaderHasInDorm = pullLeaderPriority.HasInDorm
	}
	if pullLeaderHasInDorm && currentRoomInfo.Dorm == 8 { // 8 is the dorm code for drinkward
		for _, roomID := range drinkwardTripleInDormPullExceptions {
			if currentRoomInfo.RoomID == roomID {
				isDrinkwardTripleException = true
			}
		}
	}

	// for all users who currently have not participated, set their participated field to true and partitipation time to now
	_, err = tx.Exec("UPDATE users SET participated = true, participation_time = NOW() WHERE id = ANY($1) AND participated = false", pq.Array(proposedOccupants))
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update participated field in users table"})
		tx.Rollback()
		return err
	}

	var suiteUUID = currentRoomInfo.SuiteUUID
	var leaderSuiteUUID uuid.UUID
	var pullLeaderCurrentOccupancy int

	// get the pull leader's priority
	err = tx.QueryRow("SELECT pull_priority, sgroup_uuid, suite_uuid, current_occupancy FROM rooms WHERE room_uuid = $1", pullLeaderRoomUUID).Scan(&pullLeaderPriority, &pullLeaderSuiteGroupUUID, &leaderSuiteUUID, &pullLeaderCurrentOccupancy)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query pull leader's priority from rooms table"})
		tx.Rollback()
		return err
	}

	if leaderSuiteUUID != suiteUUID {
		// error because the pull leader is not in the same suite
		c.JSON(http.StatusBadRequest, gin.H{"error": "Pull leader is not in the same suite"})
		tx.Rollback()
		return err
	}

	if pullLeaderCurrentOccupancy != 1 {
		// error because the pull leader is not in a single
		c.JSON(http.StatusBadRequest, gin.H{"error": "You can only initiate a normal pull with a single"})
		tx.Rollback()
		return err
	}

	sortedOccupants := sortUsersByPriority(occupantsInfo, currentRoomInfo.Dorm)

	var pullLeaderEffectiveInDorm bool

	// accounts for in dorm forfeit
	if pullLeaderPriority.Inherited.Valid {
		pullLeaderEffectiveInDorm = pullLeaderPriority.Inherited.HasInDorm
	} else {
		pullLeaderEffectiveInDorm = pullLeaderPriority.HasInDorm
	}

	// if the pull leader has indorm and the proposed occupants do not, it is invalid
	// however, if the drinkward triple exception is true, then the pull leader can pull three people into a drinkward triple
	if !isDrinkwardTripleException {
		for _, occupant := range sortedOccupants {
			if pullLeaderEffectiveInDorm && !(generateUserPriority(occupant, currentRoomInfo.Dorm).HasInDorm) {
				log.Println("Pull leader has in dorm and proposed occupants do not")
				err = errors.New("pull leader has in dorm and proposed occupants do not")
				c.JSON(http.StatusBadRequest, gin.H{"error": "Pull leader has in dorm and proposed occupants do not"})
				return err
			}
		}
	} else {
		log.Println("Special case: Drinkward triple exception where pull leader has in dorm and proposed occupants do not")
	}

	proposedPullPriority = generateUserPriority(sortedOccupants[0], currentRoomInfo.Dorm)
	proposedPullPriority.Valid = true
	proposedPullPriority.PullType = 2

	log.Println(proposedPullPriority)
	log.Println(pullLeaderPriority)

	if !comparePullPriority(pullLeaderPriority, proposedPullPriority) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Pull leader does not have higher priority than proposed occupants"})
		tx.Rollback()
		return err
	}

	proposedPullPriority.Inherited.Valid = true
	if pullLeaderPriority.Inherited.Valid {
		proposedPullPriority.Inherited.DrawNumber = pullLeaderPriority.Inherited.DrawNumber
		proposedPullPriority.Inherited.HasInDorm = pullLeaderPriority.Inherited.HasInDorm
		proposedPullPriority.Inherited.Year = pullLeaderPriority.Inherited.Year
	} else {
		proposedPullPriority.Inherited.DrawNumber = pullLeaderPriority.DrawNumber
		proposedPullPriority.Inherited.HasInDorm = pullLeaderPriority.HasInDorm
		proposedPullPriority.Inherited.Year = pullLeaderPriority.Year
	}

	if currentRoomInfo.PullPriority.PullType == 3 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot bump a lock pulled room"})
		err = errors.New("cannot bump a lock pulled room")
		tx.Rollback()
		return err
	}

	if !comparePullPriority(proposedPullPriority, currentRoomInfo.PullPriority) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Proposed occupants do not have higher priority than current occupants"})
		err = errors.New("proposed occupants do not have higher priority than current occupants")
		tx.Rollback()
		return err
	}

	log.Println(proposedOccupants)

	// disband the suite group if there is one
	if currentRoomInfo.SGroupUUID != uuid.Nil {
		_, err := disbandSuiteGroup(currentRoomInfo.SGroupUUID, tx)
		if err != nil {
			// use err in the response
			c.JSON(http.StatusInternalServerError, err.Error())
			return err
		}
	}

	if currentRoomInfo.CurrentOccupancy > 0 {
		// use clearRoom function to remove the current occupants from the room
		email := c.MustGet("email").(string)
		err = clearRoom(currentRoomInfo.RoomUUID, tx, notificationQueue, email)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove the current occupants of the room"})
			tx.Rollback()
			return err
		}
	}

	// update the occupants in the database and the current_occupancy
	_, err = tx.Exec("UPDATE rooms SET occupants = $1, current_occupancy = $2 WHERE room_uuid = $3", pq.Array(proposedOccupants), len(proposedOccupants), roomUUIDParam)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return err
	}

	// for each occupant, update the room_uuid field in the users table
	for _, proposedOccupant := range proposedOccupants {
		_, err = tx.Exec("UPDATE users SET room_uuid = $1 WHERE id = $2", roomUUIDParam, proposedOccupant)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update room_uuid in users table"})
			return err
		}
	}

	// update the pull_priority field in the rooms table
	proposedPullPriorityJSON, err := json.Marshal(proposedPullPriority)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal proposed pull priority"})
		return err
	}

	_, err = tx.Exec("UPDATE rooms SET pull_priority = $1 WHERE room_uuid = $2", proposedPullPriorityJSON, roomUUIDParam)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update pull_priority in rooms table"})
		return err
	}

	if pullLeaderSuiteGroupUUID == uuid.Nil {
		log.Println("Pull leader is not in a suite group")
		// create new suite group with the pull leader's priority
		pullLeaderPriorityJSON, err := json.Marshal(pullLeaderPriority)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal pull leader's pull priority"})
			return err
		}

		var suiteGroupUUID uuid.UUID = uuid.New()
		_, err = tx.Exec("INSERT INTO suitegroups (sgroup_uuid, sgroup_size, sgroup_name, sgroup_suite, pull_priority, rooms, disbanded) VALUES ($1, $2, $3, $4, $5, $6, $7)",
			suiteGroupUUID,
			2,
			"Suite Group",
			currentRoomInfo.SuiteUUID,
			pullLeaderPriorityJSON,
			pq.Array(models.UUIDArray{currentRoomInfo.RoomUUID, request.PullLeaderRoom}),
			false,
		)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert new suite group into suitegroups table"})
			return err
		}

		// update the sgroup_uuid field in the rooms table for both rooms
		_, err = tx.Exec("UPDATE rooms SET sgroup_uuid = $1 WHERE room_uuid = $2 OR room_uuid = $3", suiteGroupUUID, roomUUIDParam, request.PullLeaderRoom)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update suite_uuid in rooms table"})
			return err
		}

		// update the sgroup_uuid field in the users table for all occupants of both rooms
		_, err = tx.Exec("UPDATE users SET sgroup_uuid = $1 WHERE room_uuid = $2 OR room_uuid = $3", suiteGroupUUID, roomUUIDParam, request.PullLeaderRoom)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update sgroup_uuid in users table"})
			return err
		}

		// Update gender preferences for the suite
		err = UpdateSuiteGenderPreferencesBySuiteUUID(tx, currentRoomInfo.SuiteUUID)
		if err != nil {
			// If there's a gender preference conflict, fail the transaction
			if strings.Contains(err.Error(), "no valid intersection of gender preferences") {
				c.JSON(http.StatusConflict, gin.H{"error": "Cannot pull users with incompatible gender preferences. The users must have at least one gender preference in common."})
				return err
			}

			// For other errors, log a warning but continue
			log.Printf("Warning: Failed to update gender preferences for suite %s: %v", currentRoomInfo.SuiteUUID, err)
		}
	} else {
		log.Println("Pull leader is in a suite group")

		// first check that the room in in south and that the number of rooms in the suite is 3
		var suiteInfo models.SuiteRaw
		err = tx.QueryRow("SELECT suite_uuid, dorm, dorm_name, floor, room_count, rooms, alternative_pull FROM suites WHERE suite_uuid = $1", currentRoomInfo.SuiteUUID).Scan(
			&suiteInfo.SuiteUUID,
			&suiteInfo.Dorm,
			&suiteInfo.DormName,
			&suiteInfo.Floor,
			&suiteInfo.RoomCount,
			&suiteInfo.Rooms,
			&suiteInfo.AlternativePull,
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query suite info from suites table"})
			return err
		}

		if suiteInfo.Dorm != 3 {
			if suiteInfo.RoomCount != 3 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "You can only pull two suitemates in South in a suite with three rooms"})
				err = errors.New("you can only pull two suitemates in South in a suite with three rooms")
				tx.Rollback()
				return err
			}

			c.JSON(http.StatusBadRequest, gin.H{"error": "You can only pull two suitemates in South"})
			err = errors.New("you can only pull two suitemates in South")
			tx.Rollback()
			return err
		}

		// check if the pull leader is the leader of the suite group by checking if the suite group's pull priority is the same as the pull leader's pull priority
		var pullLeaderSuiteGroupPriority models.PullPriority
		err = tx.QueryRow("SELECT pull_priority FROM suitegroups WHERE sgroup_uuid = $1", pullLeaderSuiteGroupUUID).Scan(&pullLeaderSuiteGroupPriority)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query pull leader's suite group's pull priority from suitegroups table"})
		}

		// deep equal the pull leader's priority and the suite group's priority
		if pullLeaderSuiteGroupPriority != pullLeaderPriority {
			log.Println(pullLeaderSuiteGroupPriority)
			log.Println(pullLeaderPriority)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Pull leader is not the leader of the suite group"})
			tx.Rollback()
			return err
		}

		// add the room to the suite group
		_, err = tx.Exec("UPDATE suitegroups SET rooms = array_append(rooms, $1) WHERE sgroup_uuid = $2", roomUUIDParam, pullLeaderSuiteGroupUUID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update rooms in suitegroups table"})
			return err
		}

		// update the sgroup_uuid field in the rooms table
		_, err = tx.Exec("UPDATE rooms SET sgroup_uuid = $1 WHERE room_uuid = $2", pullLeaderSuiteGroupUUID, roomUUIDParam)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update suite_uuid in rooms table"})
			return err
		}

		// update the sgroup_uuid field in the users table for all occupants of the room
		_, err = tx.Exec("UPDATE users SET sgroup_uuid = $1 WHERE room_uuid = $2", pullLeaderSuiteGroupUUID, roomUUIDParam)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update sgroup_uuid in users table"})
			return err
		}

		// Update gender preferences for the suite
		err = UpdateSuiteGenderPreferencesBySuiteUUID(tx, currentRoomInfo.SuiteUUID)
		if err != nil {
			// If there's a gender preference conflict, fail the transaction
			if strings.Contains(err.Error(), "no valid intersection of gender preferences") {
				c.JSON(http.StatusConflict, gin.H{"error": "Cannot pull users with incompatible gender preferences. The users must have at least one gender preference in common."})
				return err
			}

			// For other errors, log a warning but continue
			log.Printf("Warning: Failed to update gender preferences for suite %s: %v", currentRoomInfo.SuiteUUID, err)
		}
	}

	return nil
}

func LockPull(c *gin.Context, request models.OccupantUpdateRequest) error {
	// the room uuid is in the url
	roomUUIDParam := c.Param("roomuuid")

	userFullName, exists := c.Get("user_full_name")
	if !exists {
		log.Print("Error: user_full_name not found in context")
		userFullName = "unknown user"
	}

	userEmail, exists := c.Get("email")
	if !exists {
		log.Print("Error: email not found in context")
		userEmail = "unknown user email"
	}

	log.Println(userFullName.(string) + " is attempting a lock pull for room " + roomUUIDParam)

	proposedOccupants := request.ProposedOccupants

	// verify that the proposed occupants are unique
	proposedOccupantsMap := make(map[int]bool)
	for _, occupant := range proposedOccupants {
		if proposedOccupantsMap[occupant] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Duplicate user was specified in the occupants list"})
			return errors.New("duplicate user was specified in the occupants list")
		}
		proposedOccupantsMap[occupant] = true
	}

	previousRoomState, err := getRoomStateRaw(roomUUIDParam)
	if err != nil { /* ... handle error ... */
		return err
	}
	if previousRoomState == nil { /* ... handle not found ... */
		return sql.ErrNoRows
	}
	// Suite (to check previous lock status)
	var previousSuiteState *models.SuiteRaw // Use pointer type
	// Need getSuiteStateRaw helper function
	previousSuiteState, err = getSuiteStateRaw(previousRoomState.SuiteUUID.String()) // Fetch suite using room's suite_uuid
	if err != nil {
		log.Printf("Error fetching previous suite state %s for LOCK_PULL: %v", previousRoomState.SuiteUUID.String(), err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve suite state"})
		return err
	}
	if previousSuiteState == nil {
		// This shouldn't happen if the room exists, implies data inconsistency
		log.Printf("ERROR: Suite %s not found for existing room %s during LOCK_PULL", previousRoomState.SuiteUUID.String(), roomUUIDParam)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal data inconsistency: Suite not found"})
		return errors.New("suite not found for room")
	}

	// Create notification queue
	notificationQueue := models.NewBumpNotificationQueue()

	// Start a transaction
	tx, err := database.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return err
	}

	var commitErr error
	var genderUpdateErr error

	// Ensure the transaction is either committed or rolled back
	defer func() {
		if r := recover(); r != nil { /* ... handle panic rollback ... */
			return
		}
		if err != nil { /* ... handle error rollback ... */
			return
		}
		commitErr = tx.Commit()
		if commitErr != nil { /* ... handle commit error ... */
			return
		}

		// --- COMMIT SUCCEEDED ---
		log.Printf("Successfully committed LOCK_PULL for room %s by %s", roomUUIDParam, userEmail)

		// Send Bump Notifications (if any, unlikely for lock pull)
		for _, notification := range notificationQueue.Notifications {
			go SendBumpNotification(notification.UserID, notification.RoomID, notification.DormName)
		}

		// --- Transactional Logging: Get New State & Log ---
		newRoomState, fetchErr := getRoomStateRaw(roomUUIDParam)
		if fetchErr != nil {
			log.Printf("Error fetching new room state for LOCK_PULL %s: %v", roomUUIDParam, fetchErr)
		}
		// Fetch new suite state
		newSuiteState, fetchSuiteErr := getSuiteStateRaw(previousRoomState.SuiteUUID.String()) // Use stored SuiteUUID
		if fetchSuiteErr != nil {
			log.Printf("Error fetching new suite state %s for LOCK_PULL: %v", previousRoomState.SuiteUUID.String(), fetchSuiteErr)
		}

		// Prepare details
		bumpedOccupantIDs := make([]int, 0, len(notificationQueue.Notifications))
		for _, n := range notificationQueue.Notifications {
			bumpedOccupantIDs = append(bumpedOccupantIDs, n.UserID)
		}

		logDetails := map[string]interface{}{
			"proposed_occupants":        proposedOccupants,
			"previous_occupants":        previousRoomState.Occupants,
			"previous_sgroup_uuid":      previousRoomState.SGroupUUID,
			"previous_suite_lock_state": previousSuiteState.LockPulledRoom, // From pre-fetched suite state
			"bumped_occupant_ids":       bumpedOccupantIDs,
			"gender_update_error":       nil,
		}
		if genderUpdateErr != nil {
			logDetails["gender_update_error"] = genderUpdateErr.Error()
		}

		// Log primary operation on the target room
		loggingErr := logging.LogOperation(
			c, "LOCK_PULL", models.EntityTypeRoom, roomUUIDParam,
			previousRoomState, // Room state before
			newRoomState,      // Room state after
			logDetails,
		)
		if loggingErr != nil {
			log.Printf("WARNING: Failed to log LOCK_PULL room operation %s: %v", roomUUIDParam, loggingErr)
		}

		// Log the change to the suite entity
		// Check if suite state actually changed (LockPulledRoom field)
		suiteStateChanged := previousSuiteState.LockPulledRoom != newSuiteState.LockPulledRoom // Assuming LockPulledRoom is comparable

		if suiteStateChanged {
			suiteLogDetails := map[string]interface{}{
				"reason":              "Suite lock status updated by LOCK_PULL",
				"locking_room_uuid":   roomUUIDParam,
				"locking_occupants":   proposedOccupants,
				"previous_lock_state": previousSuiteState.LockPulledRoom,
			}
			suiteLoggingErr := logging.LogOperation(
				c, "UPDATE_SUITE_LOCK", // More specific type
				models.EntityTypeSuite,                // Entity Type is SUITE
				previousSuiteState.SuiteUUID.String(), // Suite UUID is the entity ID
				previousSuiteState,                    // Suite state before
				newSuiteState,                         // Suite state after
				suiteLogDetails,
			)
			if suiteLoggingErr != nil {
				log.Printf("WARNING: Failed to log LOCK_PULL suite operation %s: %v", previousSuiteState.SuiteUUID.String(), suiteLoggingErr)
			}
		}

		// Send success response *after* logging attempt
		c.JSON(http.StatusOK, gin.H{"message": "Successfully updated occupants"})

	}() // End of defer func

	var currentRoomInfo models.RoomRaw
	err = tx.QueryRow("SELECT room_uuid, dorm, dorm_name, room_id, suite_uuid, max_occupancy, current_occupancy, occupants, pull_priority, has_frosh FROM rooms WHERE room_uuid = $1", roomUUIDParam).Scan(
		&currentRoomInfo.RoomUUID,
		&currentRoomInfo.Dorm,
		&currentRoomInfo.DormName,
		&currentRoomInfo.RoomID,
		&currentRoomInfo.SuiteUUID,
		&currentRoomInfo.MaxOccupancy,
		&currentRoomInfo.CurrentOccupancy,
		&currentRoomInfo.Occupants,
		&currentRoomInfo.PullPriority,
		&currentRoomInfo.HasFrosh,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query room info from rooms table"})
		return err
	}

	// log room uuid
	log.Println(currentRoomInfo.RoomUUID)

	// make sure the room does not have frosh
	if currentRoomInfo.HasFrosh {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Room has frosh"})
		err = errors.New("room has frosh")
		return err
	}

	if currentRoomInfo.PullPriority.IsPreplaced {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot pull into a preplaced room"})
		err = errors.New("room is preplaced")
		return err
	}

	// check that the proposed occupants are not more than the max occupancy
	if len(proposedOccupants) > currentRoomInfo.MaxOccupancy {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Proposed occupants exceeds max occupancy"})
		err = errors.New("proposed occupants exceeds max occupancy")
		return err
	}

	var occupantsAlreadyInRoom models.IntArray
	rows, err := tx.Query("SELECT id FROM users WHERE id = ANY($1) AND room_uuid IS NOT NULL", pq.Array(proposedOccupants))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query room_uuid from users table"})
		return err
	}

	for rows.Next() {
		var occupant int
		if err := rows.Scan(&occupant); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan room_uuid from users table"})
			return err
		}

		occupantsAlreadyInRoom = append(occupantsAlreadyInRoom, occupant)
	}

	if len(occupantsAlreadyInRoom) > 0 {
		err = errors.New("one or more of the proposed occupants is already in a room")
		c.JSON(http.StatusBadRequest, gin.H{"error": "One or more of the proposed occupants is already in a room", "occupants": occupantsAlreadyInRoom})
		return err
	}

	var proposedPullPriority models.PullPriority
	log.Println(request.PullType)

	// can only lock pull info an empty room
	if currentRoomInfo.CurrentOccupancy > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Lock pull is only allowed for empty rooms"})
		err = errors.New("lock pull is only allowed for empty rooms")
		tx.Rollback()
		return err
	}

	if len(proposedOccupants) == 0 {
		email := c.MustGet("email").(string)
		err = clearRoom(currentRoomInfo.RoomUUID, tx, notificationQueue, email)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove the current occupants of the room"})
			return err
		}
		return nil
	}

	// lock pulled room must be full
	if len(proposedOccupants) != currentRoomInfo.MaxOccupancy {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Lock pull requires the room to be full"})
		err = errors.New("lock pull requires the room to be full")
		tx.Rollback()
		return err
	}

	// get suite info for the room
	suiteInfo := models.SuiteRaw{}

	err = tx.QueryRow("SELECT suite_uuid, dorm, dorm_name, floor, room_count, rooms, alternative_pull, can_lock_pull FROM suites WHERE suite_uuid = $1", currentRoomInfo.SuiteUUID).Scan(
		&suiteInfo.SuiteUUID,
		&suiteInfo.Dorm,
		&suiteInfo.DormName,
		&suiteInfo.Floor,
		&suiteInfo.RoomCount,
		&suiteInfo.Rooms,
		&suiteInfo.AlternativePull,
		&suiteInfo.CanLockPull,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query suite info from suites table"})
		tx.Rollback()
	}

	// ensure that lock pull is allowed for the suite
	if !suiteInfo.CanLockPull {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Lock pull is not allowed for the suite"})
		tx.Rollback()
		return err
	}

	// query all the rooms and ensure that they are full
	var roomsInSuite []models.RoomRaw

	rows, err = tx.Query("SELECT room_uuid, dorm, dorm_name, room_id, suite_uuid, max_occupancy, current_occupancy, occupants, pull_priority, sgroup_uuid, has_frosh FROM rooms WHERE suite_uuid = $1", currentRoomInfo.SuiteUUID)
	if err != nil {

		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed on rooms for pull priority"})
		tx.Rollback()
		return err
	}

	for rows.Next() {
		var r models.RoomRaw
		if err := rows.Scan(&r.RoomUUID, &r.Dorm, &r.DormName, &r.RoomID, &r.SuiteUUID, &r.MaxOccupancy, &r.CurrentOccupancy, &r.Occupants, &r.PullPriority, &r.SGroupUUID, &r.HasFrosh); err != nil {
			// Handle scan error
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database scan failed on rooms for pull priority"})
			tx.Rollback()
			return err
		}
		roomsInSuite = append(roomsInSuite, r)
	}

	nonPreplacedRooms := 0

	for _, roomInSuite := range roomsInSuite {
		if roomInSuite.CurrentOccupancy < roomInSuite.MaxOccupancy && !roomInSuite.HasFrosh && roomInSuite.RoomUUID != currentRoomInfo.RoomUUID {
			c.JSON(http.StatusBadRequest, gin.H{"error": "One or more rooms in the suite are not full " + roomInSuite.RoomID})
			tx.Rollback()
			return err
		}

		// check if pull priority is not preplace
		if !roomInSuite.PullPriority.IsPreplaced && !roomInSuite.HasFrosh && roomInSuite.RoomUUID != currentRoomInfo.RoomUUID {
			nonPreplacedRooms++
		}
	}

	if nonPreplacedRooms == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot lock pull into a suite with all preplaced rooms"}) // all rooms in the suite are preplaced
		err = errors.New("cannot lock pull into a suite with all preplaced rooms")
		tx.Rollback()
		return err
	}

	var occupantsInfo []models.UserRaw
	rows, err = tx.Query("SELECT id, draw_number, year, in_dorm, participated, preplaced FROM users WHERE id = ANY($1)", pq.Array(proposedOccupants))
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed on users for pull priority"})
		tx.Rollback()
		return err
	}
	for rows.Next() {
		var u models.UserRaw
		if err := rows.Scan(&u.Id, &u.DrawNumber, &u.Year, &u.InDorm, &u.Participated, &u.Preplaced); err != nil {
			// Handle scan error
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database scan failed on users for pull priority"})
			tx.Rollback()
			return err
		}

		// if any of the proposed occupants are preplaced, return an error
		if u.Preplaced {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot pull with a preplaced user"})
			err = errors.New("cannot pull with a preplaced user")
			tx.Rollback()
			return err
		}

		occupantsInfo = append(occupantsInfo, u)
	}

	// for all users who currently have not participated, set their participated field to true and partitipation time to now
	_, err = tx.Exec("UPDATE users SET participated = true, participation_time = NOW() WHERE id = ANY($1) AND participated = false", pq.Array(proposedOccupants))
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update participated field in users table"})
		tx.Rollback()
		return err
	}

	sortedOccupants := sortUsersByPriority(occupantsInfo, currentRoomInfo.Dorm)

	proposedPullPriority = generateUserPriority(sortedOccupants[0], currentRoomInfo.Dorm)

	proposedPullPriority.Valid = true
	proposedPullPriority.PullType = 3
	proposedPullPriority.Inherited.Valid = true

	if currentRoomInfo.PullPriority.PullType == 3 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot bump a lock pulled room"})
		err = errors.New("cannot bump a lock pulled room")
		tx.Rollback()
		return err
	}

	if !comparePullPriority(proposedPullPriority, currentRoomInfo.PullPriority) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Proposed occupants do not have higher priority than current occupants"})
		err = errors.New("proposed occupants do not have higher priority than current occupants")
		tx.Rollback()
		return err
	}

	// disband the suite group if there is one
	if currentRoomInfo.SGroupUUID != uuid.Nil {
		_, err := disbandSuiteGroup(currentRoomInfo.SGroupUUID, tx)
		if err != nil {
			// use err in the response
			c.JSON(http.StatusInternalServerError, err.Error())
			return err
		}
	}

	if currentRoomInfo.CurrentOccupancy > 0 {
		// use clearRoom function to remove the current occupants from the room
		email := c.MustGet("email").(string)
		err = clearRoom(currentRoomInfo.RoomUUID, tx, notificationQueue, email)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove the current occupants of the room"})
			tx.Rollback()
			return err
		}
	}

	// update the occupants in the database and the current_occupancy
	_, err = tx.Exec("UPDATE rooms SET occupants = $1, current_occupancy = $2 WHERE room_uuid = $3", pq.Array(proposedOccupants), len(proposedOccupants), roomUUIDParam)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return err
	}

	// for each occupant, update the room_uuid field in the users table
	for _, proposedOccupant := range proposedOccupants {
		_, err = tx.Exec("UPDATE users SET room_uuid = $1 WHERE id = $2", roomUUIDParam, proposedOccupant)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update room_uuid in users table"})
			return err
		}
	}

	// update the pull_priority field in the rooms table
	proposedPullPriorityJSON, err := json.Marshal(proposedPullPriority)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal proposed pull priority"})
		return err
	}

	_, err = tx.Exec("UPDATE rooms SET pull_priority = $1 WHERE room_uuid = $2", proposedPullPriorityJSON, roomUUIDParam)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update pull_priority in rooms table"})
		return err
	}

	// update the suite's lock pull status
	_, err = tx.Exec("UPDATE suites SET lock_pulled_room = $1 WHERE suite_uuid = $2", roomUUIDParam, currentRoomInfo.SuiteUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update lock_pulled_room in suites table"})
		return err
	}

	// Update gender preferences for the suite
	err = UpdateSuiteGenderPreferencesBySuiteUUID(tx, currentRoomInfo.SuiteUUID)
	if err != nil {
		// If there's a gender preference conflict, fail the transaction
		if strings.Contains(err.Error(), "no valid intersection of gender preferences") {
			c.JSON(http.StatusConflict, gin.H{"error": "Cannot pull users with incompatible gender preferences. The users must have at least one gender preference in common."})
			return err
		}

		// For other errors, log a warning but continue
		log.Printf("Warning: Failed to update gender preferences for suite %s: %v", currentRoomInfo.SuiteUUID, err)
	}

	return nil
}

func AlternativePull(c *gin.Context, request models.OccupantUpdateRequest) error {
	// the room uuid is in the url
	roomUUIDParam := c.Param("roomuuid")

	proposedOccupants := request.ProposedOccupants

	// Convert proposedOccupants from []int to []string
	proposedOccupantStrings := make([]string, len(proposedOccupants))
	for i, occupant := range proposedOccupants {
		proposedOccupantStrings[i] = strconv.Itoa(occupant)
	}

	userFullName, exists := c.Get("user_full_name")
	if !exists {
		log.Print("Error: user_full_name not found in context")
		userFullName = "unknown user"
	}

	userEmail, exists := c.Get("email")
	if !exists {
		log.Print("Error: email not found in context")
		userEmail = "unknown user email"
	}

	previousRoomState, err := getRoomStateRaw(roomUUIDParam)
	if err != nil { /* ... handle error ... */
		return err
	}
	if previousRoomState == nil { /* ... handle not found ... */
		return sql.ErrNoRows
	}
	// Pull Leader Room
	previousPullLeaderRoomState, err := getRoomStateRaw(request.PullLeaderRoom.String())
	if err != nil { /* ... handle error ... */
		return err
	}
	if previousPullLeaderRoomState == nil { /* ... handle leader not found ... */
		return sql.ErrNoRows
	}

	log.Println(userFullName.(string) + " is attempting an alternative pull for room " + roomUUIDParam + " with occupants " + strings.Join(proposedOccupantStrings, ", "))

	// verify that the proposed occupants are unique
	proposedOccupantsMap := make(map[int]bool)
	for _, occupant := range proposedOccupants {
		if proposedOccupantsMap[occupant] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Duplicate user was specified in the occupants list"})
			return errors.New("duplicate user was specified in the occupants list")
		}
		proposedOccupantsMap[occupant] = true
	}

	// Create notification queue
	notificationQueue := models.NewBumpNotificationQueue()

	// Start a transaction
	tx, err := database.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return err
	}

	// Defer rollback/commit and logging logic
	var commitErr error
	var genderUpdateErr error
	var createdSGroupUUID uuid.UUID                        // Store created group ID for logging
	var alternativeGroupPriorityForLog models.PullPriority // Store calculated priority

	defer func() {
		if r := recover(); r != nil { /* ... handle panic rollback ... */
			return
		}
		if err != nil { /* ... handle error rollback ... */
			return
		}
		commitErr = tx.Commit()
		if commitErr != nil { /* ... handle commit error ... */
			return
		}

		// --- COMMIT SUCCEEDED ---
		log.Printf("Successfully committed ALTERNATIVE_PULL for room %s by %s", roomUUIDParam, userEmail)

		// Send Bump Notifications
		for _, notification := range notificationQueue.Notifications {
			go SendBumpNotification(notification.UserID, notification.RoomID, notification.DormName)
		}

		// --- Transactional Logging: Get New State & Log ---
		newRoomState, fetchErr := getRoomStateRaw(roomUUIDParam)
		if fetchErr != nil {
			log.Printf("Error fetching new room state %s: %v", roomUUIDParam, fetchErr)
		}
		newPullLeaderRoomState, fetchLeaderErr := getRoomStateRaw(request.PullLeaderRoom.String())
		if fetchLeaderErr != nil {
			log.Printf("Error fetching new leader room state %s: %v", request.PullLeaderRoom.String(), fetchLeaderErr)
		}

		// Prepare details
		bumpedOccupantIDs := make([]int, 0, len(notificationQueue.Notifications))
		for _, n := range notificationQueue.Notifications {
			bumpedOccupantIDs = append(bumpedOccupantIDs, n.UserID)
		}

		logDetails := map[string]interface{}{
			"proposed_occupants":          proposedOccupants,
			"pull_leader_room":            request.PullLeaderRoom,
			"previous_occupants":          previousRoomState.Occupants,
			"previous_sgroup_uuid":        previousRoomState.SGroupUUID,
			"previous_leader_sgroup_uuid": previousPullLeaderRoomState.SGroupUUID,
			"bumped_occupant_ids":         bumpedOccupantIDs,
			"created_sgroup_uuid":         createdSGroupUUID,              // Will be non-nil if group was created
			"alternative_group_priority":  alternativeGroupPriorityForLog, // The calculated 2nd best priority
			"gender_update_error":         nil,
		}
		if genderUpdateErr != nil {
			logDetails["gender_update_error"] = genderUpdateErr.Error()
		}

		// Log primary operation on the target room
		loggingErr := logging.LogOperation(
			c, "ALTERNATIVE_PULL", models.EntityTypeRoom, roomUUIDParam,
			previousRoomState, // Target room before
			newRoomState,      // Target room after
			logDetails,
		)
		if loggingErr != nil {
			log.Printf("WARNING: Failed log ALT_PULL target room %s: %v", roomUUIDParam, loggingErr)
		}

		// Log change to the pull leader room
		// Determine if leader state changed enough to warrant a separate log entry
		leaderStateChanged := !compareRoomStatesForLogging(previousPullLeaderRoomState, newPullLeaderRoomState) // Implement compareRoomStatesForLogging

		if leaderStateChanged {
			leaderLogDetails := map[string]interface{}{
				"pull_action_target_room": roomUUIDParam,
				"related_occupants":       proposedOccupants,
				"reason":                  "Participant in ALTERNATIVE_PULL",
				"previous_sgroup_uuid":    previousPullLeaderRoomState.SGroupUUID,
				"new_sgroup_uuid":         newPullLeaderRoomState.SGroupUUID,
				"previous_priority":       previousPullLeaderRoomState.PullPriority,
				"new_priority":            newPullLeaderRoomState.PullPriority, // Check if priority.inherited changed
			}
			leaderLoggingErr := logging.LogOperation(
				c, "UPDATE_ROOM_STATE", models.EntityTypeRoom, request.PullLeaderRoom.String(),
				previousPullLeaderRoomState, // Leader room before
				newPullLeaderRoomState,      // Leader room after
				leaderLogDetails,
			)
			if leaderLoggingErr != nil {
				log.Printf("WARNING: Failed log ALT_PULL leader room %s: %v", request.PullLeaderRoom.String(), leaderLoggingErr)
			}
		}

		// Send success response *after* logging attempt
		c.JSON(http.StatusOK, gin.H{"message": "Successfully updated occupants"})

	}() // End of defer func

	var currentRoomInfo models.RoomRaw
	err = tx.QueryRow("SELECT room_uuid, dorm, dorm_name, room_id, suite_uuid, max_occupancy, current_occupancy, occupants, pull_priority, has_frosh FROM rooms WHERE room_uuid = $1", roomUUIDParam).Scan(
		&currentRoomInfo.RoomUUID,
		&currentRoomInfo.Dorm,
		&currentRoomInfo.DormName,
		&currentRoomInfo.RoomID,
		&currentRoomInfo.SuiteUUID,
		&currentRoomInfo.MaxOccupancy,
		&currentRoomInfo.CurrentOccupancy,
		&currentRoomInfo.Occupants,
		&currentRoomInfo.PullPriority,
		&currentRoomInfo.HasFrosh,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query room info from rooms table"})
		return err
	}

	// log room uuid
	log.Println(currentRoomInfo.RoomUUID)

	// make sure the room does not have frosh
	if currentRoomInfo.HasFrosh {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Room has frosh"})
		err = errors.New("room has frosh")
		return err
	}

	if currentRoomInfo.PullPriority.IsPreplaced {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot pull into a preplaced room"})
		err = errors.New("room is preplaced")
		return err
	}

	// check that the proposed occupants are not more than the max occupancy
	if len(proposedOccupants) > currentRoomInfo.MaxOccupancy {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Proposed occupants exceeds max occupancy"})
		return err
	}

	var occupantsAlreadyInRoom models.IntArray
	rows, err := tx.Query("SELECT id FROM users WHERE id = ANY($1) AND room_uuid IS NOT NULL", pq.Array(proposedOccupants))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query room_uuid from users table"})
		return err
	}

	for rows.Next() {
		var occupant int
		if err := rows.Scan(&occupant); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan room_uuid from users table"})
			return err
		}

		occupantsAlreadyInRoom = append(occupantsAlreadyInRoom, occupant)
	}

	if len(occupantsAlreadyInRoom) > 0 {
		err = errors.New("one or more of the proposed occupants is already in a room")
		c.JSON(http.StatusBadRequest, gin.H{"error": "One or more of the proposed occupants is already in a room", "occupants": occupantsAlreadyInRoom})
		return err
	}

	var proposedPullPriority models.PullPriority
	var pullLeaderPriority models.PullPriority
	var alternativeGroupPriority models.PullPriority
	var pullLeaderSuiteGroupUUID uuid.UUID
	log.Println(request.PullType)

	if len(proposedOccupants) != currentRoomInfo.MaxOccupancy {
		// error because normal pull requires a full room
		c.JSON(http.StatusBadRequest, gin.H{"error": "Alternative pull requires the room to be full"})
		tx.Rollback()
		return err
	}

	if currentRoomInfo.RoomUUID == request.PullLeaderRoom {
		// error because the pull leader is already in the room
		c.JSON(http.StatusBadRequest, gin.H{"error": "Pull leader is already in the room"})
		err = errors.New("pull leader is already in the room")
		tx.Rollback()
		return err
	}

	// get suite info for the room
	suiteInfo := models.SuiteRaw{}

	err = tx.QueryRow("SELECT suite_uuid, dorm, dorm_name, floor, room_count, rooms, alternative_pull FROM suites WHERE suite_uuid = $1", currentRoomInfo.SuiteUUID).Scan(
		&suiteInfo.SuiteUUID,
		&suiteInfo.Dorm,
		&suiteInfo.DormName,
		&suiteInfo.Floor,
		&suiteInfo.RoomCount,
		&suiteInfo.Rooms,
		&suiteInfo.AlternativePull,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query suite info from suites table"})
		tx.Rollback()
		return err
	}

	// ensure that alternative pull is allowed for the suite
	if !suiteInfo.AlternativePull {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Alternative pull is not allowed for the suite"})
		tx.Rollback()
		return err
	}

	pullLeaderRoomUUID := request.PullLeaderRoom
	var occupantsInfo []models.UserRaw
	rows, err = tx.Query("SELECT id, draw_number, year, in_dorm, participated, preplaced FROM users WHERE id = ANY($1)", pq.Array(proposedOccupants))
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed on users for pull priority"})
		tx.Rollback()
		return err
	}

	for rows.Next() {
		var u models.UserRaw
		if err := rows.Scan(&u.Id, &u.DrawNumber, &u.Year, &u.InDorm, &u.Participated, &u.Preplaced); err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database scan failed on users for pull priority"})
			tx.Rollback()
			return err
		}

		// if any of the proposed occupants are preplaced, return an error
		if u.Preplaced {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot pull with a preplaced user"})
			err = errors.New("cannot pull with a preplaced user")
			tx.Rollback()
			return err
		}

		occupantsInfo = append(occupantsInfo, u)
	}

	// for all users who currently have not participated, set their participated field to true and partitipation time to now
	_, err = tx.Exec("UPDATE users SET participated = true, participation_time = NOW() WHERE id = ANY($1) AND participated = false", pq.Array(proposedOccupants))
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update participated field in users table"})
		tx.Rollback()
		return err
	}

	var suiteUUID = currentRoomInfo.SuiteUUID
	var leaderSuiteUUID uuid.UUID
	var pullLeaderCurrentOccupancy int
	var pullLeaderMaxOccupancy int

	log.Println("pull leader room uuid: " + pullLeaderRoomUUID.String())

	// get the pull leader's info
	err = tx.QueryRow("SELECT pull_priority, sgroup_uuid, suite_uuid, current_occupancy, max_occupancy FROM rooms WHERE room_uuid = $1", pullLeaderRoomUUID).Scan(&pullLeaderPriority, &pullLeaderSuiteGroupUUID, &leaderSuiteUUID, &pullLeaderCurrentOccupancy, &pullLeaderMaxOccupancy)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query pull leader's info from rooms table"})
		log.Println(err)
		tx.Rollback()
		return err
	}

	if leaderSuiteUUID != suiteUUID {
		// error because the pull leader is not in the same suite
		c.JSON(http.StatusBadRequest, gin.H{"error": "Pull leader is not in the same suite"})
		tx.Rollback()
		return err
	}

	if pullLeaderCurrentOccupancy != pullLeaderMaxOccupancy {
		log.Println(pullLeaderCurrentOccupancy, pullLeaderMaxOccupancy)
		// error because the pull leader is not in a single
		c.JSON(http.StatusBadRequest, gin.H{"error": "You can only initiate an alternative pull with a full room"})
		tx.Rollback()
		return err
	}

	// get all of the users in the pull leader's room
	var pullLeaderOccupantsInfo []models.UserRaw
	rows, err = tx.Query("SELECT id, draw_number, year, in_dorm FROM users WHERE room_uuid = $1", pullLeaderRoomUUID)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed on users for pull priority"})
		tx.Rollback()
		return err
	}

	for rows.Next() {
		var u models.UserRaw
		if err := rows.Scan(&u.Id, &u.DrawNumber, &u.Year, &u.InDorm); err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database scan failed on users for pull priority"})
			tx.Rollback()
			return err
		}
		pullLeaderOccupantsInfo = append(pullLeaderOccupantsInfo, u)
	}

	// loop through all occupants and check if they have in dorm
	// if at least one does not have in dorm, change each of the pull priorities of each occupant to not have in dorm
	for _, occupant := range occupantsInfo {
		if occupant.InDorm != currentRoomInfo.Dorm {
			log.Println("Forfeited in dorm to pull non-in dorm user")
			for i := range occupantsInfo {
				occupantsInfo[i].InDorm = 0
			}
			break
		}
	}

	var allOccupantsInfo = append(occupantsInfo, pullLeaderOccupantsInfo...)

	// loop through all occupants and check if they have in dorm
	// if at least one does not have in dorm, change each of the pull priorities of each occupant to not have in dorm
	for _, occupant := range allOccupantsInfo {
		if occupant.InDorm != currentRoomInfo.Dorm {
			log.Println("Forfeited in dorm to pull non-in dorm user")
			for i := range allOccupantsInfo {
				allOccupantsInfo[i].InDorm = 0
			}
			break
		}
	}

	// sort all of the occupants by priority and get the second highest priority
	sortedAllOccupants := sortUsersByPriority(allOccupantsInfo, currentRoomInfo.Dorm)

	sortedProposedOccupants := sortUsersByPriority(occupantsInfo, currentRoomInfo.Dorm)

	proposedPullPriority = generateUserPriority(sortedProposedOccupants[0], currentRoomInfo.Dorm)
	proposedPullPriority.Valid = true
	proposedPullPriority.PullType = 4

	alternativeGroupPriority = generateUserPriority(sortedAllOccupants[1], currentRoomInfo.Dorm)

	proposedPullPriority.Inherited.Valid = true
	proposedPullPriority.Inherited.DrawNumber = alternativeGroupPriority.DrawNumber
	proposedPullPriority.Inherited.HasInDorm = alternativeGroupPriority.HasInDorm
	proposedPullPriority.Inherited.Year = alternativeGroupPriority.Year

	if currentRoomInfo.PullPriority.PullType == 3 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot bump a lock pulled room"})
		err = errors.New("cannot bump a lock pulled room")
		tx.Rollback()
		return err
	}

	if !comparePullPriority(proposedPullPriority, currentRoomInfo.PullPriority) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Proposed occupants do not have higher priority than current occupants"})
		err = errors.New("proposed occupants do not have higher priority than current occupants")
		tx.Rollback()
		return err
	}

	log.Println(proposedOccupants)

	// disband the suite group if there is one
	if currentRoomInfo.SGroupUUID != uuid.Nil {
		_, err := disbandSuiteGroup(currentRoomInfo.SGroupUUID, tx)
		if err != nil {
			// use err in the response
			c.JSON(http.StatusInternalServerError, err.Error())
			return err
		}
	}

	if currentRoomInfo.CurrentOccupancy > 0 {
		// remove the current occupants from the room
		email := c.MustGet("email").(string)
		err = clearRoom(currentRoomInfo.RoomUUID, tx, notificationQueue, email)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove the current occupants of the room"})
			tx.Rollback()
		}
	}

	// update the occupants in the database and the current_occupancy
	_, err = tx.Exec("UPDATE rooms SET occupants = $1, current_occupancy = $2 WHERE room_uuid = $3", pq.Array(proposedOccupants), len(proposedOccupants), roomUUIDParam)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return err
	}

	// for each occupant, update the room_uuid field in the users table
	for _, proposedOccupant := range proposedOccupants {
		_, err = tx.Exec("UPDATE users SET room_uuid = $1 WHERE id = $2", roomUUIDParam, proposedOccupant)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update room_uuid in users table"})
			return err
		}
	}

	// update the pull_priority field in the rooms table
	proposedPullPriorityJSON, err := json.Marshal(proposedPullPriority)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal proposed pull priority"})
		return err
	}

	_, err = tx.Exec("UPDATE rooms SET pull_priority = $1 WHERE room_uuid = $2", proposedPullPriorityJSON, roomUUIDParam)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update pull_priority in rooms table"})
		return err
	}

	if pullLeaderSuiteGroupUUID != uuid.Nil {
		// error out because the pull leader is in a suite group
		c.JSON(http.StatusBadRequest, gin.H{"error": "Pull leader is in a suite group for alternative pull"})
		err = errors.New("pull leader is in a suite group for alternative pull")
		tx.Rollback()
		return err
	}

	// do the same thing as for pull type 2
	// create new suite group with the pull leader's priority
	alternativeGroupPriorityJSON, err := json.Marshal(alternativeGroupPriority)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal pull leader's pull priority"})
		return err
	}

	var suiteGroupUUID uuid.UUID = uuid.New()
	_, err = tx.Exec("INSERT INTO suitegroups (sgroup_uuid, sgroup_size, sgroup_name, sgroup_suite, pull_priority, rooms, disbanded) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		suiteGroupUUID,
		2,
		"Suite Group",
		currentRoomInfo.SuiteUUID,
		alternativeGroupPriorityJSON,
		pq.Array(models.UUIDArray{currentRoomInfo.RoomUUID, request.PullLeaderRoom}),
		false,
	)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert new suite group into suitegroups table"})
		return err
	}

	// update the sgroup_uuid field in the rooms table for both rooms
	_, err = tx.Exec("UPDATE rooms SET sgroup_uuid = $1 WHERE room_uuid = $2 OR room_uuid = $3", suiteGroupUUID, roomUUIDParam, request.PullLeaderRoom)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update suite_uuid in rooms table"})
		return err
	}

	// update the sgroup_uuid field in the users table for all occupants of both rooms
	_, err = tx.Exec("UPDATE users SET sgroup_uuid = $1 WHERE room_uuid = $2 OR room_uuid = $3", suiteGroupUUID, roomUUIDParam, request.PullLeaderRoom)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update sgroup_uuid in users table"})
		return err
	}

	_, err = tx.Exec(`
			UPDATE Rooms
			SET pull_priority = jsonb_set(
				jsonb_set(
					pull_priority,
					'{inherited}',
					jsonb_build_object(
						'year', $1::int,
						'valid', $2::bool,
						'drawNumber', $3::float,
						'hasInDorm', $4::bool
					)
				),
				'{pullType}',
				'4'::jsonb
			)
			WHERE room_uuid = $5`,
		alternativeGroupPriority.Year,
		true, // Assuming you want to set 'valid' to true directly
		alternativeGroupPriority.DrawNumber,
		alternativeGroupPriority.HasInDorm,
		request.PullLeaderRoom)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update inherited priority in rooms table"})
		return err
	}

	log.Println("Pull leader room: " + request.PullLeaderRoom.String())

	// Update gender preferences for the suite
	err = UpdateSuiteGenderPreferencesBySuiteUUID(tx, currentRoomInfo.SuiteUUID)
	if err != nil {
		// If there's a gender preference conflict, fail the transaction
		if strings.Contains(err.Error(), "no valid intersection of gender preferences") {
			c.JSON(http.StatusConflict, gin.H{"error": "Cannot pull users with incompatible gender preferences. The users must have at least one gender preference in common."})
			return err
		}

		// For other errors, log a warning but continue
		log.Printf("Warning: Failed to update gender preferences for suite %s: %v", currentRoomInfo.SuiteUUID, err)
	}

	return nil
}

func clearRoom(roomUUID uuid.UUID, tx *sql.Tx, notificationQueue *models.BumpNotificationQueue, requesterEmail string) error {
	// Look up the requester's user ID from their email
	var requester models.UserRaw

	err := tx.QueryRow("SELECT id, year, first_name, last_name FROM users WHERE email = $1", requesterEmail).Scan(
		&requester.Id, &requester.Year, &requester.FirstName, &requester.LastName,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			// User not found in database, continue with clearing the room
			// but don't exclude any occupant from notifications
			log.Printf("User with email %s not found in database", requesterEmail)
			requester.Id = -1 // Set to invalid ID
		} else {
			// Other database error
			return err
		}
	}

	// Get current occupants before clearing
	var currentOccupants models.IntArray
	err = tx.QueryRow("SELECT occupants FROM rooms WHERE room_uuid = $1", roomUUID).Scan(&currentOccupants)
	if err != nil {
		return err
	}

	// Get room ID and dorm name for notification
	var roomID, dormName string
	err = tx.QueryRow("SELECT room_id, dorm_name FROM rooms WHERE room_uuid = $1", roomUUID).Scan(&roomID, &dormName)
	if err != nil {
		return err
	}

	// Queue notifications for each occupant being bumped
	for _, occupantID := range currentOccupants {
		// if the occupant is the person who is clearing the room, don't send a notification
		if occupantID == requester.Id {
			log.Println("Occupant " + strconv.Itoa(occupantID) + " is the requester, so not sending a notification")
			continue
		}
		notificationQueue.Add(occupantID, roomID, dormName)
	}

	// get the suite group uuid
	var suiteGroupUUID uuid.UUID
	err = tx.QueryRow("SELECT sgroup_uuid FROM rooms WHERE room_uuid = $1", roomUUID).Scan(&suiteGroupUUID)
	if err != nil {
		return err
	}
	// if the room is in a suite group, disband the suite group
	log.Println("Clearing suite group with uuid " + suiteGroupUUID.String())
	if suiteGroupUUID != uuid.Nil {
		_, err := disbandSuiteGroup(suiteGroupUUID, tx)
		if err != nil {
			return err
		}
	}

	// check the suite if there is a lock pull
	var lockPulledRoomUUID uuid.UUID
	err = tx.QueryRow("SELECT lock_pulled_room FROM suites WHERE suite_uuid = (SELECT suite_uuid FROM rooms WHERE room_uuid = $1)", roomUUID).Scan(&lockPulledRoomUUID)
	if err != nil {
		return err
	}

	err = RemoveLockPull(roomUUID, tx)
	if err != nil {
		return err
	}

	var defaultPullPriority models.PullPriority = generateEmptyPriority()
	defaultPullPriorityJSON, err := json.Marshal(defaultPullPriority)

	if err != nil {
		return err
	}

	// clear the room by setting the occupants to nil and the current_occupancy to 0
	_, err = tx.Exec("UPDATE rooms SET occupants = $1, current_occupancy = $2, pull_priority = $3 WHERE room_uuid = $4", nil, 0, defaultPullPriorityJSON, roomUUID)
	if err != nil {
		return err
	}

	// for each occupant, set the room_uuid field in the users table to nil
	_, err = tx.Exec("UPDATE users SET room_uuid = $1 WHERE room_uuid = $2", nil, roomUUID)
	if err != nil {
		return err
	}

	return nil
}

// given a suite group uuid, disband the suite group and return the room uuids of the rooms in the suite group
func disbandSuiteGroup(sgroupUUID uuid.UUID, tx *sql.Tx) (models.UUIDArray, error) {
	// get the rooms in the suite group
	var roomsInSuiteGroup models.UUIDArray

	err := tx.QueryRow("SELECT rooms FROM suitegroups WHERE sgroup_uuid = $1", sgroupUUID).Scan(&roomsInSuiteGroup)
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec("UPDATE rooms SET pull_priority = jsonb_set(jsonb_set(pull_priority, '{inherited}', '{\"hasInDorm\": false, \"drawNumber\": 0, \"year\": 0}'::jsonb), '{pullType}', '1'::jsonb) WHERE sgroup_uuid = $1", sgroupUUID)
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec("UPDATE rooms SET sgroup_uuid = $1 WHERE sgroup_uuid = $2", nil, sgroupUUID)
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec("UPDATE users SET sgroup_uuid = $1 WHERE sgroup_uuid = $2", nil, sgroupUUID)
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec("DELETE FROM suitegroups WHERE sgroup_uuid = $1", sgroupUUID)
	if err != nil {
		return nil, err
	}

	log.Println("Disbanded suite group with uuid " + sgroupUUID.String())
	return roomsInSuiteGroup, nil
}

func PreplaceOccupants(c *gin.Context) {
	// the room uuid is in the url
	roomUUIDParam := c.Param("roomuuid")

	userFullName, exists := c.Get("user_full_name")
	if !exists {
		log.Print("Error: user_full_name not found in context")
		userFullName = "unknown user"
	}

	log.Println(userFullName.(string) + " is attempting to preplace occupants in room " + roomUUIDParam)

	// the request body should contain the occupants to be preplaced
	var request models.PreplacedRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	notificationQueue := models.NewBumpNotificationQueue()

	// Start a transaction
	tx, err := database.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}

	// Ensure the transaction is either committed or rolled back
	defer func() {
		if r := recover(); r != nil {
			log.Println("Result for " + userFullName.(string) + ": failed to preplace occupants in room " + roomUUIDParam + " because of panic " + r.(error).Error())
			tx.Rollback()
			panic(r)
		} else if err != nil {
			log.Println("Result for " + userFullName.(string) + ": failed to preplace occupants in room " + roomUUIDParam + " because of error " + err.Error())
			tx.Rollback()
		} else {
			log.Println("Result for " + userFullName.(string) + ": successfully preplaced occupants in room " + roomUUIDParam)
			err = tx.Commit()

			if err == nil {
				for _, notification := range notificationQueue.Notifications {
					SendBumpNotification(notification.UserID, notification.RoomID, notification.DormName)
				}
			}

			c.JSON(http.StatusOK, gin.H{"message": "Successfully preplaced occupants"})
		}
	}()

	var currentRoomInfo models.RoomRaw
	err = tx.QueryRow("SELECT room_uuid, dorm, dorm_name, room_id, suite_uuid, max_occupancy, current_occupancy, occupants, pull_priority, has_frosh FROM rooms WHERE room_uuid = $1", roomUUIDParam).Scan(
		&currentRoomInfo.RoomUUID,
		&currentRoomInfo.Dorm,
		&currentRoomInfo.DormName,
		&currentRoomInfo.RoomID,
		&currentRoomInfo.SuiteUUID,
		&currentRoomInfo.MaxOccupancy,
		&currentRoomInfo.CurrentOccupancy,
		&currentRoomInfo.Occupants,
		&currentRoomInfo.PullPriority,
		&currentRoomInfo.HasFrosh,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query room info from rooms table"})
		return
	}

	// log room uuid
	log.Println(currentRoomInfo.RoomUUID)

	// make sure the room does not have frosh
	if currentRoomInfo.HasFrosh {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Room has frosh"})
		err = errors.New("room has frosh")
		return
	}

	// if proposed occupants is empty, return an error - we should use the dedicated endpoint instead
	if len(request.ProposedOccupants) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "To remove preplaced occupants, use the dedicated endpoint instead"})
		err = errors.New("empty proposed occupants array not allowed in this endpoint")
		return
	}

	if currentRoomInfo.PullPriority.IsPreplaced {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot pull into a preplaced room"})
		err = errors.New("room is already preplaced")
		return
	}

	// check that the room is empty
	if currentRoomInfo.CurrentOccupancy > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Room is not empty"})
		err = errors.New("room is not empty")
		return
	}

	// check that the proposed occupants are not more than the max occupancy
	if len(request.ProposedOccupants) > currentRoomInfo.MaxOccupancy {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Proposed occupants exceeds max occupancy"})
		err = errors.New("proposed occupants exceeds max occupancy")
		return
	}

	// check that all proposed occupants are preplaced if not, return an error
	var nonPreplacedOccupants models.IntArray
	rows, err := tx.Query("SELECT id FROM users WHERE id = ANY($1) AND preplaced = false", pq.Array(request.ProposedOccupants))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query preplaced status from users table"})
		return
	}

	for rows.Next() {
		var occupant int
		if err := rows.Scan(&occupant); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan user id from query result"})
			return
		}
		nonPreplacedOccupants = append(nonPreplacedOccupants, occupant)
	}

	if len(nonPreplacedOccupants) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "One or more of the proposed occupants is not preplaced", "occupants": nonPreplacedOccupants})
		err = errors.New("one or more of the proposed occupants is not preplaced")
		return
	}

	var occupantsAlreadyInRoom models.IntArray
	rows, err = tx.Query("SELECT id FROM users WHERE id = ANY($1) AND room_uuid IS NOT NULL", pq.Array(request.ProposedOccupants))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query room_uuid from users table"})
		return
	}

	for rows.Next() {
		var occupant int
		if err := rows.Scan(&occupant); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan room_uuid from users table"})
			return
		}

		occupantsAlreadyInRoom = append(occupantsAlreadyInRoom, occupant)
	}

	if len(occupantsAlreadyInRoom) > 0 {
		err = errors.New("one or more of the proposed occupants is already in a room")
		c.JSON(http.StatusBadRequest, gin.H{"error": "One or more of the proposed occupants is already in a room", "occupants": occupantsAlreadyInRoom})
		return
	}

	// get the suite info for the room
	suiteInfo := models.SuiteRaw{}

	err = tx.QueryRow("SELECT suite_uuid, dorm, dorm_name, floor, room_count, rooms, reslife_room FROM suites WHERE suite_uuid = $1", currentRoomInfo.SuiteUUID).Scan(
		&suiteInfo.SuiteUUID,
		&suiteInfo.Dorm,
		&suiteInfo.DormName,
		&suiteInfo.Floor,
		&suiteInfo.RoomCount,
		&suiteInfo.Rooms,
		&suiteInfo.ReslifeRoom,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query suite info from suites table"})
		return
	}

	// check if either user has reslife Role of mentor or proctor, reslife_role column is in the users table and needs to be 'mentor' or 'proctor'
	var isReslife bool
	for _, occupant := range request.ProposedOccupants {
		var reslifeRole string
		err = tx.QueryRow("SELECT reslife_role FROM users WHERE id = $1", occupant).Scan(&reslifeRole)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query reslife role from users table"})
			return
		}
		if reslifeRole == "mentor" || reslifeRole == "proctor" {
			isReslife = true
			break
		}
	}

	// if isReslife is true, set the reslife_room field in the suites table to the room_uuid
	if isReslife {
		_, err = tx.Exec("UPDATE suites SET reslife_room = $1 WHERE suite_uuid = $2", currentRoomInfo.RoomUUID, currentRoomInfo.SuiteUUID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update reslife_room in suites table"})
			return
		}
	}

	// pull priority for the proposed occupants
	proposedPullPriority := generateEmptyPriority()
	proposedPullPriority.Valid = true
	proposedPullPriority.IsPreplaced = true

	// for each occupant, update the room_uuid field in the users table
	for _, proposedOccupant := range request.ProposedOccupants {
		_, err = tx.Exec("UPDATE users SET room_uuid = $1 WHERE id = $2", currentRoomInfo.RoomUUID, proposedOccupant)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update room_uuid in users table"})
			return
		}
	}

	// update the occupants in the database and the current_occupancy
	_, err = tx.Exec("UPDATE rooms SET occupants = $1, current_occupancy = $2 WHERE room_uuid = $3", pq.Array(request.ProposedOccupants), len(request.ProposedOccupants), roomUUIDParam)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// update the pull_priority field in the rooms table
	proposedPullPriorityJSON, err := json.Marshal(proposedPullPriority)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal proposed pull priority"})
		return
	}

	_, err = tx.Exec("UPDATE rooms SET pull_priority = $1 WHERE room_uuid = $2", proposedPullPriorityJSON, roomUUIDParam)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update pull_priority in rooms table"})
		return
	}

	// Update gender preferences for the suite
	updateErr := UpdateSuiteGenderPreferencesBySuiteUUID(tx, currentRoomInfo.SuiteUUID)
	if updateErr != nil {
		// If there's a gender preference conflict, fail the transaction
		if strings.Contains(updateErr.Error(), "no valid intersection of gender preferences") {
			c.JSON(http.StatusConflict, gin.H{"error": "Cannot preplace users with incompatible gender preferences. The preplaced users must have at least one gender preference in common."})
			err = updateErr // Set err to updateErr so the transaction will be rolled back
			return
		}

		// For other errors, log a warning but continue
		log.Printf("Warning: Failed to update gender preferences for suite %s: %v", currentRoomInfo.SuiteUUID, updateErr)
	}
}

// RemovePreplacedOccupantsHandler is a separate handler to specifically remove preplaced occupants from a room
func RemovePreplacedOccupantsHandler(c *gin.Context) {
	// Get the room UUID from the URL
	roomUUIDParam := c.Param("roomuuid")

	// Get the user's full name for logging
	userFullName, exists := c.Get("user_full_name")
	if !exists {
		log.Print("Error: user_full_name not found in context")
		userFullName = "unknown user"
	}

	log.Println(userFullName.(string) + " is attempting to remove preplaced occupants from room " + roomUUIDParam)

	// Create a notification queue for any occupants that need to be notified
	notificationQueue := models.NewBumpNotificationQueue()

	// Start a transaction
	tx, err := database.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}

	// Ensure the transaction is either committed or rolled back
	defer func() {
		if r := recover(); r != nil {
			log.Println("Result for " + userFullName.(string) + ": failed to remove preplaced occupants from room " + roomUUIDParam + " because of panic " + r.(error).Error())
			tx.Rollback()
			panic(r)
		} else if err != nil {
			log.Println("Result for " + userFullName.(string) + ": failed to remove preplaced occupants from room " + roomUUIDParam + " because of error " + err.Error())
			tx.Rollback()
		} else {
			log.Println("Result for " + userFullName.(string) + ": successfully removed preplaced occupants from room " + roomUUIDParam)
			err = tx.Commit()

			if err == nil {
				for _, notification := range notificationQueue.Notifications {
					SendBumpNotification(notification.UserID, notification.RoomID, notification.DormName)
				}
			}

			c.JSON(http.StatusOK, gin.H{"message": "Successfully removed preplaced occupants"})
		}
	}()

	// Get the current room information
	var currentRoomInfo models.RoomRaw
	err = tx.QueryRow("SELECT room_uuid, dorm, dorm_name, room_id, suite_uuid, max_occupancy, current_occupancy, occupants, pull_priority, has_frosh FROM rooms WHERE room_uuid = $1", roomUUIDParam).Scan(
		&currentRoomInfo.RoomUUID,
		&currentRoomInfo.Dorm,
		&currentRoomInfo.DormName,
		&currentRoomInfo.RoomID,
		&currentRoomInfo.SuiteUUID,
		&currentRoomInfo.MaxOccupancy,
		&currentRoomInfo.CurrentOccupancy,
		&currentRoomInfo.Occupants,
		&currentRoomInfo.PullPriority,
		&currentRoomInfo.HasFrosh,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query room info from rooms table"})
		return
	}

	// Check if the room has frosh
	if currentRoomInfo.HasFrosh {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Room has frosh"})
		err = errors.New("room has frosh")
		return
	}

	// Check if the room is preplaced
	if !currentRoomInfo.PullPriority.IsPreplaced {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Room is not preplaced"})
		err = errors.New("room is not preplaced")
		return
	}

	// Get the user's email for clearing the room
	email := c.MustGet("email").(string)

	// Clear the room
	err = clearRoom(currentRoomInfo.RoomUUID, tx, notificationQueue, email)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove the occupants of the room"})
		return
	}

	// Update gender preferences for the suite after clearing the room
	updateErr := UpdateSuiteGenderPreferencesBySuiteUUID(tx, currentRoomInfo.SuiteUUID)
	if updateErr != nil {
		log.Printf("Warning: Failed to update gender preferences for suite %s after removing preplaced occupants: %v", currentRoomInfo.SuiteUUID, updateErr)
	} else {
		log.Printf("Successfully updated gender preferences for suite %s after removing preplaced occupants", currentRoomInfo.SuiteUUID)
	}
}

// remove a lock pull from a suite given the room uuid
func RemoveLockPull(roomUUID uuid.UUID, tx *sql.Tx) error {
	// check the suite if there is a lock pull
	var lockPulledRoomUUID uuid.UUID
	err := tx.QueryRow("SELECT lock_pulled_room FROM suites WHERE suite_uuid = (SELECT suite_uuid FROM rooms WHERE room_uuid = $1)", roomUUID).Scan(&lockPulledRoomUUID)
	if err != nil {
		return err
	}

	if lockPulledRoomUUID != uuid.Nil {
		// set the lock_pulled_room field in the suites table to nil
		_, err = tx.Exec("UPDATE suites SET lock_pulled_room = $1 WHERE suite_uuid = (SELECT suite_uuid FROM rooms WHERE room_uuid = $2)", nil, roomUUID)
		if err != nil {
			return err
		}

		if lockPulledRoomUUID != roomUUID {
			// set the pull_type field in the pull_priority to 1 and the inherited.valid field to false
			_, err = tx.Exec(`
					UPDATE rooms 
					SET pull_priority = jsonb_set(
											jsonb_set(
												pull_priority, 
												'{pullType}', 
												'1'::jsonb
											),
											'{inherited,valid}', 
											'false'::jsonb
										)
					WHERE room_uuid = $1`,
				lockPulledRoomUUID)
			if err != nil {
				return err
			}

		}
	}

	return nil
}

func GetRoom(c *gin.Context) {
	roomUUIDParam := c.Param("roomuuid")

	// Start a transaction
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

	var roomInfo models.RoomRaw
	err = tx.QueryRow("SELECT room_uuid, dorm, dorm_name, room_id, suite_uuid, max_occupancy, current_occupancy, occupants, pull_priority, sgroup_uuid, has_frosh FROM rooms WHERE room_uuid = $1", roomUUIDParam).Scan(
		&roomInfo.RoomUUID,
		&roomInfo.Dorm,
		&roomInfo.DormName,
		&roomInfo.RoomID,
		&roomInfo.SuiteUUID,
		&roomInfo.MaxOccupancy,
		&roomInfo.CurrentOccupancy,
		&roomInfo.Occupants,
		&roomInfo.PullPriority,
		&roomInfo.SGroupUUID,
		&roomInfo.HasFrosh,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query room info"})
		return
	}

	c.JSON(http.StatusOK, roomInfo)
}

// Add this new function after one of the existing handler functions

// ClearRoomHandler handles clearing all occupants from a room
func ClearRoomHandler(c *gin.Context) {
	// Get the room UUID from the URL
	roomUUIDParam := c.Param("roomuuid")

	// Get the user's email
	email, exists := c.Get("email")
	if !exists {
		log.Print("Error: email not found in context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User email not found"})
		return
	}
	emailStr := email.(string)

	// Log the user who is attempting to clear the room
	userFullName, exists := c.Get("user_full_name")
	if !exists {
		log.Print("Error: user_full_name not found in context")
		userFullName = "unknown user"
	}
	log.Println(userFullName.(string) + " is attempting to clear room " + roomUUIDParam)

	const MAX_DAILY_CLEARS = 10 // Keep your constant

	// Get the current date in Pacific Time (needed for comparison and potential insert/update)
	loc, errLoadLocation := time.LoadLocation("America/Los_Angeles")
	if errLoadLocation != nil {
		log.Printf("Error loading Pacific timezone: %v", errLoadLocation)
		loc = time.UTC // Fallback
	}
	pacificNow := time.Now().In(loc)
	today := pacificNow.Format("2006-01-02") // YYYY-MM-DD format
	todayDate, _ := time.Parse("2006-01-02", today)

	// Check initial rate limit status before starting the main transaction
	var initialUserLimit models.UserRateLimit
	errRateLimit := database.DB.QueryRow(`
        SELECT email, clear_room_count, clear_room_date, is_blacklisted, blacklisted_at, blacklisted_reason
        FROM user_rate_limits WHERE email = $1`, emailStr).Scan(
		&initialUserLimit.Email, &initialUserLimit.ClearRoomCount, &initialUserLimit.ClearRoomDate,
		&initialUserLimit.IsBlacklisted, &initialUserLimit.BlacklistedAt, &initialUserLimit.BlacklistedReason,
	)

	if errRateLimit != nil {
		if errors.Is(errRateLimit, sql.ErrNoRows) {
			// No record yet, user is not blacklisted and count is 0. Will insert later if needed.
			initialUserLimit.Email = emailStr
			initialUserLimit.ClearRoomCount = 0
			initialUserLimit.IsBlacklisted = false
			initialUserLimit.ClearRoomDate.Valid = true // Assume we'll set today if record is created
			initialUserLimit.ClearRoomDate.Time = todayDate
			log.Printf("No existing rate limit record for %s.", emailStr)
		} else {
			log.Printf("Error checking initial rate limits for %s: %v", emailStr, errRateLimit)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check rate limits"})
			return
		}
	} else {
		// Record exists, check if blacklisted
		if initialUserLimit.IsBlacklisted {
			log.Printf("User %s is already blacklisted. Denying clear room request.", emailStr)
			c.JSON(http.StatusForbidden, gin.H{"error": "Your account is restricted due to previous activity. Please contact an administrator.", "blacklisted": true})
			return
		}

		// Check if date needs reset (do this *before* the main transaction if possible)
		recordDateStr := today // Default if not valid
		if initialUserLimit.ClearRoomDate.Valid {
			recordDateStr = initialUserLimit.ClearRoomDate.Time.Format("2006-01-02")
		}
		if recordDateStr != today {
			log.Printf("Rate limit date mismatch for %s (%s vs %s). Count will be reset.", emailStr, recordDateStr, today)
			// The update/insert logic within the transaction will handle the reset.
			initialUserLimit.ClearRoomCount = 0 // Reset count conceptually for pre-check
		}

		// Pre-check if already over limit (e.g., if MAX_DAILY_CLEARS was lowered)
		if initialUserLimit.ClearRoomCount >= MAX_DAILY_CLEARS {
			log.Printf("User %s already met or exceeded clear limit (%d) for today (%s). Denying clear room request.", emailStr, initialUserLimit.ClearRoomCount, today)
			// Optionally blacklist here, though the logic later will catch it too.
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "You have already reached the daily limit for clearing rooms.",
				"blacklisted": false, // Not necessarily blacklisted yet, just at limit
			})
			return
		}
	}

	// --- Transactional Logging: Get Previous State ---
	previousRoomState, err := getRoomStateRaw(roomUUIDParam)
	if err != nil { /* ... handle fetch error ... */
		return
	}
	if previousRoomState == nil { /* ... handle not found ... */
		return
	}

	// Create a notification queue for bump notifications
	notificationQueue := models.NewBumpNotificationQueue()

	// Start a transaction
	tx, err := database.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}

	var commitErr error
	var clearedOccupantsForLog models.IntArray = previousRoomState.Occupants
	var roomAlreadyEmpty bool // Flag to track if room was empty before clear

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Printf("PANIC during CLEAR_ROOM for %s by %s: %v", roomUUIDParam, emailStr, r)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error due to panic"})
			return
		}
		if err != nil {
			log.Printf("Rolling back transaction for CLEAR_ROOM %s by %s due to error: %v", roomUUIDParam, emailStr, err)
			tx.Rollback()
			return // Error response should have been sent
		}

		commitErr = tx.Commit()
		if commitErr != nil {
			log.Printf("Failed to commit transaction for CLEAR_ROOM %s by %s: %v", emailStr, roomUUIDParam, commitErr)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database transaction commit error"})
			return
		}

		// --- COMMIT SUCCEEDED ---
		log.Printf("Successfully committed CLEAR_ROOM for room %s by %s", roomUUIDParam, emailStr)

		// Send Bump Notifications
		for _, notification := range notificationQueue.Notifications {
			go SendBumpNotification(notification.UserID, notification.RoomID, notification.DormName)
		}

		// --- Transactional Logging: Get New State & Log ---
		newRoomState, fetchErr := getRoomStateRaw(roomUUIDParam)
		if fetchErr != nil {
			log.Printf("Error fetching new room state for CLEAR_ROOM %s: %v", roomUUIDParam, fetchErr)
		}

		logDetails := map[string]interface{}{
			"cleared_occupants": clearedOccupantsForLog,
		}
		loggingErr := logging.LogOperation(c, "CLEAR_ROOM", models.EntityTypeRoom, roomUUIDParam, previousRoomState, newRoomState, logDetails)
		if loggingErr != nil {
			log.Printf("WARNING: Failed to log CLEAR_ROOM operation for %s: %v", roomUUIDParam, loggingErr)
		}

		// --- Rate Limiting Post-Commit Fetch & Blacklist Logic ---
		var updatedUserLimit models.UserRateLimit
		// This fetch happens *after* the main transaction committed the count increment (if any)
		rateLimitFetchErr := database.DB.QueryRow(`
            SELECT email, clear_room_count, clear_room_date, is_blacklisted, blacklisted_at, blacklisted_reason
            FROM user_rate_limits WHERE email = $1`, emailStr).Scan(
			&updatedUserLimit.Email, &updatedUserLimit.ClearRoomCount, &updatedUserLimit.ClearRoomDate,
			&updatedUserLimit.IsBlacklisted, &updatedUserLimit.BlacklistedAt, &updatedUserLimit.BlacklistedReason,
		)

		if rateLimitFetchErr != nil {
			// This is problematic - the count was likely updated, but we can't confirm or check blacklist easily.
			log.Printf("CRITICAL: Error fetching updated rate limit for %s after successful clear: %v. Blacklist check skipped.", emailStr, rateLimitFetchErr)
			// Fallback: Use the initial count + 1 if the room wasn't empty? Less accurate.
			updatedUserLimit = initialUserLimit // Start with initial state
			if !roomAlreadyEmpty {
				updatedUserLimit.ClearRoomCount++ // Increment conceptually
			}
			// Cannot reliably check IsBlacklisted status here.
		} else {
			// Log the fetched updated count
			log.Printf("Fetched updated clear count for user %s: %d", emailStr, updatedUserLimit.ClearRoomCount)

			// Check if this operation pushed the user over the limit and blacklist them if so
			if updatedUserLimit.ClearRoomCount >= MAX_DAILY_CLEARS && !updatedUserLimit.IsBlacklisted {
				log.Printf("User %s reached clear limit (%d). Attempting to blacklist.", emailStr, updatedUserLimit.ClearRoomCount)
				// Start a new transaction specifically for blacklisting
				blacklistTx, btErr := database.DB.Begin()
				if btErr != nil {
					log.Printf("Error starting blacklist transaction for %s: %v", emailStr, btErr)
				} else {
					now := time.Now() // Use current time for blacklist timestamp
					reason := fmt.Sprintf("Exceeded daily clear room limit (%d) on %s", MAX_DAILY_CLEARS, today)
					_, execBlErr := blacklistTx.Exec(
						"UPDATE user_rate_limits SET is_blacklisted = true, blacklisted_at = $1, blacklisted_reason = $2 WHERE email = $3",
						now, reason, emailStr)
					if execBlErr != nil {
						log.Printf("Error executing blacklist update for %s: %v", emailStr, execBlErr)
						blacklistTx.Rollback()
					} else {
						blCommitErr := blacklistTx.Commit()
						if blCommitErr != nil {
							log.Printf("Error committing blacklist transaction for %s: %v", emailStr, blCommitErr)
						} else {
							updatedUserLimit.IsBlacklisted = true // Update local struct reflect change
							log.Printf("User %s successfully blacklisted.", emailStr)
							// TODO: Consider sending a blacklist notification email here?
						}
					}
				}
			}
		}

		// Calculate minutes until reset (uses pacificNow calculated earlier)
		midnight := time.Date(pacificNow.Year(), pacificNow.Month(), pacificNow.Day()+1, 0, 0, 0, 0, loc) // Start of *next* day in PT
		minutesUntilReset := int(time.Until(midnight).Minutes())

		// Send success response including rate limit info
		c.JSON(http.StatusOK, gin.H{
			"message":         fmt.Sprintf("Room %s cleared successfully", roomUUIDParam),
			"clearRoomCount":  updatedUserLimit.ClearRoomCount, // Use the fetched/calculated count
			"maxDailyClears":  MAX_DAILY_CLEARS,
			"remainingClears": max(0, MAX_DAILY_CLEARS-updatedUserLimit.ClearRoomCount), // Ensure non-negative
			"resetsInMinutes": minutesUntilReset,
			"pacificDate":     today,
			"isBlacklisted":   updatedUserLimit.IsBlacklisted, // Use fetched/updated status
		})

	}() // End of defer func

	// Get current room info to check if it can be cleared
	var currentRoomInfo models.RoomRaw
	err = tx.QueryRow("SELECT room_uuid, dorm, dorm_name, room_id, suite_uuid, max_occupancy, current_occupancy, occupants, pull_priority, has_frosh FROM rooms WHERE room_uuid = $1", roomUUIDParam).Scan(
		&currentRoomInfo.RoomUUID,
		&currentRoomInfo.Dorm,
		&currentRoomInfo.DormName,
		&currentRoomInfo.RoomID,
		&currentRoomInfo.SuiteUUID,
		&currentRoomInfo.MaxOccupancy,
		&currentRoomInfo.CurrentOccupancy,
		&currentRoomInfo.Occupants,
		&currentRoomInfo.PullPriority,
		&currentRoomInfo.HasFrosh,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query room info from rooms table"})
		return
	}

	// Check if the room has frosh - if so, don't allow clearing it
	if currentRoomInfo.HasFrosh {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot clear a room with frosh"})
		err = errors.New("room has frosh")
		return
	}

	// If the room is preplaced, don't allow clearing
	if currentRoomInfo.PullPriority.IsPreplaced {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot clear a preplaced room"})
		err = errors.New("room is preplaced")
		return
	}

	// Check if the room is already empty
	roomAlreadyEmpty = currentRoomInfo.CurrentOccupancy == 0
	log.Printf("Room %s current occupancy: %d (empty: %v)", roomUUIDParam, currentRoomInfo.CurrentOccupancy, roomAlreadyEmpty)

	// Clear the room
	err = clearRoom(currentRoomInfo.RoomUUID, tx, notificationQueue, emailStr)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear room"})
		return
	}

	// Only increment the clear count if the room wasn't already empty
	if !roomAlreadyEmpty {
		log.Printf("Incrementing clear count for user %s (room was not empty)", emailStr)

		// Increment the clear count within the same transaction
		_, err = tx.Exec("UPDATE user_rate_limits SET clear_room_count = clear_room_count + 1 WHERE email = $1", emailStr)
		if err != nil {
			log.Printf("Error updating clear count: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update clear count"})
			return
		}
	} else {
		log.Printf("Not incrementing clear count for user %s (room was already empty)", emailStr)
	}
}

// GetRoomsPagedAndSorted handles getting rooms with pagination, sorting and filtering
func GetRoomsPagedAndSorted(c *gin.Context) {
	// Start a transaction
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

	// Get pagination and sorting parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	sortBy := c.DefaultQuery("sort_by", "dorm_name")
	sortOrder := c.DefaultQuery("sort_order", "asc")
	dormValues := c.QueryArray("dorm")
	capacityValues := c.QueryArray("capacity")
	emptyOnly := c.DefaultQuery("empty_only", "false")

	// Validate parameters
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	// Calculate offset
	offset := (page - 1) * limit

	// Build the base query
	baseQuery := "SELECT room_uuid, dorm, dorm_name, room_id, suite_uuid, max_occupancy, current_occupancy, occupants, pull_priority, sgroup_uuid, has_frosh, frosh_room_type FROM rooms"

	// Build WHERE clause
	whereClause := ""
	args := []interface{}{}
	argCount := 1

	if len(dormValues) > 0 {
		if whereClause == "" {
			whereClause += " WHERE"
		} else {
			whereClause += " AND"
		}

		whereClause += " ("
		dormPlaceholders := make([]string, len(dormValues))
		for i, dorm := range dormValues {
			dormPlaceholders[i] = fmt.Sprintf("dorm_name = $%d", argCount)
			args = append(args, dorm)
			argCount++
		}
		whereClause += strings.Join(dormPlaceholders, " OR ")
		whereClause += ")"
	}

	if len(capacityValues) > 0 {
		if whereClause == "" {
			whereClause += " WHERE"
		} else {
			whereClause += " AND"
		}

		whereClause += " ("
		capacityPlaceholders := make([]string, 0, len(capacityValues))
		for _, capacityStr := range capacityValues {
			capacity, err := strconv.Atoi(capacityStr)
			if err == nil {
				capacityPlaceholders = append(capacityPlaceholders, fmt.Sprintf("max_occupancy = $%d", argCount))
				args = append(args, capacity)
				argCount++
			}
		}

		if len(capacityPlaceholders) > 0 {
			whereClause += strings.Join(capacityPlaceholders, " OR ")
		} else {
			whereClause += " FALSE" // If no valid capacities, add a condition that's always false
		}
		whereClause += ")"
	}

	if emptyOnly == "true" {
		if whereClause == "" {
			whereClause += " WHERE"
		} else {
			whereClause += " AND"
		}
		whereClause += " (current_occupancy < max_occupancy OR current_occupancy IS NULL)"
	}

	// Validate sort column to prevent SQL injection
	allowedSortColumns := map[string]string{
		"dorm":              "dorm",
		"dorm_name":         "dorm_name",
		"room_id":           "room_id",
		"max_occupancy":     "max_occupancy",
		"current_occupancy": "current_occupancy",
	}

	validSortColumn, exists := allowedSortColumns[sortBy]
	if !exists {
		validSortColumn = "dorm_name" // Default to dorm_name if invalid
	}

	// Validate sort order
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "asc" // Default to ascending if invalid
	}

	// Count total records query
	countQuery := "SELECT COUNT(*) FROM rooms" + whereClause
	var totalRecords int
	err = tx.QueryRow(countQuery, args...).Scan(&totalRecords)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database count query failed"})
		return
	}

	// Calculate total pages
	totalPages := (totalRecords + limit - 1) / limit

	// Build final query with sorting and pagination
	query := baseQuery + whereClause + " ORDER BY " + validSortColumn + " " + sortOrder +
		" LIMIT $" + strconv.Itoa(argCount) + " OFFSET $" + strconv.Itoa(argCount+1)
	args = append(args, limit, offset)

	// Execute the query
	rows, err := tx.Query(query, args...)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed"})
		return
	}
	defer rows.Close()

	var rooms []models.RoomRaw
	for rows.Next() {
		var room models.RoomRaw
		if err := rows.Scan(&room.RoomUUID, &room.Dorm, &room.DormName, &room.RoomID,
			&room.SuiteUUID, &room.MaxOccupancy, &room.CurrentOccupancy, &room.Occupants,
			&room.PullPriority, &room.SGroupUUID, &room.HasFrosh, &room.FroshRoomType); err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database scan failed"})
			return
		}
		rooms = append(rooms, room)
	}

	// Return the results
	c.JSON(http.StatusOK, gin.H{
		"rooms":       rooms,
		"page":        page,
		"limit":       limit,
		"total":       totalRecords,
		"total_pages": totalPages,
	})
}
