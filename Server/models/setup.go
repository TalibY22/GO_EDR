package models 


import (
	"gorm.io/gorm"
	"gorm.io/driver/sqlite"
	"fmt"
)




var DB * gorm.DB


func OpendDB(){
	database,err  := gorm.Open(sqlite.Open("logs.db"),&gorm.Config{})
	
	if err!= nil{
		fmt.Printf("Databse not working ")
	}


	err = database.AutoMigrate(&Log{},&Command{},&Output{},&Agent{},&Files{})

	if err != nil{

		fmt.Printf("error")
	}

	DB = database
}