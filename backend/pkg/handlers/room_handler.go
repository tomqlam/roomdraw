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
	db, err := database.NewDatabase()
	if err != nil {
		// Handle error opening the database
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to connect to database"})
		return
	}
	defer db.Close()

	// Example SQL query
	rows, err := db.Query("SELECT room_uuid, dorm, dorm_name, room_id, suite_uuid, max_occupancy, current_occupancy, occupants, pull_priority, sgroup_uuid FROM rooms")
	if err != nil {
		// Handle query error
		// print the error to the console
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed"})
		return
	}
	defer rows.Close()

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

	db, err := database.NewDatabase()
	if err != nil {
		// Handle error opening the database
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to connect to database"})
		return
	}

	// Start a transaction
	tx, err := db.Begin()
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
		db.Close()
	}()

	rows, err := db.Query("SELECT room_uuid, dorm, dorm_name, room_id, suite_uuid, max_occupancy, current_occupancy, occupants, pull_priority FROM rooms WHERE UPPER(dorm_name) = UPPER($1)", dormNameParam)
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

	rows, err = db.Query("SELECT suite_uuid, dorm, dorm_name, floor, room_count, rooms, alternative_pull, suite_design FROM suites WHERE UPPER(dorm_name) = UPPER($1)", dormNameParam)
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
			Rooms:       suiteToRoomMap[suiteUUIDString],
			SuiteDesign: s.SuiteDesign,
			SuiteUUID:   s.SuiteUUID,
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

	db, err := database.NewDatabase()
	if err != nil {
		// Handle error opening the database
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to connect to database"})
		return
	}

	// Start a transaction
	tx, err := db.Begin()
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
		db.Close()
	}()

	var currentRoomInfo models.RoomRaw
	err = tx.QueryRow("SELECT room_uuid, dorm, dorm_name, room_id, suite_uuid, max_occupancy, current_occupancy, occupants, pull_priority FROM rooms WHERE room_uuid = $1", roomUUIDParam).Scan(
		&currentRoomInfo.RoomUUID,
		&currentRoomInfo.Dorm,
		&currentRoomInfo.DormName,
		&currentRoomInfo.RoomID,
		&currentRoomInfo.SuiteUUID,
		&currentRoomInfo.MaxOccupancy,
		&currentRoomInfo.CurrentOccupancy,
		&currentRoomInfo.Occupants,
		&currentRoomInfo.PullPriority,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query room info from rooms table"})
		return
	}

	// log room uuid
	log.Println(currentRoomInfo.RoomUUID)

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

			if pullLeaderSuiteGroupUUID != uuid.Nil {
				// inherit the suite group's priority
				// for now throw an error because this is not implemented
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Pull leader is already in a suite group (TODO: implement this case)"})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Lock pull not implemented"})
		return
	case 4: // alternative pull
		c.JSON(http.StatusBadRequest, gin.H{"error": "Alternative pull not implemented"})
		return
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pull type: " + string(rune(request.PullType))})
		return
	}

	log.Println(proposedPullPriority)
	log.Print(currentRoomInfo.PullPriority)

	if !comparePullPriority(proposedPullPriority, currentRoomInfo.PullPriority) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Proposed occupants do not have higher priority than current occupants"})
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

	// update the occupants in the database and the current_occupancy
	_, err = tx.Exec("UPDATE rooms SET occupants = $1, current_occupancy = $2 WHERE room_uuid = $3", pq.Array(proposedOccupants), len(proposedOccupants), roomUUIDParam)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update proposed occupants in rooms table"})
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
			var suiteGroupUUID uuid.UUID
			err = tx.QueryRow("INSERT INTO suitegroups (sgroup_size, sgroup_name, sgroup_suite, pull_priority, rooms, disbanded) VALUES ($1, $2, $3, $4, $5, $6) RETURNING sgroup_uuid",
				2,
				"Suite Group",
				currentRoomInfo.SuiteUUID,
				pullLeaderPriorityJSON,
				pq.Array(models.UUIDArray{currentRoomInfo.RoomUUID, request.PullLeaderRoom}),
				false,
			).Scan(&suiteGroupUUID)
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
			// TODO: create new suite group with the pull leader's priority
			// for now throw an error because this is not implemented
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Pull leader is already in a suite group (TODO: implement this case)"})
			return
		}
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
