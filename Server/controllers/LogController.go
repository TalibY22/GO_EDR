package controllers


import (
	"net/http"
	
	"github.com/gin-gonic/gin"
	"edr/Server/models"
)





func Showlogs(c*gin.Context) {

	var logs []models.Log

	models.DB.Find(&logs)

    c.JSON(http.StatusOK, gin.H{"data": logs})
}












func StoreLogs(c*gin.Context){

	var input models.CreateLog

	if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }


	log := models.Log{

		AgentId: input.AgentId,
		Event: input.Event,
		Details: input.Details,
	}

	if err := models.DB.Create(&log).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

	models.DB.Create(&log)

	
	c.JSON(http.StatusOK, gin.H{"message":"log created"})



}