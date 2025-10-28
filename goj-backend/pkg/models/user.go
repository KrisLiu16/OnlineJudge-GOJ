package models

import (
	"github.com/KrisLiu16/OnlineJudge-GOJ/goj-backend/pkg/types"
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID               uint `gorm:"primarykey"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        gorm.DeletedAt `gorm:"index"`
	Username         string         `gorm:"size:50;not null"`
	Email            string         `gorm:"size:100;not null;unique"`
	PasswordHash     string         `gorm:"size:255;not null"`
	Avatar           string         `gorm:"size:255"`
	Bio              string         `gorm:"type:text"`
	Role             string         `gorm:"size:20;default:user"`
	Submissions      int64          `gorm:"default:0"`
	AcceptedProblems int64          `gorm:"default:0"`
	Rating           int64          `gorm:"default:1500"`
	AcceptedCount    uint           `gorm:"default:0" json:"acceptedCount"`
}

func (User) TableName() string {
	return "users"
}

func (u *User) ToDTO() *types.User {
	return &types.User{
		ID:               u.ID,
		Username:         u.Username,
		Email:            u.Email,
		Avatar:           u.Avatar,
		Bio:              u.Bio,
		Role:             u.Role,
		Submissions:      u.Submissions,
		AcceptedProblems: u.AcceptedProblems,
		Rating:           u.Rating,
		CreatedAt:        u.CreatedAt,
		UpdatedAt:        u.UpdatedAt,
	}
}
