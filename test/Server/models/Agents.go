package models


import (
	"time"
)
type Agent struct {
	ID          uint      `json:"id" gorm:"primary_key"`
	Name        string    `json:"name" gorm:"not null"`
	OS          string    `json:"os" gorm:"not null"`
	Status      string    `json:"status" gorm:"not null"`
	LastSeen    time.Time `json:"last_seen"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// CreateAgent is the input struct for creating a new agent
type CreateAgent struct {
	Name        string `json:"name" binding:"required"`
	OS          string `json:"os" binding:"required"`
	Status      string `json:"status" binding:"required"`
	Description string `json:"description"`
} 