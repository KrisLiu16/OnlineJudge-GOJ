package models

import (
	"time"

	"gorm.io/gorm"
)

type Problem struct {
	ID              string `json:"id" gorm:"primarykey;type:varchar(10)"` // 5位数字编号
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       gorm.DeletedAt `gorm:"index"`
	Title           string         `json:"title" gorm:"type:varchar(100);not null"`
	Difficulty      int            `json:"difficulty" gorm:"type:tinyint;not null"`          // 1-5 表示难度等级
	Role            string         `json:"role" gorm:"type:varchar(20);default:public"`      // public, private, contest
	Tag             string         `json:"tag" gorm:"type:varchar(50)"`                      // 题目标签,如 dp,greedy 等
	AcceptedCount   int64          `json:"acceptedCount" gorm:"default:0"`                   // 通过次数
	SubmissionCount int64          `json:"submissionCount" gorm:"default:0"`                 // 提交次数
	Source          string         `json:"source" gorm:"type:varchar(100)"`                  // 题目来源
	Languages       string         `json:"languages" gorm:"type:varchar(100)"`               // 支持的编程语言,如 "c,cpp,java,python"
	TimeLimit       int64          `json:"timeLimit" gorm:"type:int;not null;default:1000"`  // 时间限制,单位ms
	MemoryLimit     int64          `json:"memoryLimit" gorm:"type:int;not null;default:128"` // 内存限制,单位MB
	UseSPJ          bool           `json:"useSPJ" gorm:"type:tinyint;not null;default:0"`    // 是否使用SPJ
}

func (Problem) TableName() string {
	return "problems"
}
