package models


type Agent struct {
    Id uint `json:"id" gorm:"primary_key"`
	AgentID  string `json:"agent_id" gorm:"unique"`
	Name string `json:"name"`
	System string  `json:"string"`

}


type CreateAgent struct {

	Id uint `json:"id" gorm:"primary_key"`
	Name string `json:"name"`
	System string  `json:"string"`
}
