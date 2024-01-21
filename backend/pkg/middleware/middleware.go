package middleware

import (
	"sync"

	"github.com/gin-gonic/gin"
)

// Define a type for the request type (read or write)
type RequestType int

const (
	Read RequestType = iota
	Write
)

func QueueMiddleware(rwMutex *sync.RWMutex) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Determine whether the request is a read or write operation
		// This could be based on the HTTP method or specific routes
		isRead := c.Request.Method == "GET" || c.Request.Method == "HEAD" // Add other read methods as needed

		// Store this information in the context
		c.Set("isRead", isRead)

		// If it's a read operation, we can allow it to proceed in parallel with other reads
		if isRead {
			rwMutex.RLock()
			c.Next()
			rwMutex.RUnlock()
		} else {
			// If it's a write operation, it must be processed sequentially
			// So we send it to the request processor
			c.Set("rwMutex", rwMutex) // Pass the mutex for the processor to use
			c.Next()                  // The actual locking and unlocking for writes will be handled in the processor
		}
	}
}

func RequestProcessor(requestQueue <-chan *gin.Context) {
	for c := range requestQueue {
		// Retrieve the mutex from the context
		val, exists := c.Get("rwMutex")
		if !exists {
			// handle the error, the mutex was not found
			continue
		}
		rwMutex, ok := val.(*sync.RWMutex)
		if !ok {
			// handle the error, the type assertion was not successful
			continue
		}

		// Retrieve the request type (read/write) from the context
		val, exists = c.Get("isRead")
		if !exists {
			// handle the error, the request type was not found
			continue
		}
		isRead, ok := val.(bool)
		if !ok {
			// handle the error, the type assertion was not successful
			continue
		}

		if isRead {
			// It's a read request, already handled in the middleware
			// You might want to put some logic here if needed
		} else {
			// It's a write request, so we lock for writing
			rwMutex.Lock()
			// The actual processing logic goes here...
			// ... (handle the request)
			rwMutex.Unlock()
		}
	}
}

// cors

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
