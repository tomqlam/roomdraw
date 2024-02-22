package handlers

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"roomdraw/backend/pkg/database"
	"roomdraw/backend/pkg/models"

	"github.com/gin-gonic/gin"
)

func AddFroshHandler(c *gin.Context) { // should be a secured route
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
	err = tx.QueryRow("SELECT room_uuid, dorm, dorm_name, room_id, suite_uuid, max_occupancy, current_occupancy, occupants, pull_priority, sgroup_uuid, has_frosh, frosh_room_type FROM rooms WHERE room_uuid = $1", roomUUID).Scan(&room.RoomUUID, &room.Dorm, &room.DormName, &room.RoomID, &room.SuiteUUID, &room.MaxOccupancy, &room.CurrentOccupancy, &room.Occupants, &room.PullPriority, &room.SGroupUUID, &room.HasFrosh, &room.FroshRoomType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get room from database"})
		return
	}

	// ensure that frosh_room_type is not 0
	if room.FroshRoomType == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Room is not a frosh room"})
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

		log.Println("Frosh added to all rooms in suite because frosh was placed in inner dorm")
	}

	// commit the transaction
	err = tx.Commit()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Frosh added to room"})
}

func RemoveFroshHandler(c *gin.Context) { // should be a secured route
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
	err = tx.QueryRow("SELECT room_uuid, dorm, dorm_name, room_id, suite_uuid, max_occupancy, current_occupancy, occupants, pull_priority, sgroup_uuid, has_frosh, frosh_room_type FROM rooms WHERE room_uuid = $1", roomUUID).Scan(&room.RoomUUID, &room.Dorm, &room.DormName, &room.RoomID, &room.SuiteUUID, &room.MaxOccupancy, &room.CurrentOccupancy, &room.Occupants, &room.PullPriority, &room.SGroupUUID, &room.HasFrosh, &room.FroshRoomType)
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
	roomUUID := c.Param("roomuuid")

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

	var originalRoom models.RoomRaw
	err = tx.QueryRow("SELECT room_uuid, dorm, dorm_name, room_id, suite_uuid, max_occupancy, current_occupancy, occupants, pull_priority, sgroup_uuid, has_frosh, frosh_room_type FROM rooms WHERE room_uuid = $1", roomUUID).Scan(&originalRoom.RoomUUID, &originalRoom.Dorm, &originalRoom.DormName, &originalRoom.RoomID, &originalRoom.SuiteUUID, &originalRoom.MaxOccupancy, &originalRoom.CurrentOccupancy, &originalRoom.Occupants, &originalRoom.PullPriority, &originalRoom.SGroupUUID, &originalRoom.HasFrosh, &originalRoom.FroshRoomType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get room from database"})
		return
	}

	if !originalRoom.HasFrosh {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Room does not have a frosh"})
		return
	}

	// check that the reslife_room column in the suite table is null meaning a frosh is not being bumped out of a reslife suite
	var reslifeSuiteUUID sql.NullString
	err = tx.QueryRow("SELECT reslife_room FROM suites WHERE suite_uuid = $1", originalRoom.SuiteUUID).Scan(&reslifeSuiteUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get suite from database"})
		return
	}

	if reslifeSuiteUUID.Valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Frosh is in a reslife suite and cannot be bumped"})
		return
	}

	// get the target room info
	var targetRoom models.RoomRaw
	err = tx.QueryRow("SELECT room_uuid, dorm, dorm_name, room_id, suite_uuid, max_occupancy, current_occupancy, occupants, pull_priority, sgroup_uuid, has_frosh, frosh_room_type FROM rooms WHERE room_uuid = $1", bumpFroshReq.TargetRoomUUID).Scan(&targetRoom.RoomUUID, &targetRoom.Dorm, &targetRoom.DormName, &targetRoom.RoomID, &targetRoom.SuiteUUID, &targetRoom.MaxOccupancy, &targetRoom.CurrentOccupancy, &targetRoom.Occupants, &targetRoom.PullPriority, &targetRoom.SGroupUUID, &targetRoom.HasFrosh, &targetRoom.FroshRoomType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get target room from database"})
		return
	}

	// verify that the target room is in the same dorm
	if targetRoom.Dorm != originalRoom.Dorm {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Target room is not in the same dorm as the original room"})
		return
	}

	if targetRoom.HasFrosh {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Target room already has a frosh"})
		return
	}

	if targetRoom.FroshRoomType != originalRoom.FroshRoomType {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Target room is not the same type as the original room"})
		return
	}

	if originalRoom.Dorm >= 1 && originalRoom.Dorm <= 4 {
		err = BumpFroshInnerDormHelper(tx, originalRoom, targetRoom)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	} else {
		switch originalRoom.Dorm {
		case 5:
			err = BumpFroshAtwoodHelper(tx, originalRoom, targetRoom)
		case 6:
			err = BumpFroshSontagHelper(tx, originalRoom, targetRoom)
		case 7:
			err = BumpFroshLindeHelper(tx, originalRoom, targetRoom)
		case 8:
			err = BumpFroshCaseHelper(tx, originalRoom, targetRoom)
		case 9:
			err = BumpFroshDrinkwardHelper(tx, originalRoom, targetRoom)
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid dorm"})
			err = errors.New("invalid dorm")
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
}

func BumpFroshAtwoodHelper(tx *sql.Tx, originalRoom models.RoomRaw, targetRoom models.RoomRaw) error {
	var err error
	// in atwood, just verify that the target room is empty
	if targetRoom.CurrentOccupancy != 0 {
		return errors.New("target room is not empty")
	}

	// set the has_frosh field to false for the original room and true for the target room
	_, err = tx.Exec("UPDATE rooms SET has_frosh = false WHERE room_uuid = $1", originalRoom.RoomUUID)

	if err != nil {
		return err
	}

	_, err = tx.Exec("UPDATE rooms SET has_frosh = true WHERE room_uuid = $1", targetRoom.RoomUUID)

	if err != nil {
		return err
	}

	return nil
}

func BumpFroshSontagHelper(tx *sql.Tx, originalRoom models.RoomRaw, targetRoom models.RoomRaw) error {
	var err error
	// in sontag, just verify that the target room is empty
	if targetRoom.CurrentOccupancy != 0 {
		return errors.New("target room is not empty")
	}

	// set the has_frosh field to false for the original room and true for the target room
	_, err = tx.Exec("UPDATE rooms SET has_frosh = false WHERE room_uuid = $1", originalRoom.RoomUUID)
	if err != nil {
		return err
	}

	_, err = tx.Exec("UPDATE rooms SET has_frosh = true WHERE room_uuid = $1", targetRoom.RoomUUID)
	if err != nil {
		return err
	}

	return nil
}

func BumpFroshLindeHelper(tx *sql.Tx, originalRoom models.RoomRaw, targetRoom models.RoomRaw) error {
	var err error
	// in linde, verify that the target room is empty and also that the target suite doesn't already have frosh
	if targetRoom.CurrentOccupancy != 0 {
		return errors.New("target room is not empty")
	}

	var count int
	err = tx.QueryRow("SELECT COUNT(*) FROM rooms WHERE suite_uuid = $1 AND has_frosh = true", targetRoom.SuiteUUID).Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		return errors.New("target suite already has frosh")
	}

	// set the has_frosh field to false for the original room and true for the target room
	_, err = tx.Exec("UPDATE rooms SET has_frosh = false WHERE room_uuid = $1", originalRoom.RoomUUID)
	if err != nil {
		return err
	}

	_, err = tx.Exec("UPDATE rooms SET has_frosh = true WHERE room_uuid = $1", targetRoom.RoomUUID)
	if err != nil {
		return err
	}

	return nil
}

func BumpFroshCaseHelper(tx *sql.Tx, originalRoom models.RoomRaw, targetRoom models.RoomRaw) error {
	var err error
	// in case, just verify that the target room is empty
	if targetRoom.CurrentOccupancy != 0 {
		return errors.New("target room is not empty")
	}

	// set the has_frosh field to false for the original room and true for the target room
	_, err = tx.Exec("UPDATE rooms SET has_frosh = false WHERE room_uuid = $1", originalRoom.RoomUUID)
	if err != nil {
		return err
	}

	_, err = tx.Exec("UPDATE rooms SET has_frosh = true WHERE room_uuid = $1", targetRoom.RoomUUID)
	if err != nil {
		return err
	}

	return nil
}

func BumpFroshDrinkwardHelper(tx *sql.Tx, originalRoom models.RoomRaw, targetRoom models.RoomRaw) error {
	var err error
	// in case, just verify that the target room is empty
	if targetRoom.CurrentOccupancy != 0 {
		return errors.New("target room is not empty")
	}

	// set the has_frosh field to false for the original room and true for the target room
	_, err = tx.Exec("UPDATE rooms SET has_frosh = false WHERE room_uuid = $1", originalRoom.RoomUUID)
	if err != nil {
		return err
	}

	_, err = tx.Exec("UPDATE rooms SET has_frosh = true WHERE room_uuid = $1", targetRoom.RoomUUID)
	if err != nil {
		return err
	}

	return nil
}

func BumpFroshInnerDormHelper(tx *sql.Tx, originalRoom models.RoomRaw, targetRoom models.RoomRaw) error {
	var err error
	// in the inner dorms, the entire suite of frosh must be moved
	// check that all the other rooms in the target suite are empty
	var count int
	err = tx.QueryRow("SELECT COUNT(*) FROM rooms WHERE suite_uuid = $1 AND room_uuid != $2 AND current_occupancy != 0", targetRoom.SuiteUUID, targetRoom.RoomUUID).Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		return errors.New("other rooms in the target suite are not empty")
	}

	// set the has_frosh field to false for all rooms in the original suite and true for all rooms in the target suite
	_, err = tx.Exec("UPDATE rooms SET has_frosh = false WHERE suite_uuid = $1", originalRoom.SuiteUUID)
	if err != nil {
		return err
	}

	_, err = tx.Exec("UPDATE rooms SET has_frosh = true WHERE suite_uuid = $1", targetRoom.SuiteUUID)
	if err != nil {
		return err
	}

	return nil
}
