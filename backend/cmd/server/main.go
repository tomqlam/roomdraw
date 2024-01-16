package main

import (
	"roomdraw/backend/pkg/config"
	"roomdraw/backend/pkg/handlers"
	"roomdraw/backend/pkg/middleware"
	"sync"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	// Initialize the request queue and mutex
	requestQueue := make(chan *gin.Context)
	dbMutex := &sync.Mutex{}

	// Start the request processor goroutine
	go middleware.RequestProcessor(requestQueue, dbMutex)

	// Add the QueueMiddleware to the Gin engine
	router.Use(middleware.QueueMiddleware(requestQueue))

	router.GET("/rooms", handlers.GetRoomsHandler)
	router.GET("/rooms/simple/:dormName", handlers.GetSimpleFormattedDorm)
	router.PATCH("/rooms/:roomuuid", handlers.UpdateRoomOccupants)

	router.GET("/users", handlers.GetUsers)
	router.GET("/users/idmap", handlers.GetUsersIdMap)

	router.Run(config.ServerAddress) // Start server
}
