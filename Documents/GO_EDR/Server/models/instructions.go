package models

type Command struct {
	Id        uint   `json:"id" gorm:"primary_key"`
	AgentID   uint   `json:"agent_id" gorm:"not null"` // Foreign key to Agent
    Agent     Agent  `gorm:"foreignKey:AgentID"` 
	Command   string `json:"command" binding:"required"`
	Arguments string `json:"arguments"`
}

type CreateCommand struct {

	Id  uint `json:"id" gorm:"primary_key"`
	Command string `json:"command" binding:"required"`
	Arguments string `json:"arguments"`
}


type CreateAgentCommand struct {

	AgentID   uint   `json:"agent_id" binding:"required"` // Required agent ID when creating command
    Command   string `json:"command" binding:"required"`
    Arguments string `json:"arguments"`
}

type Output struct {

	Id uint `json:"id" gorm:"primary_key"`
	AgentID  string  `json:"agent_id" binding:"required"`
	Given_command string  `json:"given_command" `
    Output string `json:"output" binding:"required"` 
}




