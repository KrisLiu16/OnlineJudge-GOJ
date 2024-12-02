package models

import (
	"time"
)

type ProblemStatus string

const (
	StatusUnattempted ProblemStatus = "unattempted"
	StatusAttempted   ProblemStatus = "attempted"
	StatusAccepted    ProblemStatus = "accepted"
)

// 状态等级映射
var statusLevel = map[ProblemStatus]int{
	StatusUnattempted: 0,
	StatusAttempted:   1,
	StatusAccepted:    2,
}

// 判断状态是否可以覆盖
func (s ProblemStatus) CanOverride(target ProblemStatus) bool {
	return statusLevel[s] > statusLevel[target]
}

type UserProblemStatus struct {
	UserID    uint          `gorm:"primaryKey"`
	ProblemID string        `gorm:"primaryKey"`
	Status    ProblemStatus `gorm:"type:varchar(50);not null;default:unattempted"`
	UpdatedAt time.Time
}
