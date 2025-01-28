package models 


type Log struct{

	Id uint `json:"id" gorm:"primary_key"`
	AgentId string `json:"agent_id" binding:"required"`
	Timestamp string `json:"timestamp" binding:"required"`
	Event string `json:"event" binding:"required"`
	Details string `json:"details" binding:"required"`

}


type CreateLog struct {
	AgentId string  `json:"agent_id" binding:"required"`
	Event string  `json:"event" binding:"required"`
	Details string `json:"details" binding:"required"`
}


