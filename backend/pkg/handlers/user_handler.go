package handlers

import (
	"log"
	"net/http"
	"roomdraw/backend/pkg/database"
	"roomdraw/backend/pkg/models"

	"github.com/gin-gonic/gin"
)

func GetUsers(c *gin.Context) {
	// Example SQL query
	rows, err := database.DB.Query("SELECT id, year, first_name, last_name, draw_number, preplaced, in_dorm, sgroup_uuid, participated, participation_time, room_uuid FROM users")
	if err != nil {

		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed"})
		return
	}
	defer rows.Close()

	var users []models.UserRaw
	for rows.Next() {
		var user models.UserRaw
		if err := rows.Scan(&user.Id, &user.Year, &user.FirstName, &user.LastName, &user.DrawNumber, &user.Preplaced, &user.InDorm, &user.SGroupUUID, &user.Participated, &user.PartitipationTime, &user.RoomUUID); err != nil {
			// Handle scan error
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database scan failed"})
			return
		}

		users = append(users, user)
	}

	c.JSON(http.StatusOK, users)
}

func GetUsersIdMap(c *gin.Context) {
	// Example SQL query
	rows, err := database.DB.Query("SELECT id, year, first_name, last_name, draw_number, preplaced, in_dorm, sgroup_uuid, participated, participation_time,room_uuid, reslife_role FROM users")
	if err != nil {

		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed"})
		return
	}
	defer rows.Close()

	var users []models.UserRaw
	for rows.Next() {
		var user models.UserRaw
		if err := rows.Scan(&user.Id, &user.Year, &user.FirstName, &user.LastName, &user.DrawNumber, &user.Preplaced, &user.InDorm, &user.SGroupUUID, &user.Participated, &user.PartitipationTime, &user.RoomUUID, &user.ReslifeRole); err != nil {
			// Handle scan error
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database scan failed"})
			return
		}
		users = append(users, user)
	}

	// create a map of user ids to user objects
	userMap := make(map[int]models.UserRaw)
	for _, user := range users {
		userMap[user.Id] = user
	}

	c.JSON(http.StatusOK, userMap)
}

func GetUser(c *gin.Context) {
	// Get the user id from the URL
	userid := c.Param("userid")

	rows, err := database.DB.Query("SELECT id, year, first_name, last_name, draw_number, preplaced, in_dorm, sgroup_uuid, participated, room_uuid FROM users WHERE id=$1", userid)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed"})
		return
	}
	defer rows.Close()

	var user models.UserRaw
	for rows.Next() {
		if err := rows.Scan(&user.Id, &user.Year, &user.FirstName, &user.LastName, &user.DrawNumber, &user.Preplaced, &user.InDorm, &user.SGroupUUID, &user.Participated, &user.RoomUUID); err != nil {
			// Handle scan error
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database scan failed"})
			return
		}
	}

	c.JSON(http.StatusOK, user)
}
