package handler

import (
	"encoding/json"
	"fmt"
	"goj/pkg/config"
	"goj/pkg/judge/types"
	"goj/pkg/models"
	"log"
	"net/http"
	"strings"
	"time"

	"gorm.io/gorm"
)

// 添加日志级别常量
const (
	LogLevelDebug = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
)

// 当前日志级别
var logLevel = LogLevelInfo

// 添加日志函数
func logDebug(format string, v ...interface{}) {
	if logLevel <= LogLevelDebug {
		log.Printf(format, v...)
	}
}

func logError(format string, v ...interface{}) {
	if logLevel <= LogLevelError {
		log.Printf("\033[31m"+format+"\033[0m", v...)
	}
}

type ResultHandler struct {
	db        *gorm.DB
	ws        *WebSocketManager
	judgeAddr string   // 添加评测机地址
	cachedIds []string // 添加缓存ID列表
}

func NewResultHandler(ws *WebSocketManager, judgeAddr string) *ResultHandler {
	return &ResultHandler{
		db:        config.DB,
		ws:        ws,
		judgeAddr: judgeAddr,
	}
}

func (h *ResultHandler) HandleResult(result *types.JudgeResult) error {
	defer h.cleanupCache() // 确保在处理完后清理缓存

	logDebug("[ResultHandler] Processing result for submission %d", result.ID)
	tx := h.db.Begin()
	if tx.Error != nil {
		logError("[ResultHandler] Failed to begin transaction: %v", tx.Error)
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var submission models.Submission
	if err := tx.First(&submission, result.ID).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 如果是比赛提交，尝试处理比赛相关数据
	if submission.ContestID != "" {
		isValidContestSubmission := false
		// 检查比赛状态和题目
		var contest models.Contest
		if err := tx.Where("id = ?", submission.ContestID).First(&contest).Error; err == nil {
			now := time.Now()
			if now.After(contest.StartTime) && now.Before(contest.EndTime) {
				// 检查题目是否在比赛中
				problemList := strings.Split(contest.Problems, ",")
				for _, pid := range problemList {
					if pid == submission.ProblemID {
						// 先更新 submission 的 role
						if err := tx.Model(&submission).Update("role", "admin").Error; err != nil {
							logError("更新提交角色失败: %v", err)
							break
						}

						// 更新比赛提交状态
						if err := h.processContestSubmission(tx, &submission); err != nil {
							logError("处理比赛提交失败: %v", err)
							break
						}

						// 所有操作都成功才标记为有效
						isValidContestSubmission = true
						break
					}
				}
			}
		}

		// 如果不是有效的比赛提交，清除比赛ID和role
		if !isValidContestSubmission {
			if err := tx.Model(&submission).Updates(map[string]interface{}{
				"contest_id": nil,
				"role":       "user",
			}).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("清除比赛信息失败: %v", err)
			}
		}
	}

	// 1. 更新主提交记录
	if err := h.updateSubmission(tx, &submission, result); err != nil {
		tx.Rollback()
		return err
	}

	// 2. 更新题目统计
	if err := h.updateProblemStats(tx, &submission, result.Status); err != nil {
		tx.Rollback()
		return err
	}

	// 3. 更新用户统计
	if err := h.updateUserStats(tx, submission.UserID); err != nil {
		tx.Rollback()
		return err
	}

	// 4. 更新用户题目状态
	if err := h.updateUserProblemStatus(tx, submission.UserID, submission.ProblemID, result.Status); err != nil {
		tx.Rollback()
		return err
	}

	// // 推送评测状态更新
	// if h.ws != nil {
	// 	msg := map[string]interface{}{
	// 		"type": "judge_status",
	// 		"data": map[string]interface{}{
	// 			"id":     result.ID,
	// 			"status": result.Status,
	// 		},
	// 	}
	// 	if err := h.ws.BroadcastToUser(result.UserID, msg); err != nil {
	// 		logError("[ResultHandler] Failed to broadcast status: %v", err)
	// 	}
	// }

	// // 推送最终结果
	// if h.ws != nil {
	// 	msg := map[string]interface{}{
	// 		"type": "judge_result",
	// 		"data": result,
	// 	}
	// 	if err := h.ws.BroadcastToUser(result.UserID, msg); err != nil {
	// 		logError("[ResultHandler] Failed to broadcast result: %v", err)
	// 	}
	// }

	return tx.Commit().Error
}

// processContestSubmission 处理比赛提交
func (h *ResultHandler) processContestSubmission(tx *gorm.DB, submission *models.Submission) error {
	logDebug("[ResultHandler] Processing contest submission %d for contest %s",
		submission.ID, submission.ContestID)
	// 使用临时表或子查询来避免 UPDATE 和 FROM 同时使用同一个表
	query := `
		INSERT INTO contest_submission_status (contest_id, submission_ids, updated_at)
		WITH current_status AS (
			SELECT submission_ids 
			FROM contest_submission_status 
			WHERE contest_id = ?
		)
		SELECT ?, 
			   COALESCE(
				   (SELECT JSON_ARRAY_APPEND(submission_ids, '$', ?) FROM current_status),
				   JSON_ARRAY(?)
			   ),
			   ?
		ON DUPLICATE KEY UPDATE
		submission_ids = VALUES(submission_ids),
		updated_at = VALUES(updated_at)
	`

	return tx.Exec(query,
		submission.ContestID,
		submission.ContestID,
		submission.ID,
		submission.ID,
		time.Now(),
	).Error
}

// updateSubmission 更新主提交记录
func (h *ResultHandler) updateSubmission(tx *gorm.DB, submission *models.Submission, result *types.JudgeResult) error {
	logDebug("[ResultHandler] Updating submission %d with status %s", submission.ID, result.Status)
	now := time.Now()
	updates := map[string]interface{}{
		"status":      result.Status,
		"time_used":   result.TimeUsed,
		"memory_used": result.MemoryUsed,
		"error_info":  result.ErrorInfo,
		"judge_time":  &now,
	}

	// 将测试点结果转换为JSON
	if len(result.TestCaseResults) > 0 {
		testCaseResultsJson, err := json.Marshal(result.TestCaseResults)
		if err == nil {
			updates["testcase_results"] = string(testCaseResultsJson)
		}
	}

	// 兼容旧版
	if len(result.TestcasesStatus) > 0 {
		statusJson, err := json.Marshal(result.TestcasesStatus)
		if err == nil {
			updates["testcases_status"] = string(statusJson)
		}
	}
	if len(result.TestCasesInfo) > 0 {
		infoJson, err := json.Marshal(result.TestCasesInfo)
		if err == nil {
			updates["testcases_info"] = string(infoJson)
		}
	}

	return tx.Model(&submission).Updates(updates).Error
}

// updateProblemStats 更新题目统计信息
func (h *ResultHandler) updateProblemStats(tx *gorm.DB, submission *models.Submission, status string) error {
	logDebug("[ResultHandler] Updating stats for problem %s", submission.ProblemID)
	return tx.Model(&models.Problem{}).
		Where("id = ?", submission.ProblemID).
		Updates(map[string]interface{}{
			"submission_count": gorm.Expr("submission_count + ?", 1),
			"accepted_count":   gorm.Expr("CASE WHEN ? = 'Accepted' THEN accepted_count + 1 ELSE accepted_count END", status),
		}).Error
}

// updateUserStats 更新用户统计信息
func (h *ResultHandler) updateUserStats(tx *gorm.DB, userID uint) error {
	logDebug("[ResultHandler] Updating stats for user %d", userID)
	return tx.Exec(`
		UPDATE users 
		SET 
			submissions = (
				SELECT COUNT(*) 
				FROM submissions 
				WHERE user_id = ?
			),
			accepted_problems = (
				SELECT COUNT(DISTINCT problem_id) 
				FROM submissions 
				WHERE user_id = ? AND status = 'Accepted'
			)
		WHERE id = ?
	`, userID, userID, userID).Error
}

// updateUserProblemStatus 更新用户题目状态
func (h *ResultHandler) updateUserProblemStatus(tx *gorm.DB, userID uint, problemID string, status string) error {
	var currentStatus models.UserProblemStatus
	err := tx.Where("user_id = ? AND problem_id = ?", userID, problemID).
		First(&currentStatus).Error

	var newStatus models.ProblemStatus
	if err == nil {
		if currentStatus.Status == models.StatusAccepted {
			newStatus = models.StatusAccepted
		} else {
			if status == types.StatusAccepted {
				newStatus = models.StatusAccepted
			} else {
				newStatus = models.StatusAttempted
			}
		}
	} else if err == gorm.ErrRecordNotFound {
		if status == types.StatusAccepted {
			newStatus = models.StatusAccepted
		} else {
			newStatus = models.StatusAttempted
		}
	} else {
		return err
	}

	return tx.Model(&models.UserProblemStatus{}).
		Where("user_id = ? AND problem_id = ?", userID, problemID).
		Assign(map[string]interface{}{
			"user_id":    userID,
			"problem_id": problemID,
			"status":     newStatus,
			"updated_at": time.Now(),
		}).
		FirstOrCreate(&models.UserProblemStatus{}).Error
}

// 添加缓存ID
func (h *ResultHandler) AddCachedId(id string) {
	h.cachedIds = append(h.cachedIds, id)
}

// 清理缓存
func (h *ResultHandler) cleanupCache() {
	for _, id := range h.cachedIds {
		if err := h.deleteCachedFile(id); err != nil {
			logError("[ResultHandler] Failed to delete cached file %s: %v", id, err)
		}
	}
	h.cachedIds = nil
}

// 删除单个缓存文件
func (h *ResultHandler) deleteCachedFile(id string) error {
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodDelete, h.judgeAddr+"/file/"+id, nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}
