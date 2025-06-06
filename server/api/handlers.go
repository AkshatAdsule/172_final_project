package api

import (
	"b3/server/database"
	"b3/server/models"
	"b3/server/mqttsubscriber"
	"b3/server/ride"
	"database/sql"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// RegisterRideHandlers sets up the ride-related API routes.
func RegisterRideHandlers(router *gin.RouterGroup, db *sql.DB) {
	router.GET("/rides", func(c *gin.Context) { getRidesListHandler(c, db) })
	router.GET("/rides/:id", func(c *gin.Context) { getRideDetailHandler(c, db) })
}

// RegisterLockHandlers sets up the lock-related API routes.
func RegisterLockHandlers(router *gin.RouterGroup, rideManager *ride.RideManager, publisher *mqttsubscriber.Publisher) {
	router.POST("/setLockStatus", func(c *gin.Context) { setLockStatusHandler(c, rideManager, publisher) })
	router.GET("/getLockStatus", func(c *gin.Context) { getLockStatusHandler(c, rideManager) })
}

func getRidesListHandler(c *gin.Context, db *sql.DB) {
	// Parse pagination parameters
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")
	dateStr := c.Query("date") // Optional date filter in YYYY-MM-DD format

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 10
	}

	var dateFilter *time.Time
	if dateStr != "" {
		parsedDate, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format. Use YYYY-MM-DD"})
			return
		}
		dateFilter = &parsedDate
	}

	rides, err := database.GetAllRidesSummaryWithPagination(db, page, limit, dateFilter)
	if err != nil {
		log.Printf("Error fetching ride summaries: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve rides"})
		return
	}
	if rides == nil {
		rides = []models.RideSummary{}
	}
	c.JSON(http.StatusOK, rides)
}

func getRideDetailHandler(c *gin.Context, db *sql.DB) {
	idStr := c.Param("id")
	rideID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ride ID format"})
		return
	}

	rideDetail, err := database.GetRideDetails(db, rideID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Ride not found"})
		} else {
			log.Printf("Error fetching ride detail for ID %d: %v", rideID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve ride details"})
		}
		return
	}
	c.JSON(http.StatusOK, rideDetail)
}

// LockStatusRequest represents the request body for setting lock status
type LockStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

// LockStatusResponse represents the response for lock status operations
type LockStatusResponse struct {
	Status string `json:"status"`
}

func setLockStatusHandler(c *gin.Context, rideManager *ride.RideManager, publisher *mqttsubscriber.Publisher) {
	var request LockStatusRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// Validate lock status
	if request.Status != "LOCKED" && request.Status != "UNLOCKED" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid lock status. Must be 'LOCKED' or 'UNLOCKED'"})
		return
	}

	// Update the lock status in the ride manager
	rideManager.SetLockStatus(request.Status)

	// Publish the update to the IoT shadow
	if publisher != nil {
		if err := publisher.UpdateLockStatus(request.Status); err != nil {
			log.Printf("Failed to publish lock status to IoT shadow: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update IoT shadow"})
			return
		}
	}

	log.Printf("Lock status updated to: %s", request.Status)
	c.JSON(http.StatusOK, LockStatusResponse{Status: request.Status})
}

func getLockStatusHandler(c *gin.Context, rideManager *ride.RideManager) {
	status := rideManager.GetLockStatus()
	c.JSON(http.StatusOK, LockStatusResponse{Status: status})
}
