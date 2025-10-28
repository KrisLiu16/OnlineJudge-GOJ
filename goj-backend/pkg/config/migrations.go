package config

import (
	"fmt"
	"github.com/KrisLiu16/OnlineJudge-GOJ/goj-backend/pkg/models"
	"log"
)

func AutoMigrate() error {
	// 自动迁移所有模型
	if err := DB.AutoMigrate(
		&models.User{},
		&models.Problem{},
		&models.Submission{},
		&models.Contest{},
		&models.ContestParticipant{},
		&models.Discussion{},
		&models.Comment{},
		&models.UserProblemStatus{},
		&models.ContestSubmissionStatus{},
		&models.DiscussionLike{},
		&models.DiscussionStar{},
		&models.RatingHistory{},
		&models.WebsiteSetting{},
	); err != nil {
		log.Printf("Failed to migrate database: %v", err)
		return err
	}

	// 添加新的列
	if err := DB.AutoMigrate(&models.Submission{}); err != nil {
		return fmt.Errorf("failed to migrate database: %v", err)
	}

	// 检查列是否存在，如果不存在则添加
	if !DB.Migrator().HasColumn(&models.Submission{}, "testcases_status") {
		if err := DB.Migrator().AddColumn(&models.Submission{}, "testcases_status"); err != nil {
			return fmt.Errorf("failed to add testcases_status column: %v", err)
		}
	}

	if !DB.Migrator().HasColumn(&models.Submission{}, "testcases_info") {
		if err := DB.Migrator().AddColumn(&models.Submission{}, "testcases_info"); err != nil {
			return fmt.Errorf("failed to add testcases_info column: %v", err)
		}
	}

	// 创建 discussion_seq 表
	if err := DB.Exec(`
		CREATE TABLE IF NOT EXISTS discussion_seq (
			id BIGINT NOT NULL AUTO_INCREMENT,
			PRIMARY KEY (id)
		) ENGINE=InnoDB;
	`).Error; err != nil {
		return fmt.Errorf("failed to create discussion_seq table: %v", err)
	}

	log.Println("Database migration completed successfully")
	return nil
}
