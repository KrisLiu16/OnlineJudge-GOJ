package models

import (
	"time"
)

type RatingHistory struct {
	ID        uint      `gorm:"primarykey;autoIncrement"`
	UserID    uint      `gorm:"index;not null"`
	ContestID string    `gorm:"index;not null"`
	OldRating int64     `gorm:"not null"`
	NewRating int64     `gorm:"not null"`
	Rank      int       `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

func (RatingHistory) TableName() string {
	return "rating_histories"
}
