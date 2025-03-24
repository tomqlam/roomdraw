package main

import (
	"log"
	"roomdraw/backend/pkg/config"
	"roomdraw/backend/pkg/database"
	"roomdraw/backend/pkg/handlers"
	"roomdraw/backend/pkg/middleware"

	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		panic(err)
	}

	requireAuth := (os.Getenv("REQUIRE_AUTH") == "True")

	err = database.InitDB()
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
	requestQueue := make(chan *gin.Context)

	// Apply the middleware globally
	router.Use(cors.New(corsConfig))

	// Start the request processor goroutine for write requests
	go middleware.RequestProcessor(requestQueue)

	// Group routes by read and write operations
	readGroup := router.Group("/")
	writeGroup := router.Group("/").Use(middleware.QueueMiddleware(requestQueue))
	writeGroupAdmin := router.Group("/").Use(middleware.QueueMiddleware(requestQueue))

	if requireAuth {
		readGroup.Use(middleware.JWTAuthMiddleware(false))
		writeGroup.Use(middleware.JWTAuthMiddleware(false))
		writeGroupAdmin.Use(middleware.JWTAuthMiddleware(true))
	}

	// Define read-only routes
	readGroup.GET("/rooms", handlers.GetRoomsHandler)
	readGroup.GET("/rooms/simple/:dormName", handlers.GetSimpleFormattedDorm)
	readGroup.GET("/rooms/simpler/:dormName", handlers.GetSimplerFormattedDorm)
	readGroup.GET("/rooms/:roomuuid", handlers.GetRoom)
	readGroup.GET("/users", handlers.GetUsers)
	readGroup.GET("/users/idmap", handlers.GetUsersIdMap)
	readGroup.GET("/users/:userid", handlers.GetUser)
	readGroup.GET("/users/notifications", handlers.GetNotificationPreference)

	// Define write routes
	writeGroup.POST("/rooms/:roomuuid", handlers.UpdateRoomOccupants)
	writeGroup.POST("/rooms/indorm/:roomuuid", handlers.ToggleInDorm)
	writeGroup.POST("/suites/design/:suiteuuid", handlers.SetSuiteDesignNew)
	writeGroup.POST("/suites/design/remove/:suiteuuid", handlers.DeleteSuiteDesign)
	writeGroup.POST("/frosh/bump/:roomuuid", handlers.BumpFroshHandler)
	writeGroup.POST("/users/notifications", handlers.SetNotificationPreference)

	// Define admin write routes
	writeGroupAdmin.POST("/frosh/:roomuuid", handlers.AddFroshHandler)
	writeGroupAdmin.POST("/frosh/remove/:roomuuid", handlers.RemoveFroshHandler)
	writeGroupAdmin.POST("/rooms/preplace/:roomuuid", handlers.PreplaceOccupants)

	log.Println(requireAuth)

	// Start the server
	router.Run(config.ServerAddress)
}
