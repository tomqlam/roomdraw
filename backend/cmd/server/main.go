package main

import (
	"roomdraw/backend/pkg/config"
	"roomdraw/backend/pkg/handlers"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	// Set up routes
	router.GET("/dorms", handlers.GetDorms)
	router.POST("/dorms", handlers.CreateDorm)
	router.PUT("/dorms/:id", handlers.UpdateDorm)
	router.DELETE("/dorms/:id", handlers.DeleteDorm)

	router.GET("/rooms", handlers.GetRoomsHandler)
	router.GET("/rooms/simple/:dormName", handlers.GetSimpleFormattedDorm)

	router.GET("/users", handlers.GetUsers)
	router.GET("/users/idmap", handlers.GetUsersIdMap)

	router.Run(config.ServerAddress) // Start server
}
