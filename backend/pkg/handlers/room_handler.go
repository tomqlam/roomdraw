package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"roomdraw/backend/pkg/database"
	"roomdraw/backend/pkg/models"
	"strconv"
	"strings"
	"sync"

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

	rows, err = tx.Query("SELECT suite_uuid, dorm, dorm_name, floor, room_count, rooms, alternative_pull, suite_design, can_lock_pull, reslife_room, gender_preference FROM suites WHERE UPPER(dorm_name) = UPPER($1)", dormNameParam)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed on suites"})
		return
	}

	var suites []models.SuiteRaw
	for rows.Next() {
		var s models.SuiteRaw
		if err := rows.Scan(&s.SuiteUUID, &s.Dorm, &s.DormName, &s.Floor, &s.RoomCount, &s.Rooms, &s.AlternativePull, &s.SuiteDesign, &s.CanLockPull, &s.ReslifeRoom, &s.GenderPreference); err != nil {
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
			Rooms:            suiteToRoomMap[suiteUUIDString],
			SuiteDesign:      s.SuiteDesign,
			SuiteUUID:        s.SuiteUUID,
			AlternativePull:  s.AlternativePull,
			CanLockPull:      s.CanLockPull,
			GenderPreference: s.GenderPreference,
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
	var err error
	// Retrieve the doneChan from the context
	doneChanInterface, exists := c.Get("doneChan")
	if !exists {
		// If for some reason it doesn't exist, log an error and return
		log.Print("Error: doneChan not found in context")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// Assert the type of doneChan to be a chan bool
	doneChan, ok := doneChanInterface.(chan bool)
	if !ok {
		// If the assertion fails, log an error and return
		log.Print("Error: doneChan is not of type chan bool")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// Retrieve the closeOnce from the context
	closeOnceInterface, exists := c.Get("closeOnce")
	if !exists {
		// If for some reason it doesn't exist, log an error and return
		log.Print("Error: closeOnce not found in context")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// Assert the type of closeOnce to be a *sync.Once
	closeOnce, ok := closeOnceInterface.(*sync.Once)
	if !ok {
		// If the assertion fails, log an error and return
		log.Print("Error: closeOnce is not of type *sync.Once")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// Ensure that a signal is sent to doneChan when the function exits, make sure this happens only once
	defer func() {
		closeOnce.Do(func() {
			close(doneChan)
			log.Println("Closed doneChan for request")
		})
	}()

	// constantly listen for the doneChan to be closed (meaning the request was timed out) and return error
	go func() {
		<-doneChan
		log.Println("Request was fulfilled or timed out")
		// write to global error variable
		err = errors.New("request was fulfilled or timed out")
	}()

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
	var err error
	// Retrieve the doneChan from the context
	doneChanInterface, exists := c.Get("doneChan")
	if !exists {
		// If for some reason it doesn't exist, log an error and return
		log.Print("Error: doneChan not found in context")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// Assert the type of doneChan to be a chan bool
	doneChan, ok := doneChanInterface.(chan bool)
	if !ok {
		// If the assertion fails, log an error and return
		log.Print("Error: doneChan is not of type chan bool")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// Retrieve the closeOnce from the context
	closeOnceInterface, exists := c.Get("closeOnce")
	if !exists {
		// If for some reason it doesn't exist, log an error and return
		log.Print("Error: closeOnce not found in context")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// Assert the type of closeOnce to be a *sync.Once
	closeOnce, ok := closeOnceInterface.(*sync.Once)
	if !ok {
		// If the assertion fails, log an error and return
		log.Print("Error: closeOnce is not of type *sync.Once")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// Ensure that a signal is sent to doneChan when the function exits, make sure this happens only once
	defer func() {
		closeOnce.Do(func() {
			close(doneChan)
			log.Println("Closed doneChan for request")
		})
	}()

	// constantly listen for the doneChan to be closed (meaning the request was timed out) and return error
	go func() {
		<-doneChan
		log.Println("Request was fulfilled or timed out")
		// write to global error variable
		err = errors.New("request was fulfilled or timed out")
	}()

	var request models.OccupantUpdateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		log.Printf("JSON unmarshal error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

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
		err = clearRoom(currentRoomInfo.RoomUUID, tx)
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

	sortedOccupants := sortUsersByPriority(occupantsInfo, currentRoomInfo.Dorm)

	proposedPullPriority = generateUserPriority(sortedOccupants[0], currentRoomInfo.Dorm)

	// if the proposed pull priority has in dorm
	if proposedPullPriority.HasInDorm {
		// check if all users have in dorm
		for _, occupant := range sortedOccupants {
			if occupant.InDorm != currentRoomInfo.Dorm {
				// if not, set the proposed pull priority to have in dorm as false
				log.Println("Forfeited in dorm to pull non-in dorm user")
				proposedPullPriority.HasInDorm = false
				break
			}
		}
	}

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
		err = clearRoom(currentRoomInfo.RoomUUID, tx)
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

	// Start a transaction
	tx, err := database.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return err
	}

	// Ensure the transaction is either committed or rolled back
	defer func() {
		if r := recover(); r != nil {
			log.Println("Result for " + userFullName.(string) + ": failed to Normal Pull room " + roomUUIDParam + " because of panic " + r.(error).Error())
			tx.Rollback()
			panic(r)
		} else if err != nil {
			log.Println("Result for " + userFullName.(string) + ": failed to Normal Pull room " + roomUUIDParam + " because of error " + err.Error())
			tx.Rollback()
		} else {
			log.Println("Result for " + userFullName.(string) + ": successfully Normal Pulled room " + roomUUIDParam)
			err = tx.Commit()
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
		err = clearRoom(currentRoomInfo.RoomUUID, tx)
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

	// Start a transaction
	tx, err := database.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return err
	}

	// Ensure the transaction is either committed or rolled back
	defer func() {
		if r := recover(); r != nil {
			log.Println("Result for " + userFullName.(string) + ": failed to Lock Pull room " + roomUUIDParam + " because of panic " + r.(error).Error())
			tx.Rollback()
			panic(r)
		} else if err != nil {
			log.Println("Result for " + userFullName.(string) + ": failed to Lock Pull room " + roomUUIDParam + " because of error " + err.Error())
			tx.Rollback()
		} else {
			log.Println("Result for " + userFullName.(string) + ": successfully Lock Pulled room " + roomUUIDParam)
			err = tx.Commit()
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
		// error because lock pull requires at least one occupant
		c.JSON(http.StatusBadRequest, gin.H{"error": "Lock pull requires at least one occupant"})
		err = errors.New("lock pull requires at least one occupant")
		tx.Rollback()
		return err
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
		err = clearRoom(currentRoomInfo.RoomUUID, tx)
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

	// update the suite's lock pull status
	_, err = tx.Exec("UPDATE suites SET lock_pulled_room = $1 WHERE suite_uuid = $2", roomUUIDParam, currentRoomInfo.SuiteUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update lock_pulled_room in suites table"})
		return err
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

	// Start a transaction
	tx, err := database.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return err
	}

	// Ensure the transaction is either committed or rolled back
	defer func() {
		if r := recover(); r != nil {
			log.Println("Result for " + userFullName.(string) + ": failed to Alternative Pull room " + roomUUIDParam + " because of panic " + r.(error).Error())
			tx.Rollback()
			panic(r)
		} else if err != nil {
			log.Println("Result for " + userFullName.(string) + ": failed to Alternative Pull room " + roomUUIDParam + " because of error " + err.Error())
			tx.Rollback()
		} else {
			log.Println("Result for " + userFullName.(string) + ": successfully Alternative Pulled room " + roomUUIDParam)
			err = tx.Commit()
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

	var allOccupantsInfo = append(occupantsInfo, pullLeaderOccupantsInfo...)
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
		err = clearRoom(currentRoomInfo.RoomUUID, tx)
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

	c.JSON(http.StatusOK, gin.H{"message": "Successfully updated occupants"})
	return nil
}

func clearRoom(roomUUID uuid.UUID, tx *sql.Tx) error {
	// get the suite group uuid
	var suiteGroupUUID uuid.UUID
	err := tx.QueryRow("SELECT sgroup_uuid FROM rooms WHERE room_uuid = $1", roomUUID).Scan(&suiteGroupUUID)
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
	var err error
	// Retrieve the doneChan from the context
	doneChanInterface, exists := c.Get("doneChan")
	if !exists {
		// If for some reason it doesn't exist, log an error and return
		log.Print("Error: doneChan not found in context")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// Assert the type of doneChan to be a chan bool
	doneChan, ok := doneChanInterface.(chan bool)
	if !ok {
		// If the assertion fails, log an error and return
		log.Print("Error: doneChan is not of type chan bool")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// Retrieve the closeOnce from the context
	closeOnceInterface, exists := c.Get("closeOnce")
	if !exists {
		// If for some reason it doesn't exist, log an error and return
		log.Print("Error: closeOnce not found in context")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// Assert the type of closeOnce to be a *sync.Once
	closeOnce, ok := closeOnceInterface.(*sync.Once)
	if !ok {
		// If the assertion fails, log an error and return
		log.Print("Error: closeOnce is not of type *sync.Once")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// Ensure that a signal is sent to doneChan when the function exits, make sure this happens only once
	defer func() {
		closeOnce.Do(func() {
			close(doneChan)
			log.Println("Closed doneChan for request")
		})
	}()

	// constantly listen for the doneChan to be closed (meaning the request was timed out) and return error
	go func() {
		<-doneChan
		log.Println("Request was fulfilled or timed out")
		// write to global error variable
		err = errors.New("request was fulfilled or timed out")
	}()

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
	if err = c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

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

	// if proposed occupants is empty, clear the room
	if len(request.ProposedOccupants) == 0 {
		err = clearRoom(currentRoomInfo.RoomUUID, tx)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove the current occupants of the room"})
			return
		}
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

	var occupantsAlreadyInRoom models.IntArray
	rows, err := tx.Query("SELECT id FROM users WHERE id = ANY($1) AND room_uuid IS NOT NULL", pq.Array(request.ProposedOccupants))
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
