package handlers

import (
	"log"
	"net/http"
	"roomdraw/backend/pkg/database"
	"roomdraw/backend/pkg/models"

	"github.com/gin-gonic/gin"
)

func AddFroshHandler(c *gin.Context) {
	// get the room uuid from the request url
	roomUUID := c.Param("roomuuid")

	// start the transaction
	tx, err := database.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}

	// ensure the transaction is either committed or rolled back
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

	// get the room from the database
	var room models.RoomRaw
	err = tx.QueryRow("SELECT room_uuid, dorm, dorm_name, room_id, suite_uuid, max_occupancy, current_occupancy, occupants, pull_priority, sgroup_uuid, has_frosh FROM rooms WHERE room_uuid = $1", roomUUID).Scan(&room.RoomUUID, &room.Dorm, &room.DormName, &room.RoomID, &room.SuiteUUID, &room.MaxOccupancy, &room.CurrentOccupancy, &room.Occupants, &room.PullPriority, &room.SGroupUUID, &room.HasFrosh)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get room from database"})
		return
	}

	// make sure room is empty
	if room.CurrentOccupancy != 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Room is not empty"})
		return
	}

	// add the frosh to the room
	_, err = tx.Exec("UPDATE rooms SET has_frosh = true WHERE room_uuid = $1", roomUUID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add frosh to room"})
		return
	}

	// if room.Dorm between 1 and 4
	if room.Dorm >= 1 && room.Dorm <= 4 {
		// check that all the other rooms in the suite are empty
		var count int
		err = tx.QueryRow("SELECT COUNT(*) FROM rooms WHERE suite_uuid = $1 AND room_uuid != $2 AND current_occupancy != 0", room.SuiteUUID, roomUUID).Scan(&count)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check other rooms in the suite"})
			return
		}

		if count > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Other rooms in the suite are not empty"})
			return
		}

		// add frosh to all the rooms with the same suite_uuid
		_, err = tx.Exec("UPDATE rooms SET has_frosh = true WHERE suite_uuid = $1", room.SuiteUUID)

		log.Println("Frosh added to suite because frosh was placed in inner dorm")
	}

	// commit the transaction
	err = tx.Commit()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Frosh added to room"})
}

func RemoveFroshHandler(c *gin.Context) {
	// get the room uuid from the request url
	roomUUID := c.Param("roomuuid")

	// start the transaction
	tx, err := database.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}

	// ensure the transaction is either committed or rolled back
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

	// get the room from the database
	var room models.RoomRaw
	err = tx.QueryRow("SELECT room_uuid, dorm, dorm_name, room_id, suite_uuid, max_occupancy, current_occupancy, occupants, pull_priority, sgroup_uuid, has_frosh FROM rooms WHERE room_uuid = $1", roomUUID).Scan(&room.RoomUUID, &room.Dorm, &room.DormName, &room.RoomID, &room.SuiteUUID, &room.MaxOccupancy, &room.CurrentOccupancy, &room.Occupants, &room.PullPriority, &room.SGroupUUID, &room.HasFrosh)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get room from database"})
		return
	}

	// remove the frosh from the room
	_, err = tx.Exec("UPDATE rooms SET has_frosh = false WHERE room_uuid = $1", roomUUID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove frosh from room"})
		return
	}

	// if room.Dorm between 1 and 4
	if room.Dorm >= 1 && room.Dorm <= 4 {
		// remove frosh from all the rooms with the same suite_uuid
		_, err = tx.Exec("UPDATE rooms SET has_frosh = false WHERE suite_uuid = $1", room.SuiteUUID)

		log.Println("Frosh removed from suite because frosh was removed from inner dorm")
	}

	// commit the transaction
	err = tx.Commit()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Frosh removed from room"})
}

func BumpFroshHandler(c *gin.Context) {
	// get the room uuid from the request url
	// roomUUID := c.Param("roomuuid")

	// get the bump frosh request from the request body
	var bumpFroshReq models.BumpFroshRequest
	if err := c.ShouldBindJSON(&bumpFroshReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// start the transaction

	tx, err := database.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}

	// ensure the transaction is either committed or rolled back
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

}
