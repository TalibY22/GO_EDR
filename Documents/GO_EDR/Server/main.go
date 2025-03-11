package main

import (
	"edr/Server/controllers"
	"edr/Server/models"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var db = make(map[string]string)

func setupRouter() *gin.Engine {
	r := gin.Default()

	// Configure CORS - This must come BEFORE routes are defined
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:5173"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	config.AllowHeaders = []string{
		"Origin",
		"Content-Length",
		"Content-Type",
		"Accept",
		"Authorization",
	}
	config.AllowCredentials = true
	r.Use(cors.New(config))

	// Ping test
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	// User routes
	r.GET("/user/:name", func(c *gin.Context) {
		user := c.Params.ByName("name")
		value, ok := db[user]
		if ok {
			c.JSON(http.StatusOK, gin.H{"user": user, "value": value})
		} else {
			c.JSON(http.StatusOK, gin.H{"user": user, "status": "no value"})
		}
	})

	// Authorized group
	authorized := r.Group("/", gin.BasicAuth(gin.Accounts{
		"foo":  "bar",
		"manu": "123",
	}))

	authorized.POST("admin", func(c *gin.Context) {
		user := c.MustGet(gin.AuthUserKey).(string)
		var json struct {
			Value string `json:"value" binding:"required"`
		}
		if c.Bind(&json) == nil {
			db[user] = json.Value
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		}
	})

	// API routes
	r.GET("/logs", controllers.Showlogs)
	r.GET("/logs/latest", controllers.GetLatestLogs)
	r.POST("/logs", controllers.StoreLogs)
	r.DELETE("/logs/:id", controllers.DeleteLog)
	r.DELETE("/logs/clear", controllers.ClearLogs)
	r.GET("/bash-history", controllers.GetBashHistory)
	r.GET("/command", controllers.ShowCommands)
	r.POST("/command", controllers.StoreCommands)
	r.POST("/output", controllers.RecieveOutput)
	r.GET("/output", controllers.ShowOutput)
	r.GET("/agents", controllers.ShowAgents)
	r.POST("/agents", controllers.StoreAgents)
	r.DELETE("/agents/:id", controllers.DeleteAgent)
	r.GET("/agents/:id/download", controllers.DownloadAgent)
	r.POST("/upload", controllers.StoreFiles)
	//r.GET("/files", controllers.ShowFiles)

	return r
}

func main() {
	r := setupRouter()
	models.OpendDB()
	r.Run(":8080")
}
