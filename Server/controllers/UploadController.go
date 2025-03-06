package controllers



import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

//Location of where the extracted file will be stored 
var uploaddir = ".Server/Uploads"

func storefiles(c*gin.Context){

     file,err := c.FormFile("file");


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

	

}