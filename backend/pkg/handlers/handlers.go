package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetDorms(c *gin.Context) {
	// Implement logic to retrieve dorms
	c.JSON(http.StatusOK, gin.H{"message": "Dorms retrieved!!!!!"})
}

func CreateDorm(c *gin.Context) {
	// Implement logic to create a dorm
	c.JSON(http.StatusCreated, gin.H{"message": "Dorm created"})
}

func UpdateDorm(c *gin.Context) {
	// Implement logic to update a dorm
	c.JSON(http.StatusOK, gin.H{"message": "Dorm updated"})
}

func DeleteDorm(c *gin.Context) {
	// Implement logic to delete a dorm
	c.JSON(http.StatusOK, gin.H{"message": "Dorm deleted"})
}
