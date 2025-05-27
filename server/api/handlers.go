package api

import (
	"b3/server/database"
	"b3/server/models"
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
