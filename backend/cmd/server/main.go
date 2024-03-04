package main

import (
	"roomdraw/backend/pkg/config"
	"roomdraw/backend/pkg/database"
	"roomdraw/backend/pkg/handlers"
	"roomdraw/backend/pkg/middleware"

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

	// Apply the middleware globally
	router.Use(cors.New(corsConfig))

	// Group routes by read and write operations
	readGroup := router.Group("/").Use(middleware.JWTAuthMiddleware(false))
	{
		// Define read-only routes here
		readGroup.GET("/rooms", handlers.GetRoomsHandler)
		readGroup.GET("/rooms/simple/:dormName", handlers.GetSimpleFormattedDorm)
		readGroup.GET("/rooms/simpler/:dormName", handlers.GetSimplerFormattedDorm)
		readGroup.GET("/users", handlers.GetUsers)
		readGroup.GET("/users/idmap", handlers.GetUsersIdMap)
	}

	// Initialize the RWMutex and request queue
	requestQueue := make(chan *gin.Context)

	// For write operations, use a separate group and apply the queue middleware
	writeGroup := router.Group("/").Use(middleware.QueueMiddleware(requestQueue), middleware.JWTAuthMiddleware(false))
	{
		// Define write routes here
		writeGroup.POST("/rooms/:roomuuid", handlers.UpdateRoomOccupants)
		writeGroup.POST("/suites/design/:suiteuuid", handlers.SetSuiteDesign)
		writeGroup.POST("/frosh/:roomuuid", handlers.AddFroshHandler)
		writeGroup.DELETE("/frosh/:roomuuid", handlers.RemoveFroshHandler)
		writeGroup.POST("/frosh/bump/:roomuuid", handlers.BumpFroshHandler)
	}

	// Start the request processor goroutine for write requests
	go middleware.RequestProcessor(requestQueue)

	// Start the server
	router.Run(config.ServerAddress)
}
