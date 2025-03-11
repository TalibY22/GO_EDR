package controllers

import (
	"edr/Server/models"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// ShowAgents retrieves all agents from the database
func ShowAgents(c *gin.Context) {
	var agents []models.Agent

	models.DB.Find(&agents)

	c.JSON(http.StatusOK, gin.H{"data": agents})
}

// StoreAgents creates a new agent in the database
func StoreAgents(c *gin.Context) {
	var input models.CreateAgent

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create agent
	agent := models.Agent{
		Name:        input.Name,
		OS:          input.OS,
		Status:      input.Status,
		Description: input.Description,
		LastSeen:    time.Now(),
	}

	if err := models.DB.Create(&agent).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Compile the agent binary
	if err := compileAgentBinary(agent); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Agent created but binary compilation failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": agent})
}

// DeleteAgent deletes an agent from the database
func DeleteAgent(c *gin.Context) {
	id := c.Param("id")

	var agent models.Agent
	if err := models.DB.First(&agent, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Agent not found"})
		return
	}

	if err := models.DB.Delete(&agent).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Delete the agent binary if it exists
	binaryPath := filepath.Join("./Server/Binaries", "edr-agent-"+id)
	if _, err := os.Stat(binaryPath); err == nil {
		os.Remove(binaryPath)
	}

	c.JSON(http.StatusOK, gin.H{"data": true})
}

// DownloadAgent allows downloading the agent binary
func DownloadAgent(c *gin.Context) {
	id := c.Param("id")

	var agent models.Agent
	if err := models.DB.First(&agent, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Agent not found"})
		return
	}

	// Check if binary exists
	binaryPath := filepath.Join("./Server/Binaries", "edr-agent-"+id)
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		// Try to compile it if it doesn't exist
		if err := compileAgentBinary(agent); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to compile agent binary: " + err.Error()})
			return
		}
	}

	// Set appropriate headers for binary download
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", "attachment; filename=edr-agent-"+agent.Name)
	c.Header("Content-Type", "application/octet-stream")

	// Serve the file
	c.File(binaryPath)
}

// Helper function to compile the agent binary
func compileAgentBinary(agent models.Agent) error {
	// Create binaries directory if it doesn't exist
	binariesDir := "./Server/Binaries"
	if err := os.MkdirAll(binariesDir, os.ModePerm); err != nil {
		return err
	}

	// Determine output file name
	outputFile := filepath.Join(binariesDir, "edr-agent-"+strconv.Itoa(int(agent.ID)))

	// Set up the build command based on the target OS
	var cmd *exec.Cmd
	switch agent.OS {
	case "linux":
		cmd = exec.Command("go", "build", "-o", outputFile, "./Agent/LinuxAgent.go")
	case "windows":
		cmd = exec.Command("go", "build", "-o", outputFile+".exe", "./Agent/LinuxAgent.go") // Replace with WindowsAgent.go when available
	case "macos":
		cmd = exec.Command("go", "build", "-o", outputFile, "./Agent/LinuxAgent.go") // Replace with MacAgent.go when available
	default:
		return nil // Skip compilation for unknown OS
	}

	// Set environment variables for cross-compilation if needed
	env := os.Environ()
	if agent.OS == "windows" {
		env = append(env, "GOOS=windows")
	} else if agent.OS == "macos" {
		env = append(env, "GOOS=darwin")
	}
	cmd.Env = env

	// Run the build command
	output, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}

	// Log the output for debugging
	if len(output) > 0 {
		// Log the output somewhere
	}

	return nil
}
