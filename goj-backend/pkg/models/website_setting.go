package models

import (
	"time"
)

type WebsiteSetting struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	Title     string `json:"title" gorm:"type:varchar(100)"`
	Subtitle  string `json:"subtitle" gorm:"type:varchar(200)"`
	ICP       string `json:"icp" gorm:"type:varchar(50)"`
	ICPLink   string `json:"icpLink" gorm:"type:varchar(200)"`
	About     string `json:"about" gorm:"type:text"`
	Email     string `json:"email" gorm:"type:varchar(100)"`
	Github    string `json:"github" gorm:"type:varchar(100)"`
	Feature1  string `json:"feature1" gorm:"type:text"`
	Feature2  string `json:"feature2" gorm:"type:text"`
	Feature3  string `json:"feature3" gorm:"type:text"`
}

func (WebsiteSetting) TableName() string {
	return "website_settings"
}
