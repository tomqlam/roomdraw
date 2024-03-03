package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"roomdraw/backend/pkg/database"
	"roomdraw/backend/pkg/models"

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
	rows, err := tx.Query("SELECT room_uuid, dorm, dorm_name, room_id, suite_uuid, max_occupancy, current_occupancy, occupants, pull_priority, sgroup_uuid FROM rooms")
	if err != nil {
		// Handle query error
		// print the error to the console
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed"})
		return
	}

	var rooms []models.RoomRaw
	for rows.Next() {
		var d models.RoomRaw
		if err := rows.Scan(&d.RoomUUID, &d.Dorm, &d.DormName, &d.RoomID, &d.SuiteUUID, &d.MaxOccupancy, &d.CurrentOccupancy, &d.Occupants, &d.PullPriority, &d.SGroupUUID); err != nil {
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

	rows, err := tx.Query("SELECT room_uuid, dorm, dorm_name, room_id, suite_uuid, max_occupancy, current_occupancy, occupants, pull_priority, has_frosh FROM rooms WHERE UPPER(dorm_name) = UPPER($1)", dormNameParam)
	if err != nil {
		// Handle query error
		// print the error to the console
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed on rooms"})
		return
	}

	var rooms []models.RoomRaw
	for rows.Next() {
		var d models.RoomRaw
		if err := rows.Scan(&d.RoomUUID, &d.Dorm, &d.DormName, &d.RoomID, &d.SuiteUUID, &d.MaxOccupancy, &d.CurrentOccupancy, &d.Occupants, &d.PullPriority, &d.HasFrosh); err != nil {
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

	var suiteToRoomMap = make(map[string][]models.RoomSimple)
	for _, r := range rooms {
		suiteUUIDString := r.SuiteUUID.String()
		room := models.RoomSimple{
			RoomNumber:   r.RoomID,
			PullPriority: r.PullPriority,
			MaxOccupancy: r.MaxOccupancy,
			RoomUUID:     r.RoomUUID,
			HasFrosh:     r.HasFrosh,
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
			Rooms:           suiteToRoomMap[suiteUUIDString],
			SuiteDesign:     s.SuiteDesign,
			SuiteUUID:       s.SuiteUUID,
			AlternativePull: s.AlternativePull,
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
		// Handle query error
		// print the error to the console
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

func UpdateRoomOccupants(c *gin.Context) {
	// the room uuid is in the url
	roomUUIDParam := c.Param("roomuuid")

	var request models.OccupantUpdateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		log.Printf("JSON unmarshal error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	proposedOccupants := request.ProposedOccupants

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

	// check that the proposed occupants are not more than the max occupancy
	if len(proposedOccupants) > currentRoomInfo.MaxOccupancy {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Proposed occupants exceeds max occupancy"})
		return
	}

	occupantsAlreadyInRoom := []int{} // list of occupants already in the room

	// check that all of the proposed occupants are not already in a room
	for _, occupant := range proposedOccupants {
		var roomUUID uuid.UUID
		err = tx.QueryRow("SELECT room_uuid FROM users WHERE id = $1", occupant).Scan(&roomUUID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query room_uuid from users table"})
			return
		}

		if roomUUID != uuid.Nil {
			occupantsAlreadyInRoom = append(occupantsAlreadyInRoom, occupant)
		}
	}

	if len(occupantsAlreadyInRoom) > 0 {
		err = errors.New("one or more of the proposed occupants is already in a room")
		c.JSON(http.StatusBadRequest, gin.H{"error": "One or more of the proposed occupants is already in a room", "occupants": occupantsAlreadyInRoom})
		return
	}

	var proposedPullPriority models.PullPriority
	var pullLeaderPriority models.PullPriority
	var alternativeGroupPriority models.PullPriority
	var pullLeaderSuiteGroupUUID uuid.UUID
	log.Println(request.PullType)

	switch request.PullType {
	case 1: // self pull
		if len(proposedOccupants) > 0 {
			log.Println("Self pull")
			var occupantsInfo []models.UserRaw
			rows, err := tx.Query("SELECT id, draw_number, year, in_dorm FROM users WHERE id = ANY($1)", pq.Array(proposedOccupants))
			if err != nil {
				// Handle query error
				// print the error to the console
				log.Println(err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed on users for pull priority"})
				return
			}
			for rows.Next() {
				var u models.UserRaw
				if err := rows.Scan(&u.Id, &u.DrawNumber, &u.Year, &u.InDorm); err != nil {
					// Handle scan error
					log.Println(err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Database scan failed on users for pull priority"})
					return
				}
				occupantsInfo = append(occupantsInfo, u)
			}

			sortedOccupants := sortUsersByPriority(occupantsInfo, currentRoomInfo.Dorm)

			proposedPullPriority = generateUserPriority(sortedOccupants[0], currentRoomInfo.Dorm)

			proposedPullPriority.Valid = true
			proposedPullPriority.PullType = 1
		} else {
			err = clearRoom(currentRoomInfo.RoomUUID, tx)
			if err != nil {
				log.Println(err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear room"})
			}
			return
		}
	case 2: // normal pull
		if len(proposedOccupants) == 0 {
			// error because normal pull requires at least one occupant
			c.JSON(http.StatusBadRequest, gin.H{"error": "Normal pull requires at least one occupant"})
			err = errors.New("normal pull requires at least one occupant")
			tx.Rollback()
			return
		}

		if currentRoomInfo.RoomUUID == request.PullLeaderRoom {
			// error because the pull leader is already in the room
			c.JSON(http.StatusBadRequest, gin.H{"error": "Pull leader is already in the room"})
			err = errors.New("pull leader is already in the room")
			tx.Rollback()
			return
		}

		if currentRoomInfo.MaxOccupancy > 1 {
			// error because normal pull is not allowed for rooms with max occupancy > 1
			c.JSON(http.StatusBadRequest, gin.H{"error": "You may only initiate a normal pull for singles"})
			err = errors.New("normal pull is not allowed for rooms with max occupancy > 1")
			tx.Rollback()
			return
		} else if currentRoomInfo.MaxOccupancy == 1 {
			pullLeaderRoomUUID := request.PullLeaderRoom
			var occupantsInfo []models.UserRaw
			rows, err := tx.Query("SELECT id, draw_number, year, in_dorm FROM users WHERE id = ANY($1)", pq.Array(proposedOccupants))
			if err != nil {
				log.Println(err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed on users for pull priority"})
				tx.Rollback()
				return
			}

			for rows.Next() {
				var u models.UserRaw
				if err := rows.Scan(&u.Id, &u.DrawNumber, &u.Year, &u.InDorm); err != nil {
					log.Println(err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Database scan failed on users for pull priority"})
					tx.Rollback()
					return
				}
				occupantsInfo = append(occupantsInfo, u)
			}

			var suiteUUID = currentRoomInfo.SuiteUUID
			var leaderSuiteUUID uuid.UUID
			var pullLeaderCurrentOccupancy int

			// get the pull leader's priority
			err = tx.QueryRow("SELECT pull_priority, sgroup_uuid, suite_uuid, current_occupancy FROM rooms WHERE room_uuid = $1", pullLeaderRoomUUID).Scan(&pullLeaderPriority, &pullLeaderSuiteGroupUUID, &leaderSuiteUUID, &pullLeaderCurrentOccupancy)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query pull leader's priority from rooms table"})
				tx.Rollback()
				return
			}

			if leaderSuiteUUID != suiteUUID {
				// error because the pull leader is not in the same suite
				c.JSON(http.StatusBadRequest, gin.H{"error": "Pull leader is not in the same suite"})
				tx.Rollback()
				return
			}

			if pullLeaderCurrentOccupancy != 1 {
				// error because the pull leader is not in a single
				c.JSON(http.StatusBadRequest, gin.H{"error": "You can only initiate a normal pull with a single"})
				tx.Rollback()
				return
			}

			sortedOccupants := sortUsersByPriority(occupantsInfo, currentRoomInfo.Dorm)

			proposedPullPriority = generateUserPriority(sortedOccupants[0], currentRoomInfo.Dorm)
			proposedPullPriority.Valid = true
			proposedPullPriority.PullType = 2

			log.Println(proposedPullPriority)
			log.Println(pullLeaderPriority)

			if !comparePullPriority(pullLeaderPriority, proposedPullPriority) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Pull leader does not have higher priority than proposed occupants"})
				tx.Rollback()
				return
			}

			proposedPullPriority.Inherited.Valid = true
			proposedPullPriority.Inherited.DrawNumber = pullLeaderPriority.DrawNumber
			proposedPullPriority.Inherited.HasInDorm = pullLeaderPriority.HasInDorm
			proposedPullPriority.Inherited.Year = pullLeaderPriority.Year
		}
	case 3: // lock pull
		// can only lock pull info an empty room
		if currentRoomInfo.CurrentOccupancy > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Lock pull is only allowed for empty rooms"})
			err = errors.New("lock pull is only allowed for empty rooms")
			tx.Rollback()
			return
		}

		if len(proposedOccupants) == 0 {
			// error because lock pull requires at least one occupant
			c.JSON(http.StatusBadRequest, gin.H{"error": "Lock pull requires at least one occupant"})
			err = errors.New("lock pull requires at least one occupant")
			tx.Rollback()
			return
		}

		// lock pulled room must be full
		if len(proposedOccupants) != currentRoomInfo.MaxOccupancy {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Lock pull requires the room to be full"})
			err = errors.New("lock pull requires the room to be full")
			tx.Rollback()
			return
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
		}

		// query all the rooms and ensure that they are full
		var roomsInSuite []models.RoomRaw

		rows, err := tx.Query("SELECT room_uuid, dorm, dorm_name, room_id, suite_uuid, max_occupancy, current_occupancy, occupants, pull_priority, sgroup_uuid FROM rooms WHERE suite_uuid = $1", currentRoomInfo.SuiteUUID)
		if err != nil {
			// Handle query error
			// print the error to the console
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed on rooms for pull priority"})
			tx.Rollback()
			return
		}

		for rows.Next() {
			var r models.RoomRaw
			if err := rows.Scan(&r.RoomUUID, &r.Dorm, &r.DormName, &r.RoomID, &r.SuiteUUID, &r.MaxOccupancy, &r.CurrentOccupancy, &r.Occupants, &r.PullPriority, &r.SGroupUUID); err != nil {
				// Handle scan error
				log.Println(err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Database scan failed on rooms for pull priority"})
				tx.Rollback()
				return
			}
			roomsInSuite = append(roomsInSuite, r)
		}

		for _, roomInSuite := range roomsInSuite {
			if roomInSuite.CurrentOccupancy < roomInSuite.MaxOccupancy && roomInSuite.RoomUUID != currentRoomInfo.RoomUUID {
				c.JSON(http.StatusBadRequest, gin.H{"error": "One or more rooms in the suite are not full"})
				tx.Rollback()
				return
			}
		}

		var occupantsInfo []models.UserRaw
		rows, err = tx.Query("SELECT id, draw_number, year, in_dorm FROM users WHERE id = ANY($1)", pq.Array(proposedOccupants))
		if err != nil {
			// Handle query error
			// print the error to the console
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed on users for pull priority"})
			return
		}
		for rows.Next() {
			var u models.UserRaw
			if err := rows.Scan(&u.Id, &u.DrawNumber, &u.Year, &u.InDorm); err != nil {
				// Handle scan error
				log.Println(err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Database scan failed on users for pull priority"})
				return
			}
			occupantsInfo = append(occupantsInfo, u)
		}

		sortedOccupants := sortUsersByPriority(occupantsInfo, currentRoomInfo.Dorm)

		proposedPullPriority = generateUserPriority(sortedOccupants[0], currentRoomInfo.Dorm)

		proposedPullPriority.Valid = true
		proposedPullPriority.PullType = 3
		proposedPullPriority.Inherited.Valid = true
	case 4: // alternative pull
		if len(proposedOccupants) != currentRoomInfo.MaxOccupancy {
			// error because normal pull requires a full room
			c.JSON(http.StatusBadRequest, gin.H{"error": "Alternative pull requires the room to be full"})
			tx.Rollback()
			return
		}

		if currentRoomInfo.RoomUUID == request.PullLeaderRoom {
			// error because the pull leader is already in the room
			c.JSON(http.StatusBadRequest, gin.H{"error": "Pull leader is already in the room"})
			err = errors.New("pull leader is already in the room")
			tx.Rollback()
			return
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
			return
		}

		// ensure that alternative pull is allowed for the suite
		if !suiteInfo.AlternativePull {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Alternative pull is not allowed for the suite"})
			tx.Rollback()
			return
		}

		pullLeaderRoomUUID := request.PullLeaderRoom
		var occupantsInfo []models.UserRaw
		rows, err := tx.Query("SELECT id, draw_number, year, in_dorm FROM users WHERE id = ANY($1)", pq.Array(proposedOccupants))
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed on users for pull priority"})
			tx.Rollback()
			return
		}

		for rows.Next() {
			var u models.UserRaw
			if err := rows.Scan(&u.Id, &u.DrawNumber, &u.Year, &u.InDorm); err != nil {
				log.Println(err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Database scan failed on users for pull priority"})
				tx.Rollback()
				return
			}
			occupantsInfo = append(occupantsInfo, u)
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
			return
		}

		if leaderSuiteUUID != suiteUUID {
			// error because the pull leader is not in the same suite
			c.JSON(http.StatusBadRequest, gin.H{"error": "Pull leader is not in the same suite"})
			tx.Rollback()
			return
		}

		if pullLeaderCurrentOccupancy != pullLeaderMaxOccupancy {
			log.Println(pullLeaderCurrentOccupancy, pullLeaderMaxOccupancy)
			// error because the pull leader is not in a single
			c.JSON(http.StatusBadRequest, gin.H{"error": "You can only initiate an alternative pull with a full room"})
			tx.Rollback()
			return
		}

		// get all of the users in the pull leader's room
		var pullLeaderOccupantsInfo []models.UserRaw
		rows, err = tx.Query("SELECT id, draw_number, year, in_dorm FROM users WHERE room_uuid = $1", pullLeaderRoomUUID)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed on users for pull priority"})
			tx.Rollback()
			return
		}

		for rows.Next() {
			var u models.UserRaw
			if err := rows.Scan(&u.Id, &u.DrawNumber, &u.Year, &u.InDorm); err != nil {
				log.Println(err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Database scan failed on users for pull priority"})
				tx.Rollback()
				return
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

	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pull type: " + string(rune(request.PullType))})
		err = errors.New("invalid pull type")
		return
	}

	log.Println(proposedPullPriority)
	log.Print(currentRoomInfo.PullPriority)

	if !comparePullPriority(proposedPullPriority, currentRoomInfo.PullPriority) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Proposed occupants do not have higher priority than current occupants"})
		err = errors.New("proposed occupants do not have higher priority than current occupants")
		tx.Rollback()
		return
	}

	log.Println(proposedOccupants)

	// disband the suite group if there is one
	if currentRoomInfo.SGroupUUID != uuid.Nil {
		_, err := disbandSuiteGroup(currentRoomInfo.SGroupUUID, tx)
		if err != nil {
			// use err in the response
			c.JSON(http.StatusInternalServerError, err.Error())
			return
		}
	}

	if currentRoomInfo.CurrentOccupancy > 0 {
		// remove the current occupants from the room
		_, err = tx.Exec("UPDATE rooms SET occupants = $1, current_occupancy = $2 WHERE room_uuid = $3", nil, 0, roomUUIDParam)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update current occupants in rooms table"})
			return
		}

		// for each current occupant, nullify the room_uuid field in the users table
		_, err = tx.Exec("UPDATE users SET room_uuid = $1 WHERE room_uuid = $2", nil, roomUUIDParam)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update room_uuid in users table"})
			return
		}
	}

	// update the occupants in the database and the current_occupancy
	_, err = tx.Exec("UPDATE rooms SET occupants = $1, current_occupancy = $2 WHERE room_uuid = $3", pq.Array(proposedOccupants), len(proposedOccupants), roomUUIDParam)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// for each occupant, update the room_uuid field in the users table
	for _, proposedOccupant := range proposedOccupants {
		_, err = tx.Exec("UPDATE users SET room_uuid = $1 WHERE id = $2", roomUUIDParam, proposedOccupant)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update room_uuid in users table"})
			return
		}
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

	if request.PullType == 2 {
		if pullLeaderSuiteGroupUUID == uuid.Nil {
			log.Println("Pull leader is not in a suite group")
			// create new suite group with the pull leader's priority
			pullLeaderPriorityJSON, err := json.Marshal(pullLeaderPriority)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal pull leader's pull priority"})
				return
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
				return
			}

			// update the sgroup_uuid field in the rooms table for both rooms
			_, err = tx.Exec("UPDATE rooms SET sgroup_uuid = $1 WHERE room_uuid = $2 OR room_uuid = $3", suiteGroupUUID, roomUUIDParam, request.PullLeaderRoom)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update suite_uuid in rooms table"})
				return
			}

			// update the sgroup_uuid field in the users table for all occupants of both rooms
			_, err = tx.Exec("UPDATE users SET sgroup_uuid = $1 WHERE room_uuid = $2 OR room_uuid = $3", suiteGroupUUID, roomUUIDParam, request.PullLeaderRoom)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update sgroup_uuid in users table"})
				return
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
				return
			}

			if suiteInfo.Dorm != 3 {
				if suiteInfo.RoomCount != 3 {
					c.JSON(http.StatusBadRequest, gin.H{"error": "You can only pull two suitemates in South in a suite with three rooms"})
					err = errors.New("you can only pull two suitemates in South in a suite with three rooms")
					tx.Rollback()
					return
				}

				c.JSON(http.StatusBadRequest, gin.H{"error": "You can only pull two suitemates in South"})
				err = errors.New("you can only pull two suitemates in South")
				tx.Rollback()
				return
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
				return
			}

			// add the room to the suite group
			_, err = tx.Exec("UPDATE suitegroups SET rooms = array_append(rooms, $1) WHERE sgroup_uuid = $2", roomUUIDParam, pullLeaderSuiteGroupUUID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update rooms in suitegroups table"})
				return
			}

			// update the sgroup_uuid field in the rooms table
			_, err = tx.Exec("UPDATE rooms SET sgroup_uuid = $1 WHERE room_uuid = $2", pullLeaderSuiteGroupUUID, roomUUIDParam)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update suite_uuid in rooms table"})
				return
			}

			// update the sgroup_uuid field in the users table for all occupants of the room
			_, err = tx.Exec("UPDATE users SET sgroup_uuid = $1 WHERE room_uuid = $2", pullLeaderSuiteGroupUUID, roomUUIDParam)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update sgroup_uuid in users table"})
				return
			}

		}
	} else if request.PullType == 3 {
		// update the suite's lock pull status
		_, err = tx.Exec("UPDATE suites SET lock_pulled_room = $1 WHERE suite_uuid = $2", roomUUIDParam, currentRoomInfo.SuiteUUID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update lock_pulled_room in suites table"})
			return
		}
	} else if request.PullType == 4 {
		if pullLeaderSuiteGroupUUID != uuid.Nil {
			// error out because the pull leader is in a suite group
			c.JSON(http.StatusBadRequest, gin.H{"error": "Pull leader is in a suite group for alternative pull"})
			err = errors.New("pull leader is in a suite group for alternative pull")
			tx.Rollback()
			return
		}

		// do the same thing as for pull type 2
		// create new suite group with the pull leader's priority
		alternativeGroupPriorityJSON, err := json.Marshal(alternativeGroupPriority)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal pull leader's pull priority"})
			return
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
			return
		}

		// update the sgroup_uuid field in the rooms table for both rooms
		_, err = tx.Exec("UPDATE rooms SET sgroup_uuid = $1 WHERE room_uuid = $2 OR room_uuid = $3", suiteGroupUUID, roomUUIDParam, request.PullLeaderRoom)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update suite_uuid in rooms table"})
			return
		}

		// update the sgroup_uuid field in the users table for all occupants of both rooms
		_, err = tx.Exec("UPDATE users SET sgroup_uuid = $1 WHERE room_uuid = $2 OR room_uuid = $3", suiteGroupUUID, roomUUIDParam, request.PullLeaderRoom)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update sgroup_uuid in users table"})
			return
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
			return
		}

		log.Println("Pull leader room: " + request.PullLeaderRoom.String())
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully updated occupants"})
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
