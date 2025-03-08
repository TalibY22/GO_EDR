package controllers

import (
	"edr/Server/models"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)


var uploaddir = "./Server/Uploads"










func StoreFiles(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB) 

	
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file received"})
		return
	}

	
	if err := os.MkdirAll(uploaddir, os.ModePerm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create upload directory"})
		return
	}

	
	dst := filepath.Join(uploaddir, file.Filename)

	
	if err := c.SaveUploadedFile(file, dst); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get agent ID from request
	agentID := c.PostForm("agent")

	// Construct file URL (assuming server runs on localhost)
	fileURL := "http://localhost:8080/uploads/" + file.Filename

	// Create file record
	fileRecord := models.Files{
		Filename: file.Filename,
		FileUrl:  fileURL,
		Agent:    agentID,
	}

	// Save file metadata to database
	if err := db.Create(&fileRecord).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file metadata"})
		return
	}

	// Return success response
	c.JSON(http.StatusOK, gin.H{
		"message": "File uploaded successfully",
		"file":    file.Filename,
		"url":     fileURL,
	})
}
