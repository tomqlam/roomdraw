package main

import (
	"log"
	"roomdraw/backend/pkg/config"
	"roomdraw/backend/pkg/database"
	"roomdraw/backend/pkg/handlers"
	"roomdraw/backend/pkg/middleware"
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
	writeGroup := router.Group("/").Use(middleware.QueueMiddleware(requestQueue))
	writeGroupAdmin := router.Group("/").Use(middleware.QueueMiddleware(requestQueue))

	if config.RequireAuth {
		readGroup.Use(middleware.JWTAuthMiddleware(false))
		writeGroup.Use(middleware.JWTAuthMiddleware(false), middleware.BlacklistCheckMiddleware())
		writeGroupAdmin.Use(middleware.JWTAuthMiddleware(true))
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
	writeGroupAdmin.GET("/admin/blacklisted-users", handlers.GetBlacklistedUsers)
	writeGroupAdmin.POST("/admin/blacklist/remove/:email", handlers.RemoveUserBlacklist)
	writeGroupAdmin.POST("/admin/suites/update-gender-preferences", handlers.UpdateSuiteGenderPreference)

	log.Println("RequireAuth:", config.RequireAuth)

	// Start the server
	router.Run(config.ServerAddress)
}
