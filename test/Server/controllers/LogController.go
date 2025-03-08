package controllers

import (
	"net/http"
	"strconv"

	"edr/Server/models"

	"github.com/gin-gonic/gin"
)

// Showlogs retrieves logs with pagination
func Showlogs(c *gin.Context) {
	var logs []models.Log

	// Get pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	// Calculate offset
	offset := (page - 1) * limit

	// Get event type filter if provided
	eventType := c.Query("event")

	// Get agent ID filter if provided
	agentID := c.Query("agent_id")

	// Build query
	query := models.DB

	// Apply filters if provided
	if eventType != "" {
		query = query.Where("event = ?", eventType)
	}

	if agentID != "" {
		query = query.Where("agent_id = ?", agentID)
	}

	// Get total count for pagination
	var total int64
	query.Model(&models.Log{}).Count(&total)

	// Get logs with pagination
	query.Order("id desc").Limit(limit).Offset(offset).Find(&logs)

	c.JSON(http.StatusOK, gin.H{
		"data": logs,
		"pagination": gin.H{
			"total": total,
			"page":  page,
			"limit": limit,
			"pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// StoreLogs creates a new log entry
func StoreLogs(c *gin.Context) {
	var input models.CreateLog

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log := models.Log{
		AgentId: input.AgentId,
		Event:   input.Event,
		Details: input.Details,
	}

	if err := models.DB.Create(&log).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "log created", "data": log})
}

// DeleteLog deletes a specific log entry
func DeleteLog(c *gin.Context) {
	id := c.Param("id")

	var log models.Log
	if err := models.DB.First(&log, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Log not found"})
		return
	}

	if err := models.DB.Delete(&log).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Log deleted successfully"})
}

// ClearLogs deletes all logs
func ClearLogs(c *gin.Context) {
	if err := models.DB.Exec("DELETE FROM logs").Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "All logs cleared successfully"})
}

// GetBashHistory retrieves bash history logs
func GetBashHistory(c *gin.Context) {
	var bashHistory []models.Log

	// Get pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	// Calculate offset
	offset := (page - 1) * limit

	// Get agent ID filter if provided
	agentID := c.Query("agent_id")

	// Build query
	query := models.DB.Where("event = ?", "BASH")

	// Apply agent filter if provided
	if agentID != "" {
		query = query.Where("agent_id = ?", agentID)
	}

	// Get total count for pagination
	var total int64
	query.Model(&models.Log{}).Count(&total)

	// Get bash history with pagination
	query.Order("id desc").Limit(limit).Offset(offset).Find(&bashHistory)

	c.JSON(http.StatusOK, gin.H{
		"data": bashHistory,
		"pagination": gin.H{
			"total": total,
			"page":  page,
			"limit": limit,
			"pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}
