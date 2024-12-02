package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// StringArray 用于JSON序列化和反序列化字符串数组
type StringArray []string

// Value 实现 driver.Valuer 接口
func (a StringArray) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}
	return json.Marshal(a)
}

// Scan 实现 sql.Scanner 接口
func (a *StringArray) Scan(value interface{}) error {
	if value == nil {
		*a = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return json.Unmarshal([]byte(value.(string)), &a)
	}
	return json.Unmarshal(bytes, &a)
}

type Submission struct {
	ID              uint        `gorm:"primarykey;autoIncrement:100000001"`
	UserID          uint        `gorm:"index;not null"`
	Username        string      `gorm:"type:varchar(50);default:'NULL'"`
	ProblemID       string      `gorm:"index;not null"`
	ContestID       string      `gorm:"index"`
	Language        string      `gorm:"type:varchar(50);not null"`
	Code            string      `gorm:"type:text;not null"`
	Status          string      `gorm:"type:varchar(50);not null;default:Pending"`
	TimeUsed        int         `gorm:"default:0"`
	MemoryUsed      int         `gorm:"default:0"`
	ErrorInfo       string      `gorm:"type:text"`
	SubmitTime      time.Time   `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
	JudgeTime       *time.Time  `gorm:"type:timestamp;null"`
	CreatedAt       time.Time   `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt       time.Time   `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP"`
	TestcasesStatus StringArray `gorm:"type:json"`
	TestcasesInfo   StringArray `gorm:"type:json"`
	Role            string      `gorm:"type:varchar(50);not null;default:user"`
	TestCaseResults string      `gorm:"column:testcase_results;type:text"`
}

func (Submission) TableName() string {
	return "submissions"
}
