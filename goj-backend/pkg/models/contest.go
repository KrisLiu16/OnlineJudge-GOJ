package models

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

type Contest struct {
	ID               string `json:"id" gorm:"primarykey;type:varchar(10)"` // 5位数字编号
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        gorm.DeletedAt `gorm:"index"`
	Title            string         `json:"title" gorm:"type:varchar(100);not null"`
	Description      string         `json:"description" gorm:"type:text"`
	StartTime        time.Time      `json:"startTime" gorm:"type:datetime;not null"`
	EndTime          time.Time      `json:"endTime" gorm:"type:datetime;not null"`
	Role             string         `json:"role" gorm:"type:varchar(20);default:public"` // public, private
	Status           string         `json:"status" gorm:"type:varchar(20)"`              // not_started, running, ended
	ParticipantCount int64          `json:"participantCount" gorm:"default:0"`
	Problems         string         `json:"problems" gorm:"type:text"`     // 题目ID列表，用逗号分隔
	PenaltyTime      int            `json:"penaltyTime" gorm:"default:20"` // 罚时（分钟）
}

func (Contest) TableName() string {
	return "contests"
}

// 添加自定义 JSON 序列化方法
func (c Contest) MarshalJSON() ([]byte, error) {
	type Alias Contest
	return json.Marshal(&struct {
		Alias
		StartTime string `json:"startTime"`
		EndTime   string `json:"endTime"`
	}{
		Alias:     Alias(c),
		StartTime: c.StartTime.UTC().Format(time.RFC3339),
		EndTime:   c.EndTime.UTC().Format(time.RFC3339),
	})
}
