package controllers

import (
	"fmt"
	"goj/pkg/config"
	"goj/pkg/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// 获取讨论列表
func GetDiscussions(c *gin.Context) {
	var discussions []models.Discussion
	result := config.DB.Order("created_at desc").Find(&discussions)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取讨论列表失败"})
		return
	}

	// 获取每个讨论的作者信息和统计数据
	type DiscussionResponse struct {
		models.Discussion
		Author struct {
			ID       uint   `json:"id"`
			Username string `json:"username"`
			Avatar   string `json:"avatar"`
		} `json:"author"`
		Stats struct {
			Likes    int `json:"likes"`
			Comments int `json:"comments"`
			Stars    int `json:"stars"`
		} `json:"stats"`
	}

	var response []DiscussionResponse
	for _, d := range discussions {
		var user models.User
		config.DB.First(&user, d.UserID)

		var dr DiscussionResponse
		dr.Discussion = d
		dr.Author.ID = user.ID
		dr.Author.Username = user.Username
		dr.Author.Avatar = user.Avatar
		dr.Stats.Likes = d.Likes
		dr.Stats.Comments = d.Comments
		dr.Stats.Stars = d.Stars

		response = append(response, dr)
	}

	c.JSON(http.StatusOK, gin.H{"discussions": response})
}

// 获取讨论详情
func GetDiscussion(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("userID") // 获取当前用户ID，即使未登录也继续

	var discussion models.Discussion
	if err := config.DB.First(&discussion, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "讨论不存在"})
		return
	}

	var user models.User
	config.DB.First(&user, discussion.UserID)

	// 获取用户互动状态
	var isLiked, isStarred bool
	if userID != nil {
		var like models.DiscussionLike
		if err := config.DB.Where("user_id = ? AND discussion_id = ?", userID, id).First(&like).Error; err == nil {
			isLiked = true
		}

		var star models.DiscussionStar
		if err := config.DB.Where("user_id = ? AND discussion_id = ?", userID, id).First(&star).Error; err == nil {
			isStarred = true
		}
	}

	response := gin.H{
		"id":        discussion.ID,
		"title":     discussion.Title,
		"content":   discussion.Content,
		"category":  discussion.Category,
		"createdAt": discussion.CreatedAt,
		"author": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"avatar":   user.Avatar,
			"role":     user.Role,
		},
		"stats": gin.H{
			"likes":    discussion.Likes,
			"comments": discussion.Comments,
			"stars":    discussion.Stars,
		},
		"interactions": gin.H{
			"isLiked":   isLiked,
			"isStarred": isStarred,
		},
	}

	c.JSON(http.StatusOK, response)
}

// 创建讨论
func CreateDiscussion(c *gin.Context) {
	var input struct {
		Title    string `json:"title" binding:"required"`
		Content  string `json:"content" binding:"required"`
		Category string `json:"category" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	// 开启事务
	tx := config.DB.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建讨论失败"})
		return
	}

	// 获取新的讨论ID
	var seq struct {
		ID uint
	}

	// 先插入序列号
	if err := tx.Exec("INSERT INTO discussion_seq VALUES (NULL)").Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建讨论失败"})
		return
	}

	// 然后获取最后插入的ID
	if err := tx.Raw("SELECT LAST_INSERT_ID() as id").Scan(&seq).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建讨论失败"})
		return
	}

	// 创建讨论
	discussion := models.Discussion{
		ID:       fmt.Sprintf("%d", seq.ID), // 使用生成的ID
		UserID:   userID.(uint),
		Title:    input.Title,
		Content:  input.Content,
		Category: input.Category,
	}

	if err := tx.Create(&discussion).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建讨论失败"})
		return
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建讨论失败"})
		return
	}

	c.JSON(http.StatusCreated, discussion)
}

// 更新讨论
func UpdateDiscussion(c *gin.Context) {
	id := c.Param("id")
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	var discussion models.Discussion
	if err := config.DB.First(&discussion, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "讨论不存在"})
		return
	}

	if discussion.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权修改"})
		return
	}

	var input struct {
		Title    string `json:"title"`
		Content  string `json:"content"`
		Category string `json:"category"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	config.DB.Model(&discussion).Updates(models.Discussion{
		Title:    input.Title,
		Content:  input.Content,
		Category: input.Category,
	})

	c.JSON(http.StatusOK, discussion)
}

// 删除讨论
func DeleteDiscussion(c *gin.Context) {
	id := c.Param("id")
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	var discussion models.Discussion
	if err := config.DB.First(&discussion, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "讨论不存在"})
		return
	}

	if discussion.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权删除"})
		return
	}

	config.DB.Delete(&discussion)
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// 获取评论列表
func GetComments(c *gin.Context) {
	discussionID := c.Param("id")

	var comments []models.Comment
	if err := config.DB.Where("discussion_id = ?", discussionID).Find(&comments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取评论失败"})
		return
	}

	var response []gin.H
	for _, comment := range comments {
		var user models.User
		config.DB.First(&user, comment.UserID)

		response = append(response, gin.H{
			"id":        comment.ID,
			"content":   comment.Content,
			"createdAt": comment.CreatedAt,
			"author": gin.H{
				"id":       user.ID,
				"username": user.Username,
				"avatar":   user.Avatar,
			},
		})
	}

	c.JSON(http.StatusOK, gin.H{"comments": response})
}

// 创建评论
func CreateComment(c *gin.Context) {
	discussionID := c.Param("id")
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	var input struct {
		Content string `json:"content" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	comment := models.Comment{
		UserID:       userID.(uint),
		DiscussionID: discussionID,
		Content:      input.Content,
	}

	err := config.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&comment).Error; err != nil {
			return err
		}

		// 更新讨论的评论数
		if err := tx.Model(&models.Discussion{}).Where("id = ?", discussionID).
			UpdateColumn("comments", gorm.Expr("comments + ?", 1)).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建评论失败"})
		return
	}

	c.JSON(http.StatusCreated, comment)
}

// 点赞/取消点赞
func ToggleLike(c *gin.Context) {
	discussionID := c.Param("id")
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	var like models.DiscussionLike
	result := config.DB.Where("user_id = ? AND discussion_id = ?", userID, discussionID).First(&like)

	var currentStatus bool
	err := config.DB.Transaction(func(tx *gorm.DB) error {
		if result.Error == gorm.ErrRecordNotFound {
			// 创建点赞
			like = models.DiscussionLike{
				UserID:       userID.(uint),
				DiscussionID: discussionID,
			}
			if err := tx.Create(&like).Error; err != nil {
				return err
			}
			currentStatus = true

			// 更新点赞数
			if err := tx.Model(&models.Discussion{}).Where("id = ?", discussionID).
				UpdateColumn("likes", gorm.Expr("likes + ?", 1)).Error; err != nil {
				return err
			}
		} else {
			// 取消点赞
			if err := tx.Delete(&like).Error; err != nil {
				return err
			}

			// 更新点赞数
			if err := tx.Model(&models.Discussion{}).Where("id = ?", discussionID).
				UpdateColumn("likes", gorm.Expr("likes - ?", 1)).Error; err != nil {
				return err
			}
			currentStatus = false
		}
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "操作成功",
		"isLiked": currentStatus,
	})
}

// 收藏/取消收藏
func ToggleStar(c *gin.Context) {
	discussionID := c.Param("id")
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	var star models.DiscussionStar
	result := config.DB.Where("user_id = ? AND discussion_id = ?", userID, discussionID).First(&star)

	var currentStatus bool
	err := config.DB.Transaction(func(tx *gorm.DB) error {
		if result.Error == gorm.ErrRecordNotFound {
			// 创建收藏
			star = models.DiscussionStar{
				UserID:       userID.(uint),
				DiscussionID: discussionID,
			}
			if err := tx.Create(&star).Error; err != nil {
				return err
			}
			currentStatus = true

			// 更新收藏数
			if err := tx.Model(&models.Discussion{}).Where("id = ?", discussionID).
				UpdateColumn("stars", gorm.Expr("stars + ?", 1)).Error; err != nil {
				return err
			}
		} else {
			// 取消收藏
			if err := tx.Delete(&star).Error; err != nil {
				return err
			}

			// 更新收藏数
			if err := tx.Model(&models.Discussion{}).Where("id = ?", discussionID).
				UpdateColumn("stars", gorm.Expr("stars - ?", 1)).Error; err != nil {
				return err
			}
			currentStatus = false
		}
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "操作成功",
		"isStarred": currentStatus,
	})
}

// 获取用户互动状态
func GetUserInteractions(c *gin.Context) {
	discussionID := c.Param("id")
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	var isLiked bool
	var isStarred bool

	var like models.DiscussionLike
	if err := config.DB.Where("user_id = ? AND discussion_id = ?", userID, discussionID).First(&like).Error; err == nil {
		isLiked = true
	}

	var star models.DiscussionStar
	if err := config.DB.Where("user_id = ? AND discussion_id = ?", userID, discussionID).First(&star).Error; err == nil {
		isStarred = true
	}

	c.JSON(http.StatusOK, gin.H{
		"isLiked":   isLiked,
		"isStarred": isStarred,
	})
}
