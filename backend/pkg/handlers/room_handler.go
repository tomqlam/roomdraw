package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"roomdraw/backend/pkg/database"
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

func ToggleInDorm(c *gin.Context) {
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

	userFullName, exists := c.Get("user_full_name")
	if !exists {
		log.Print("Error: user_full_name not found in context")
		userFullName = "unknown user"
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

	// Ensure the transaction is either committed or rolled back
	defer func() {
		if r := recover(); r != nil {
			log.Println("Result for " + userFullName.(string) + ": failed to Self Pull room " + roomUUIDParam + " because of panic " + r.(error).Error())
			tx.Rollback()
			panic(r)
		} else if err != nil {
			log.Println("Result for " + userFullName.(string) + ": failed to Self Pull room " + roomUUIDParam + " because of error " + err.Error())
			tx.Rollback()
		} else {
			log.Println("Result for " + userFullName.(string) + ": successfully Self Pulled room " + roomUUIDParam)
			err = tx.Commit()

			if err == nil {
				for _, notification := range notificationQueue.Notifications {
					log.Println("Notifying bumped users ...")
					SendBumpNotification(notification.UserID, notification.RoomID, notification.DormName)
				}
			}
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

	c.JSON(http.StatusOK, gin.H{"message": "Successfully updated occupants"})
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

	// Create notification queue
	notificationQueue := models.NewBumpNotificationQueue()

	// Start a transaction
	tx, err := database.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return err
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
			// Send notifications only after successful commit
			if err == nil {
				for _, notification := range notificationQueue.Notifications {
					SendBumpNotification(notification.UserID, notification.RoomID, notification.DormName)
				}
			}
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

	for _, occupant := range sortedOccupants {
		var pullLeaderEffectiveInDorm bool

		// accounts for in dorm forfeit
		if pullLeaderPriority.Inherited.Valid {
			pullLeaderEffectiveInDorm = pullLeaderPriority.Inherited.HasInDorm
		} else {
			pullLeaderEffectiveInDorm = pullLeaderPriority.HasInDorm
		}

		// if the pull leader has indorm and the proposed occupants do not, it is invalid
		if pullLeaderEffectiveInDorm && !(generateUserPriority(occupant, currentRoomInfo.Dorm).HasInDorm) {
			log.Println("Pull leader has in dorm and proposed occupants do not")
			err = errors.New("pull leader has in dorm and proposed occupants do not")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Pull leader has in dorm and proposed occupants do not"})
			return err
		}
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

	c.JSON(http.StatusOK, gin.H{"message": "Successfully updated occupants"})
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

	// Create notification queue
	notificationQueue := models.NewBumpNotificationQueue()

	// Start a transaction
	tx, err := database.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return err
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
			// Send notifications only after successful commit
			if err == nil {
				for _, notification := range notificationQueue.Notifications {
					SendBumpNotification(notification.UserID, notification.RoomID, notification.DormName)
				}
			}
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

	c.JSON(http.StatusOK, gin.H{"message": "Successfully updated occupants"})
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

	// Ensure the transaction is either committed or rolled back
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
			// Send notifications only after successful commit
			if err == nil {
				for _, notification := range notificationQueue.Notifications {
					SendBumpNotification(notification.UserID, notification.RoomID, notification.DormName)
				}
			}
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

	c.JSON(http.StatusOK, gin.H{"message": "Successfully updated occupants"})
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

	// Check rate limits for clear room operations
	const MAX_DAILY_CLEARS = 10

	// Get the current date in Pacific Time
	loc, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		log.Printf("Error loading Pacific timezone: %v", err)
		// Fall back to UTC if unable to load Pacific time
		loc = time.UTC
	}
	pacificNow := time.Now().In(loc)
	today := pacificNow.Format("2006-01-02") // YYYY-MM-DD format
	todayDate, _ := time.Parse("2006-01-02", today)

	// Check if the user has a rate limit record for today
	var userLimit models.UserRateLimit
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
			// First clear for this user, create a new record
			_, err = database.DB.Exec("INSERT INTO user_rate_limits (email, clear_room_count, clear_room_date) VALUES ($1, 0, $2)", emailStr, today)
			if err != nil {
				log.Printf("Error creating rate limit record: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
				return
			}
			userLimit.Email = emailStr
			userLimit.ClearRoomCount = 0
			userLimit.ClearRoomDate.Valid = true
			userLimit.ClearRoomDate.Time = todayDate
			userLimit.IsBlacklisted = false
		} else {
			log.Printf("Error checking rate limits: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}
	}

	// Check if we need to reset the counter (if date has changed)
	recordDateStr := today // Default to today if ClearRoomDate is not valid
	if userLimit.ClearRoomDate.Valid {
		recordDateStr = userLimit.ClearRoomDate.Time.Format("2006-01-02")
	}

	if recordDateStr != today {
		// Reset the counter for a new day
		_, err = database.DB.Exec("UPDATE user_rate_limits SET clear_room_count = 0, clear_room_date = $1 WHERE email = $2",
			today, emailStr)
		if err != nil {
			log.Printf("Error resetting rate limit: %v", err)
		}
		userLimit.ClearRoomCount = 0
	}

	// Check if the user is already at or has exceeded their limit
	if userLimit.ClearRoomCount >= MAX_DAILY_CLEARS {
		// Blacklist the user
		now := time.Now()
		reason := "Exceeded daily clear room limit"
		_, err = database.DB.Exec(
			"UPDATE user_rate_limits SET is_blacklisted = true, blacklisted_at = $1, blacklisted_reason = $2 WHERE email = $3",
			now, reason, emailStr)
		if err != nil {
			log.Printf("Error blacklisting user: %v", err)
		}

		log.Printf("User %s has been blacklisted for exceeding clear room rate limit", emailStr)
		c.JSON(http.StatusTooManyRequests, gin.H{
			"error":       "You have exceeded the daily limit for clearing rooms. Your account has been temporarily restricted. Please contact an administrator.",
			"blacklisted": true,
		})
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

	// Ensure the transaction is either committed or rolled back
	defer func() {
		if r := recover(); r != nil {
			log.Println("Result for " + userFullName.(string) + ": failed to clear room " + roomUUIDParam + " because of panic")
			tx.Rollback()
			panic(r)
		} else if err != nil {
			log.Println("Result for " + userFullName.(string) + ": failed to clear room " + roomUUIDParam + " because of error " + err.Error())
			tx.Rollback()
		} else {
			log.Println("Result for " + userFullName.(string) + ": successfully cleared room " + roomUUIDParam)
			err = tx.Commit()

			if err == nil {
				// Send notifications if needed
				for _, notification := range notificationQueue.Notifications {
					log.Println("Notifying bumped users ...")
					SendBumpNotification(notification.UserID, notification.RoomID, notification.DormName)
				}

				// Fetch the updated count from the database using the model
				var updatedUserLimit models.UserRateLimit
				err = database.DB.QueryRow(`
					SELECT email, clear_room_count, clear_room_date, is_blacklisted, blacklisted_at, blacklisted_reason 
					FROM user_rate_limits 
					WHERE email = $1
				`, emailStr).Scan(
					&updatedUserLimit.Email,
					&updatedUserLimit.ClearRoomCount,
					&updatedUserLimit.ClearRoomDate,
					&updatedUserLimit.IsBlacklisted,
					&updatedUserLimit.BlacklistedAt,
					&updatedUserLimit.BlacklistedReason,
				)

				if err != nil {
					log.Printf("Error fetching updated clear count: %v", err)
					// If room was already empty, keep the same count, otherwise increment
					updatedUserLimit.ClearRoomCount = userLimit.ClearRoomCount
					// Define roomAlreadyEmpty by checking if room occupancy is 0
					roomAlreadyEmpty := false
					var currentOccupancy int
					err = database.DB.QueryRow("SELECT current_occupancy FROM rooms WHERE room_uuid = $1", roomUUIDParam).Scan(&currentOccupancy)
					if err == nil {
						roomAlreadyEmpty = currentOccupancy == 0
					}
					if !roomAlreadyEmpty {
						updatedUserLimit.ClearRoomCount += 1 // Fallback to calculation if fetch fails
					}
				}

				// Log the updated values
				log.Printf("Updated clear count for user %s: %d", emailStr, updatedUserLimit.ClearRoomCount)

				// Check if this operation pushed the user over the limit and blacklist them if so
				if updatedUserLimit.ClearRoomCount >= MAX_DAILY_CLEARS && !updatedUserLimit.IsBlacklisted {
					// Start a new transaction for blacklisting
					blacklistTx, err := database.DB.Begin()
					if err != nil {
						log.Printf("Error starting blacklist transaction: %v", err)
					} else {
						now := time.Now()
						reason := "Exceeded daily clear room limit"
						_, err = blacklistTx.Exec(
							"UPDATE user_rate_limits SET is_blacklisted = true, blacklisted_at = $1, blacklisted_reason = $2 WHERE email = $3",
							now, reason, emailStr)
						if err != nil {
							log.Printf("Error blacklisting user: %v", err)
							blacklistTx.Rollback()
						} else {
							err = blacklistTx.Commit()
							if err != nil {
								log.Printf("Error committing blacklist transaction: %v", err)
							} else {
								updatedUserLimit.IsBlacklisted = true
								log.Printf("User %s has been blacklisted after reaching clear room limit", emailStr)
							}
						}
					}
				}

				// Calculate minutes until midnight Pacific Time
				midnight := time.Date(pacificNow.Year(), pacificNow.Month(), pacificNow.Day(), 0, 0, 0, 0, loc)
				if pacificNow.After(midnight) {
					midnight = midnight.Add(24 * time.Hour)
				}
				minutesUntilReset := int(midnight.Sub(pacificNow).Minutes())

				// Return the updated count to the client
				c.JSON(http.StatusOK, gin.H{
					"message":         "Successfully cleared room",
					"clearRoomCount":  updatedUserLimit.ClearRoomCount,
					"maxDailyClears":  MAX_DAILY_CLEARS,
					"remainingClears": MAX_DAILY_CLEARS - updatedUserLimit.ClearRoomCount,
					"resetsInMinutes": minutesUntilReset,
					"pacificDate":     today,
					"isBlacklisted":   updatedUserLimit.IsBlacklisted,
				})
			}
		}
	}()

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
	roomAlreadyEmpty := currentRoomInfo.CurrentOccupancy == 0
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
