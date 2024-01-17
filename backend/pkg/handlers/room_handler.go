package handlers

import (
	"encoding/json"
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
	rows, err := db.Query("SELECT * FROM rooms")
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
		if err := rows.Scan(&d.RoomUUID, &d.Dorm, &d.DormName, &d.RoomID, &d.SuiteUUID, &d.MaxOccupancy, &d.CurrentOccupancy, &d.Occupants, &d.PullPriority); err != nil {
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
	defer db.Close()

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
			Rooms: suiteToRoomMap[suiteUUIDString],
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

	occupants := request.ProposedOccupants

	db, err := database.NewDatabase()
	if err != nil {
		// Handle error opening the database
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to connect to database"})
		return
	}
	defer db.Close()

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
	}()

	var roomInfo models.RoomRaw
	err = tx.QueryRow("SELECT room_uuid, dorm, dorm_name, room_id, suite_uuid, max_occupancy, current_occupancy, occupants, pull_priority FROM rooms WHERE room_uuid = $1", roomUUIDParam).Scan(&roomInfo.RoomUUID, &roomInfo.Dorm, &roomInfo.DormName, &roomInfo.RoomID, &roomInfo.SuiteUUID, &roomInfo.MaxOccupancy, &roomInfo.CurrentOccupancy, &roomInfo.Occupants, &roomInfo.PullPriority)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query room info from rooms table"})
		return
	}

	// log room uuid
	log.Println(roomInfo.RoomUUID)

	// check that the proposed occupants are not more than the max occupancy
	if len(occupants) > roomInfo.MaxOccupancy {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Proposed occupants exceeds max occupancy"})
		return
	}

	// check that all of the proposed occupants are not already in a room
	for _, occupant := range occupants {
		var roomUUID uuid.UUID
		err = tx.QueryRow("SELECT room_uuid FROM users WHERE id = $1", occupant).Scan(&roomUUID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query room_uuid from users table"})
			return
		}
		if roomUUID != uuid.Nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "One or more of the proposed occupants is already in a room"})
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
	_, err = tx.Exec("UPDATE rooms SET occupants = $1, current_occupancy = $2 WHERE room_uuid = $3", pq.Array(occupants), len(occupants), roomUUIDParam)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update proposed occupants in rooms table"})
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// for each occupant, update the room_uuid field in the users table
	for _, occupant := range occupants {
		_, err = tx.Exec("UPDATE users SET room_uuid = $1 WHERE id = $2", roomUUIDParam, occupant)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update room_uuid in users table"})
			return
		}
	}

	// calculate the new pull priority by sorting the occupants by draw number and getting the first one
	var newPullPriority models.PullPriority

	switch request.PullType {
	case 1: // self pull
		if len(occupants) > 0 {
			var occupantsInfo []models.UserRaw
			rows, err := tx.Query("SELECT id, draw_number, year, in_dorm FROM users WHERE id = ANY($1)", pq.Array(occupants))
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

			sortedOccupants := sortUsersByPriority(occupantsInfo, roomInfo.Dorm)

			newPullPriority = generateUserPriority(sortedOccupants[0], roomInfo.Dorm)

			newPullPriority.Valid = true
			newPullPriority.PullType = 1

			newPullPriorityJSON, err := json.Marshal(newPullPriority)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal new pull priority"})
				return
			}

			_, err = tx.Exec("UPDATE rooms SET pull_priority = $1 WHERE room_uuid = $2", newPullPriorityJSON, roomUUIDParam)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update pull priority in rooms table"})
				return
			}
		} else {
			emptyPriority := generateEmptyPriority()
			emptyPriorityJSON, err := json.Marshal(emptyPriority)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal empty pull priority"})
				return
			}

			_, err = tx.Exec("UPDATE rooms SET pull_priority = $1 WHERE room_uuid = $2", emptyPriorityJSON, roomUUIDParam)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update empty pull priority in rooms table"})
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}
	case 2: // normal pull
		if len(occupants) > 0 {
			pullLeaderRoomUUID := request.PullLeaderRoom
			var occupantsInfo []models.UserRaw
			rows, err := tx.Query("SELECT id, draw_number, year, in_dorm FROM users WHERE id = ANY($1)", pq.Array(occupants))
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

			var pullLeaderPriority models.PullPriority
			// get the pull leader's priority
			err = tx.QueryRow("SELECT pull_priority FROM rooms WHERE room_uuid = $1", pullLeaderRoomUUID).Scan(&pullLeaderPriority)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query pull leader's priority from rooms table"})
				return
			}

			pullLeaderPriorityJSON, err := json.Marshal(pullLeaderPriority)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal pull leader's priority"})
				return
			}

			sortedOccupants := sortUsersByPriority(occupantsInfo, roomInfo.Dorm)

			newPullPriority = generateUserPriority(sortedOccupants[0], roomInfo.Dorm)

			newPullPriority.Valid = true
			newPullPriority.PullType = 1

			newPullPriority.Inherited.DrawNumber = pullLeaderPriority.DrawNumber
			newPullPriority.Inherited.HasInDorm = pullLeaderPriority.HasInDorm
			newPullPriority.Inherited.Year = pullLeaderPriority.Year

			newPullPriorityJSON, err := json.Marshal(newPullPriority)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal new pull priority"})
				return
			}

			_, err = tx.Exec("UPDATE rooms SET pull_priority = $1 WHERE room_uuid = $2", newPullPriorityJSON, roomUUIDParam)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update pull priority in rooms table"})
				return
			}

			// create new suite group
			// don't need to generate new UUID, because it will be generated by the database
			// priority is the pull leader's priority
			var suiteGroupUUID uuid.UUID
			err = tx.QueryRow("INSERT INTO suite_groups (sgroup_size, sgroup_name, sgroup_suite, sgroup_priority, rooms, disbanded) VALUES ($1, $2, $3, $4, $5, $6) RETURNING sgroup_uuid", len(occupants), "Suite Group", roomInfo.SuiteUUID, pullLeaderPriorityJSON, pq.Array([]uuid.UUID{roomInfo.RoomUUID, pullLeaderRoomUUID}), false).Scan(&suiteGroupUUID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert new suite group into suite_groups table"})
				return
			}

			// update the suite_uuid field in the rooms table for both rooms

		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Normal pull requires at least one occupant"})
		}
	case 3: // lock pull
		break
	case 4: // alternative pull
		break
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pull type"})
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully updated occupants"})
}
