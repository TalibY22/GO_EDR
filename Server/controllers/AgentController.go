package controllers



import (
     
	"net/http"
	
	"github.com/gin-gonic/gin"
	"edr/Server/models"

)




func ShowAgents(c* gin.Context) {
	var agents []models.Agent

	models.DB.Find(&agents)

	c.JSON(http.StatusOK, gin.H{"data": agents})
}


func StoreAgents (c* gin.Context) {

	var input models.CreateAgent

	//Check if a value is requied but its not in the request 
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest,gin.H{"error" : err.Error()})
		return
	}

	agent := models.Agent{
		Name: input.Name,
		System: input.System,

	}

	if err:= models.DB.Create(&agent).Error; err != nil{
		c.JSON(http.StatusInternalServerError, gin.H{"error":err.Error()})
		return
	}

	models.DB.Create(&agent)

    c.JSON(http.StatusOK,gin.H{"message":"Command added"})
}

