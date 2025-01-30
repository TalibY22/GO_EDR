package main

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"edr/Server/models"
	"edr/Server/controllers"
)

var db = make(map[string]string)

func setupRouter() *gin.Engine {
	// Disable Console Color
	// gin.DisableConsoleColor()
	r := gin.Default()

	// Ping test
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	// Get user value
	r.GET("/user/:name", func(c *gin.Context) {
		user := c.Params.ByName("name")
		value, ok := db[user]
		if ok {
			c.JSON(http.StatusOK, gin.H{"user": user, "value": value})
		} else {
			c.JSON(http.StatusOK, gin.H{"user": user, "status": "no value"})
		}
	})

	// Authorized group (uses gin.BasicAuth() middleware)
	authorized := r.Group("/", gin.BasicAuth(gin.Accounts{
		"foo":  "bar", // user:foo password:bar
		"manu": "123", // user:manu password:123
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

	return r
}





func main() {
	r := setupRouter()

	models.OpendDB() 

	
	r.GET("/logs", controllers.Showlogs)
	r.POST("/logs", controllers.StoreLogs)
	r.GET("/command",controllers.ShowCommands)
	r.POST("/command",controllers.StoreCommands)
	r.POST("/output",controllers.RecieveOutput)
	r.GET("/output",controllers.ShowOutput)


	



	r.Run(":8080") 
}
