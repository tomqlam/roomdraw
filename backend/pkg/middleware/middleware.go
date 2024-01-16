package middleware

import (
	"sync"

	"github.com/gin-gonic/gin"
)

func QueueMiddleware(requestQueue chan<- *gin.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestQueue <- c
		c.Next()
	}
}

func RequestProcessor(requestQueue <-chan *gin.Context, dbMutex *sync.Mutex) {
	for _ = range requestQueue {
		dbMutex.Lock()
		// Perform validation and database operations
		dbMutex.Unlock()
	}
}
