package handlers

import (
	"net/http"
	"roomdraw/backend/pkg/database"
	"roomdraw/backend/pkg/models"

	"github.com/gin-gonic/gin"
)

func SetSuiteDesign(c *gin.Context) {
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

	// Replace the suite design
	_, err = tx.Exec("UPDATE suite_designs SET design = $1 WHERE id = $2", suiteDesignUpdateReq.SuiteDesign, suiteDesignUpdateReq.SuiteUUID)
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
