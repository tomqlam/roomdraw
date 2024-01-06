package handlers

import (
	"log"
	"net/http"
	"roomdraw/backend/pkg/database"
	"roomdraw/backend/pkg/models"

	"github.com/gin-gonic/gin"
)

func GetUsers(c *gin.Context) {
	db, err := database.NewDatabase()
	if err != nil {
		// Handle error opening the database
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to connect to database"})
		return
	}
	defer db.Close()

	// Example SQL query
	rows, err := db.Query("SELECT id, year, first_name, last_name, draw_number, preplaced, in_dorm, sgroup_uuid, participated, room_uuid FROM users")
	if err != nil {
		// Handle query error
		// print the error to the console
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed"})
		return
	}
	defer rows.Close()

	var users []models.UserRaw
	for rows.Next() {
		var user models.UserRaw
		if err := rows.Scan(&user.Id, &user.Year, &user.FirstName, &user.LastName, &user.DrawNumber, &user.Preplaced, &user.InDorm, &user.SGroupUUID, &user.Participated, &user.RoomUUID); err != nil {
			// Handle scan error
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database scan failed"})
			return
		}
		users = append(users, user)
	}

	c.JSON(http.StatusOK, users)
}
