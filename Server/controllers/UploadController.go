package controllers

import (
	"edr/Server/models"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"mime/multipart"

	"github.com/gin-gonic/gin"
)

var uploaddir = "./Server/Uploads"

func StoreFiles(c *gin.Context) {
	// Try to get the file from either "file" or "screenshot" field
	var file *multipart.FileHeader
	var err error

	file, err = c.FormFile("file")
	if err != nil {
		// If "file" field doesn't exist, try "screenshot" field
		file, err = c.FormFile("screenshot")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No file received: " + err.Error()})
			return
		}
	}

	// Get agent ID from request - try both "agent" and "agent_id" fields
	agentID := c.PostForm("agent")
	if agentID == "" {
		agentID = c.PostForm("agent_id")
		if agentID == "" {
			// Default to "unknown" if no agent ID is provided
			agentID = "unknown"
		}
	}

	// Create upload directory if it doesn't exist
	if err := os.MkdirAll(uploaddir, os.ModePerm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create upload directory: " + err.Error()})
		return
	}

	// Generate a unique filename with timestamp
	timestamp := time.Now().Format("20060102-150405")
	filename := filepath.Base(file.Filename)

	// If it's a screenshot, use a more descriptive name
	if _, err := c.FormFile("screenshot"); err == nil {
		filename = "screenshot-" + agentID + "-" + timestamp + filepath.Ext(file.Filename)
	} else {
		// For regular files, just add timestamp to avoid overwriting
		ext := filepath.Ext(filename)
		basename := filename[:len(filename)-len(ext)]
		filename = basename + "-" + timestamp + ext
	}

	dst := filepath.Join(uploaddir, filename)

	// Save the file
	if err := c.SaveUploadedFile(file, dst); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file: " + err.Error()})
		return
	}

	// Construct file URL
	fileURL := "/uploads/" + filename

	// Create file record
	fileRecord := models.Files{
		Filename: filename,
		FileUrl:  fileURL,
		Agent:    agentID,
	}

	// Save file metadata to database using the global DB instance
	if err := models.DB.Create(&fileRecord).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file metadata: " + err.Error()})
		return
	}

	// Return success response
	c.JSON(http.StatusOK, gin.H{
		"message": "File uploaded successfully",
		"file":    filename,
		"url":     fileURL,
	})
}
