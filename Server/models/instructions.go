package models

type Command struct {
	Id        uint   `json:"id" gorm:"primary_key"`
	Command   string `json:"command" binding:"required"`
	Arguments string `json:"arguments"`
}

type CreateCommand struct {

	Id  uint `json:"id" gorm:"primary_key"`
	Command string `json:"command" binding:"required"`
	Arguments string `json:"arguments"`
}


