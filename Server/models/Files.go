package models



type Files struct {

	ID uint `json:"id" gorm:"primary_key"`
	Filename string 
    FileUrl string 
    Agent string 
}


