package handlers

import (
	"log"
	"net/http"
	"roomdraw/backend/pkg/database"
	"roomdraw/backend/pkg/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
			// Handle scan error
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database scan failed on rooms"})
			return
		}
		rooms = append(rooms, d)
	}

	rows, err = db.Query("SELECT suite_uuid, dorm, dorm_name, floor, room_count, rooms, alternative_pull, suite_design FROM suites WHERE UPPER(dorm_name) = UPPER($1)", dormNameParam)
	if err != nil {
		// Handle query error
		// print the error to the console
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed on suites"})
		return
	}

	var suites []models.SuiteRaw
	for rows.Next() {
		var s models.SuiteRaw
		if err := rows.Scan(&s.SuiteUUID, &s.Dorm, &s.DormName, &s.Floor, &s.RoomCount, &s.Rooms, &s.AlternativePull, &s.SuiteDesign); err != nil {
			// Handle scan error
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

	// get the list of user ids from the request body
	var occupants []int
	if err := c.ShouldBindJSON(&occupants); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

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

	// remove the current occupants from the room
	_, err = tx.Exec("UPDATE rooms SET occupants = $1, current_occupancy = $2 WHERE room_uuid = $3", nil, 0, roomUUIDParam)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update occupants in rooms table"})
		return
	}

	// for each current occupant, nullify the room_uuid field in the users table
	_, err = tx.Exec("UPDATE users SET room_uuid = $1 WHERE room_uuid = $2", nil, roomUUIDParam)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update room_uuid in users table"})
		return
	}

	// update the occupants in the database and the current_occupancy
	_, err = tx.Exec("UPDATE rooms SET occupants = $1, current_occupancy = $2 WHERE room_uuid = $3", occupants, len(occupants), roomUUIDParam)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update occupants in rooms table"})
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

	c.JSON(http.StatusOK, gin.H{"message": "Successfully updated occupants"})
}
