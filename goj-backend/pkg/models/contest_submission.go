package models

import (
	"time"
)

// ContestSubmissionStatus 比赛提交记录
type ContestSubmissionStatus struct {
	ContestID     string    `gorm:"primarykey"` // 比赛ID作为主键
	SubmissionIDs string    `gorm:"type:text"`  // 提交ID列表，JSON字符串
	UpdatedAt     time.Time `gorm:"type:timestamp;not null"`
}

func (ContestSubmissionStatus) TableName() string {
	return "contest_submission_status"
}
