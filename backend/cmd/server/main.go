package main

import (
	"log"
	"roomdraw/backend/pkg/config"
	"roomdraw/backend/pkg/database"
	"roomdraw/backend/pkg/handlers"
	"roomdraw/backend/pkg/middleware"
	"roomdraw/backend/pkg/logging"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration once
	if err := config.LoadConfig(); err != nil {
		panic(err)
	}

	// Initialize email service after config is loaded
	handlers.InitializeEmailService()

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

	// Initialize the request queue with concurrency of 1
	// This ensures that all write operations are processed one at a time
	requestQueue := middleware.NewRequestQueue(1)

	// Apply the middleware globally
	router.Use(cors.New(corsConfig))

	// Group routes by read and write operations
	readGroup := router.Group("/")
	if config.RequireAuth {
		// Apply JWT only if required, non-blocking for read? Adjust as needed.
		// If JWTAuthMiddleware aborts on failure, unauthenticated users can't read.
		// Consider a less strict auth check for reads if needed.
		readGroup.Use(middleware.JWTAuthMiddleware(false)) // Check token if present, but don't *require* it? Or require for all?
	}

	// Write group - applies Queue, Logging, JWT (required), Blocklist
	writeGroup := router.Group("/")
	writeGroup.Use(middleware.QueueMiddleware(requestQueue))       // 1. Serialize
	if config.RequireAuth {                                        // Only add Auth/Blocklist if required
		writeGroup.Use(logging.TransactionLogMiddleware())         // 2. Add Request ID
		writeGroup.Use(middleware.JWTAuthMiddleware(false))    // 3. Authenticate (non-admin) & add user info to context
		writeGroup.Use(middleware.BlocklistCheckMiddleware())  // 4. Check blocklist
	}

	// Admin Write group - applies Queue, Logging, JWT (admin required)
	writeGroupAdmin := router.Group("/")
	writeGroupAdmin.Use(middleware.QueueMiddleware(requestQueue))  // 1. Serialize
	if config.RequireAuth {                                       // Only add Auth if required
		writeGroupAdmin.Use(logging.TransactionLogMiddleware())    // 2. Add Request ID
		writeGroupAdmin.Use(middleware.JWTAuthMiddleware(true)) // 3. Authenticate (admin required) & add user info
		// No BlocklistCheck needed for admins? Add if needed.
	}

	// Define read-only routes
	readGroup.GET("/rooms", handlers.GetRoomsHandler)
	readGroup.GET("/rooms/simple/:dormName", handlers.GetSimpleFormattedDorm)
	readGroup.GET("/rooms/simpler/:dormName", handlers.GetSimplerFormattedDorm)
	readGroup.GET("/rooms/:roomuuid", handlers.GetRoom)
	readGroup.GET("/users", handlers.GetUsers)
	readGroup.GET("/users/idmap", handlers.GetUsersIdMap)
	readGroup.GET("/users/email", handlers.GetUserByEmail)
	readGroup.GET("/users/:userid", handlers.GetUser)
	readGroup.GET("/users/notifications", handlers.GetNotificationPreference)
	readGroup.GET("/users/clear-room-stats", handlers.GetUserClearRoomStats)

	// New paginated and sorted endpoints
	readGroup.GET("/search/rooms", handlers.GetRoomsPagedAndSorted)
	readGroup.GET("/search/users", handlers.GetUsersPagedAndSorted)

	// Define write routes
	writeGroup.POST("/rooms/:roomuuid", handlers.UpdateRoomOccupants)
	writeGroup.POST("/rooms/indorm/:roomuuid", handlers.ToggleInDorm)
	writeGroup.POST("/rooms/clear/:roomuuid", handlers.ClearRoomHandler)
	writeGroup.POST("/suites/design/:suiteuuid", handlers.SetSuiteDesign)
	writeGroup.POST("/suites/design/remove/:suiteuuid", handlers.DeleteSuiteDesign)
	writeGroup.POST("/frosh/bump/:roomuuid", handlers.BumpFroshHandler)
	writeGroup.POST("/users/notifications", handlers.SetNotificationPreference)

	// Define admin write routes
	writeGroupAdmin.POST("/frosh/:roomuuid", handlers.AddFroshHandler)
	writeGroupAdmin.POST("/frosh/remove/:roomuuid", handlers.RemoveFroshHandler)
	writeGroupAdmin.POST("/rooms/preplace/:roomuuid", handlers.PreplaceOccupants)
	writeGroupAdmin.POST("/rooms/preplace/remove/:roomuuid", handlers.RemovePreplacedOccupantsHandler)
	writeGroupAdmin.GET("/admin/blocklist", handlers.GetBlocklistedUsers)
	writeGroupAdmin.POST("/admin/blocklist/remove/:email", handlers.RemoveUserBlocklist)
	writeGroupAdmin.POST("/admin/suites/update-gender-preferences", handlers.UpdateSuiteGenderPreference)

	log.Println("RequireAuth:", config.RequireAuth)

	// Start the server
	router.Run(config.ServerAddress)
}
