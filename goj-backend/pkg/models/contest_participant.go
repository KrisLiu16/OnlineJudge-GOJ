package models

import (
	"time"

	"gorm.io/gorm"
)

type ContestParticipant struct {
	ID        uint `json:"id" gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
	ContestID string         `json:"contestId" gorm:"type:varchar(10);not null"`
	UserID    uint           `json:"userID" gorm:"not null"`
	Score     int            `json:"score" gorm:"default:0"`
	Rank      int            `json:"rank" gorm:"default:0"`
	Status    string         `json:"status" gorm:"type:varchar(20);default:registered"` // registered, participating, finished
}

func (ContestParticipant) TableName() string {
	return "contest_participants"
}
