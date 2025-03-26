package handlers

import (
	"errors"
	"log"
	"net/http"
	"os"
	"roomdraw/backend/pkg/database"
	"roomdraw/backend/pkg/models"
	"strings"
	"sync"

	"git.sr.ht/~jamesponddotco/bunnystorage-go"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

func SetSuiteDesign(c *gin.Context) {
	var err error
	// Retrieve the doneChan from the context
	doneChanInterface, exists := c.Get("doneChan")
	if !exists {
		// If for some reason it doesn't exist, log an error and return
		log.Print("Error: doneChan not found in context")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// Assert the type of doneChan to be a chan bool
	doneChan, ok := doneChanInterface.(chan bool)
	if !ok {
		// If the assertion fails, log an error and return
		log.Print("Error: doneChan is not of type chan bool")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// Retrieve the closeOnce from the context
	closeOnceInterface, exists := c.Get("closeOnce")
	if !exists {
		// If for some reason it doesn't exist, log an error and return
		log.Print("Error: closeOnce not found in context")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// Assert the type of closeOnce to be a *sync.Once
	closeOnce, ok := closeOnceInterface.(*sync.Once)
	if !ok {
		// If the assertion fails, log an error and return
		log.Print("Error: closeOnce is not of type *sync.Once")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// Ensure that a signal is sent to doneChan when the function exits, make sure this happens only once
	defer func() {
		closeOnce.Do(func() {
			close(doneChan)
			log.Println("Closed doneChan for request")
		})
	}()

	// constantly listen for the doneChan to be closed (meaning the request was timed out) and return error
	go func() {
		<-doneChan
		log.Println("Request was fulfilled or timed out")
		// write to global error variable
		err = errors.New("request was fulfilled or timed out")
	}()

	suiteUUID := c.Param("suiteuuid")

	var suiteDesignUpdateReq models.SuiteDesignUpdateRequest
	if err := c.ShouldBindJSON(&suiteDesignUpdateReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx, err := database.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}

	// Ensure the transaction is either committed or rolled back
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	_, err = tx.Exec("UPDATE suites SET suite_design = $1 WHERE suite_uuid = $2", suiteDesignUpdateReq.SuiteDesign, suiteUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update suite design"})
		return
	}

	// commit the transaction
	err = tx.Commit()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Suite design updated"})
}

func SetSuiteDesignNew(c *gin.Context) {
	var err error
	// Retrieve the doneChan from the context
	doneChanInterface, exists := c.Get("doneChan")
	if !exists {
		// If for some reason it doesn't exist, log an error and return
		log.Print("Error: doneChan not found in context")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// Assert the type of doneChan to be a chan bool
	doneChan, ok := doneChanInterface.(chan bool)
	if !ok {
		// If the assertion fails, log an error and return
		log.Print("Error: doneChan is not of type chan bool")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// Retrieve the closeOnce from the context
	closeOnceInterface, exists := c.Get("closeOnce")
	if !exists {
		// If for some reason it doesn't exist, log an error and return
		log.Print("Error: closeOnce not found in context")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// Assert the type of closeOnce to be a *sync.Once
	closeOnce, ok := closeOnceInterface.(*sync.Once)
	if !ok {
		// If the assertion fails, log an error and return
		log.Print("Error: closeOnce is not of type *sync.Once")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// Ensure that a signal is sent to doneChan when the function exits, make sure this happens only once
	defer func() {
		closeOnce.Do(func() {
			close(doneChan)
			log.Println("Closed doneChan for request")
		})
	}()

	// constantly listen for the doneChan to be closed (meaning the request was timed out) and return error
	go func() {
		<-doneChan
		log.Println("Request was fulfilled or timed out")
		// write to global error variable
		err = errors.New("request was fulfilled or timed out")
	}()

	suiteUUID := c.Param("suiteuuid")

	tx, err := database.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}

	// Ensure the transaction is either committed or rolled back
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	// ensure the suite exists
	var suiteExists bool
	err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM suites WHERE suite_uuid = $1)", suiteUUID).Scan(&suiteExists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check if suite exists"})
		return
	}

	if !suiteExists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Suite not found"})
		return
	}

	// query for the current suite design
	var currentSuiteDesign string
	err = tx.QueryRow("SELECT suite_design FROM suites WHERE suite_uuid = $1", suiteUUID).Scan(&currentSuiteDesign)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get current suite design"})
		return
	}

	// extract file name from URL
	currentSuiteDesign = currentSuiteDesign[strings.LastIndex(currentSuiteDesign, "/")+1:]

	readOnlyKey, ok := os.LookupEnv("BUNNYNET_READ_API_KEY")
	if !ok {
		log.Fatal("missing env var: BUNNYNET_READ_API_KEY")
		err = errors.New("missing env var: BUNNYNET_READ_API_KEY")
		return
	}

	readWriteKey, ok := os.LookupEnv("BUNNYNET_WRITE_API_KEY")
	if !ok {
		log.Fatal("missing env var: BUNNYNET_WRITE_API_KEY")
		err = errors.New("missing env var: BUNNYNET_WRITE_API_KEY")
		return
	}

	// Create new Config to be initialize a Client.
	cfg := &bunnystorage.Config{
		StorageZone: os.Getenv("BUNNYNET_STORAGE_ZONE"),
		Key:         readWriteKey,
		ReadOnlyKey: readOnlyKey,
		Endpoint:    bunnystorage.EndpointLosAngeles,
	}

	client, err := bunnystorage.NewClient(cfg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create BunnyStorage client"})
		return
	}

	// delete the current suite design from BunnyStorage
	_, err = client.Delete(c, "suite_designs", currentSuiteDesign)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete current suite design"})
		return
	}

	fileHeader, err := c.FormFile("suite_design")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No suite design file uploaded"})
		return
	}

	// Open the file
	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open suite design file"})
		return
	}
	defer file.Close()

	// Read the first 512 bytes to determine the content type
	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read the file"})
		return
	}

	// Reset the read pointer back to the start of the file
	_, err = file.Seek(0, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reset file read pointer"})
		return
	}

	contentType := http.DetectContentType(buffer)
	// Check if the content type is one of the allowed types
	if contentType != "image/svg+xml" && contentType != "image/png" && contentType != "image/jpeg" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File must be an SVG, PNG, or JPEG image"})
		return
	}

	// create random filename
	imageFilename := uuid.New().String()

	if contentType == "image/png" {
		imageFilename += ".png"
	} else if contentType == "image/jpeg" {
		imageFilename += ".jpg"
	} else {
		imageFilename += ".svg"
	}

	// upload the suite design to BunnyStorage
	uploadRes, err := client.Upload(c, "suite_designs", imageFilename, "", file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload suite design"})
		return
	}

	if uploadRes.Status != http.StatusCreated {
		log.Println("Failed to upload suite design to BunnyStorage:", uploadRes.Status)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload suite design"})
		return
	}

	err = godotenv.Load(".env")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load .env file"})
		return
	}

	cdnURL := os.Getenv("CDN_URL")

	imageUrl := cdnURL + "/suite_designs/" + imageFilename

	log.Println(suiteUUID)

	// Add a new link to the CDN to the suite design
	_, err = tx.Exec("UPDATE suites SET suite_design = $1 WHERE suite_uuid = $2", imageUrl, suiteUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update suite design"})
		return
	}

	log.Println("Uploaded suite design to BunnyStorage:", imageUrl)

	// commit the transaction
	err = tx.Commit()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Suite design updated"})
}

func DeleteSuiteDesign(c *gin.Context) {
	var err error
	// Retrieve the doneChan from the context
	doneChanInterface, exists := c.Get("doneChan")
	if !exists {
		// If for some reason it doesn't exist, log an error and return
		log.Print("Error: doneChan not found in context")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// Assert the type of doneChan to be a chan bool
	doneChan, ok := doneChanInterface.(chan bool)
	if !ok {
		// If the assertion fails, log an error and return
		log.Print("Error: doneChan is not of type chan bool")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// Retrieve the closeOnce from the context
	closeOnceInterface, exists := c.Get("closeOnce")
	if !exists {
		// If for some reason it doesn't exist, log an error and return
		log.Print("Error: closeOnce not found in context")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// Assert the type of closeOnce to be a *sync.Once
	closeOnce, ok := closeOnceInterface.(*sync.Once)
	if !ok {
		// If the assertion fails, log an error and return
		log.Print("Error: closeOnce is not of type *sync.Once")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// Ensure that a signal is sent to doneChan when the function exits, make sure this happens only once
	defer func() {
		closeOnce.Do(func() {
			close(doneChan)
			log.Println("Closed doneChan for request")
		})
	}()

	// constantly listen for the doneChan to be closed (meaning the request was timed out) and return error
	go func() {
		<-doneChan
		log.Println("Request was fulfilled or timed out")
		// write to global error variable
		err = errors.New("request was fulfilled or timed out")
	}()

	suiteUUID := c.Param("suiteuuid")
	tx, err := database.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}

	// Ensure the transaction is either committed or rolled back
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	// ensure the suite exists
	var suiteExists bool
	err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM suites WHERE suite_uuid = $1)", suiteUUID).Scan(&suiteExists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check if suite exists"})
		return
	}

	if !suiteExists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Suite not found"})
		return
	}

	// query for the current suite design
	var currentSuiteDesign string
	err = tx.QueryRow("SELECT suite_design FROM suites WHERE suite_uuid = $1", suiteUUID).Scan(&currentSuiteDesign)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get current suite design"})
		return
	}

	// extract file name from URL
	currentSuiteDesign = currentSuiteDesign[strings.LastIndex(currentSuiteDesign, "/")+1:]

	readOnlyKey, ok := os.LookupEnv("BUNNYNET_READ_API_KEY")
	if !ok {
		log.Fatal("missing env var: BUNNYNET_READ_API_KEY")
	}

	readWriteKey, ok := os.LookupEnv("BUNNYNET_WRITE_API_KEY")
	if !ok {
		log.Fatal("missing env var: BUNNYNET_WRITE_API_KEY")
	}

	// Create new Config to be initialize a Client.
	cfg := &bunnystorage.Config{
		StorageZone: os.Getenv("BUNNYNET_STORAGE_ZONE"),
		Key:         readWriteKey,
		ReadOnlyKey: readOnlyKey,
		Endpoint:    bunnystorage.EndpointLosAngeles,
	}

	client, err := bunnystorage.NewClient(cfg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create BunnyStorage client"})
		return
	}

	// delete the current suite design from BunnyStorage
	_, err = client.Delete(c, "suite_designs", currentSuiteDesign)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete current suite design"})
		return
	}

	// remove the suite design from the suite default ''
	_, err = tx.Exec("UPDATE suites SET suite_design = '' WHERE suite_uuid = $1", suiteUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove suite design"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Suite design deleted"})
}
