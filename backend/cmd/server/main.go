package main

import (
	"roomdraw/backend/pkg/config"
	"roomdraw/backend/pkg/database"
	"roomdraw/backend/pkg/handlers"
	"roomdraw/backend/pkg/middleware"
	"sync"

	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	err := database.InitDB()
	if err != nil {
		panic(err)
	}
	defer database.DB.Close()

	router := gin.Default()

	// Configure CORS middleware options
	corsConfig := cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:8081", "https://www.cs.hmc.edu"},
		AllowMethods:     []string{"PUT", "PATCH", "GET", "POST", "DELETE"},
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			return origin == "http://localhost:3000" || origin == "http://localhost:8081" || origin == "https://www.cs.hmc.edu"
		},
		MaxAge: 12 * time.Hour,
	}

	// Initialize the RWMutex and request queue
	rwMutex := &sync.RWMutex{}
	requestQueue := make(chan *gin.Context)

	// Start the request processor goroutine for write requests
	go middleware.RequestProcessor(requestQueue)

	// Apply the middleware globally
	router.Use(middleware.QueueMiddleware(rwMutex))
	router.Use(cors.New(corsConfig))

	// Define your routes
	// router.GET("/rooms", handlers.GetRoomsHandler) // Read
	router.GET("/rooms", middleware.JWTAuthMiddleware(), handlers.GetRoomsHandler) // Read
	router.GET("/rooms/simple/:dormName", handlers.GetSimpleFormattedDorm)         // Read
	router.GET("/rooms/simpler/:dormName", handlers.GetSimplerFormattedDorm)       // Read
	router.PATCH("/rooms/:roomuuid", handlers.UpdateRoomOccupants)                 // Write
	router.GET("/users", handlers.GetUsers)                                        // Read
	router.GET("/users/idmap", middleware.JWTAuthMiddleware(), handlers.GetUsersIdMap)
	router.POST("/suites/design", handlers.SetSuiteDesign)

	router.POST("/frosh/:roomuuid", handlers.AddFroshHandler)
	router.DELETE("/frosh/:roomuuid", handlers.RemoveFroshHandler)
	router.PATCH("/frosh/:roomuuid", handlers.BumpFroshHandler)

	// Start the server
	router.Run(config.ServerAddress)
}
