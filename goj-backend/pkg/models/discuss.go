package models

import (
	"time"

	"gorm.io/gorm"
)

// Discussion 讨论模型
type Discussion struct {
	ID        string `json:"id" gorm:"primarykey;type:varchar(10)"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
	UserID    uint           `gorm:"index"`
	Title     string         `gorm:"type:varchar(100)"`
	Content   string         `gorm:"type:text"`
	Category  string         `gorm:"type:varchar(20)"` // discussion, solution, notice
	Likes     int            `gorm:"default:0"`
	Comments  int            `gorm:"default:0"`
	Stars     int            `gorm:"default:0"`
}

// Comment 评论模型
type Comment struct {
	ID           uint `gorm:"primarykey"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`
	UserID       uint           `gorm:"index"`
	DiscussionID string         `gorm:"type:varchar(10);index"`
	Content      string         `gorm:"type:text"`
}

// DiscussionLike 点赞记录
type DiscussionLike struct {
	UserID       uint   `gorm:"primarykey"`
	DiscussionID string `gorm:"primarykey;type:varchar(10)"`
	CreatedAt    time.Time
}

// DiscussionStar 收藏记录
type DiscussionStar struct {
	UserID       uint   `gorm:"primarykey"`
	DiscussionID string `gorm:"primarykey;type:varchar(10)"`
	CreatedAt    time.Time
}

func (Discussion) TableName() string {
	return "discussions"
}

func (Comment) TableName() string {
	return "comments"
}

func (DiscussionLike) TableName() string {
	return "discussion_likes"
}

func (DiscussionStar) TableName() string {
	return "discussion_stars"
}
