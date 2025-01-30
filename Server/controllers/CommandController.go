package controllers


import (
     
	"net/http"
	
	"github.com/gin-gonic/gin"
	"edr/Server/models"

)



func ShowCommands(c* gin.Context) {
	var commands []models.Command

	models.DB.Find(&commands)

	c.JSON(http.StatusOK, gin.H{"data": commands})
}



func StoreCommands (c* gin.Context) {

	var input models.CreateCommand

	//Check if a value is requied but its not in the request 
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest,gin.H{"error" : err.Error()})
		return
	}

	command := models.Command{
		Command: input.Command,
		Arguments: input.Arguments,
	}

	if err:= models.DB.Create(&command).Error; err != nil{
		c.JSON(http.StatusInternalServerError, gin.H{"error":err.Error()})
		return
	}

	models.DB.Create(&command)

    c.JSON(http.StatusOK,gin.H{"message":"Command added"})
}



func RecieveOutput ( c * gin.Context ) {

	var input models.Output  

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest,gin.H{"error": err.Error()})
		return
	} 

    output := models.Output{
          AgentID: input.AgentID,
		  Given_command: input.Given_command,
		  Output: input.Output,
	}

	if err := models.DB.Create(&output).Error; err != nil {
		c.JSON(http.StatusInternalServerError,gin.H{"error":err.Error()})
		return
	}

	models.DB.Create(&output)

	c.JSON(http.StatusOK,gin.H{"message" : "output recieved"})
}  


func ShowOutput (c * gin.Context) {

	var outputs []models.Output
    
	models.DB.Find(&outputs)

	c.JSON(http.StatusOK,gin.H{"data":outputs})

}



