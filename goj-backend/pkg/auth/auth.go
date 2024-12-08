package auth

import (
	"fmt"
	"goj/pkg/config"
	"goj/pkg/models"
	"goj/pkg/types"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// POST /api/auth/login
// 请求:
type LoginRequest struct {
	Account  string `json:"account" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
}

// 响应:
type LoginResponse struct {
	Token string     `json:"token"`
	User  types.User `json:"user"`
}

// POST /api/auth/register
// 请求:
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=2,max=12"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// GET /api/user/profile
// 响应:
type User struct {
	ID        uint      `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Avatar    string    `json:"avatar"`
	Bio       string    `json:"bio"`
	Role      string    `json:"role"`
	Score     int       `json:"score"`
	Accepted  int       `json:"accepted"`
	Submitted int       `json:"submitted"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// PUT /api/user/profile
// 请求:
type UpdateProfileRequest struct {
	Email string `json:"email" binding:"omitempty,email"`
	Bio   string `json:"bio"`
}

// 修改密码请求
type UpdatePasswordRequest struct {
	OldPassword string `json:"oldPassword" binding:"required,min=6"`
	NewPassword string `json:"newPassword" binding:"required,min=6"`
}

// 通用响应格式
type Response struct {
	Code    int         `json:"code"`    // 状态码
	Message string      `json:"message"` // 提示信息
	Data    interface{} `json:"data"`    // 数据
}

const (
	MaxAvatarSize = 256 * 1024 // 256KB
)

// Claims JWT claims 结构体
type Claims struct {
	UserID uint   `json:"userID"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// 在 middleware 包中使用的验证函数
func ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return config.JWTSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

// 处理头像上传
func UploadAvatar(c *gin.Context) {
	log.Printf("Received avatar upload request. Content-Type: %s", c.GetHeader("Content-Type"))

	userID, exists := c.Get("userID")
	if !exists {
		log.Printf("User ID not found in context")
		c.JSON(http.StatusUnauthorized, Response{
			Code:    401,
			Message: "未授权",
			Data:    nil,
		})
		return
	}

	// 获取上传的文件
	file, err := c.FormFile("avatar")
	if err != nil {
		log.Printf("Failed to get form file: %v", err)
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "请选择要上传的头像",
			Data:    nil,
		})
		return
	}

	log.Printf("Received file: name=%s, size=%d, type=%s",
		file.Filename, file.Size, file.Header.Get("Content-Type"))

	// 检查文件大小
	if file.Size > MaxAvatarSize {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "头像文件大小不能超过256KB",
			Data:    nil,
		})
		return
	}

	// 检查文件类型
	if !strings.HasPrefix(file.Header.Get("Content-Type"), "image/") {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "只能上传图片文件",
			Data:    nil,
		})
		return
	}

	// 使用固定的文件名格式：userID.jpg
	filename := fmt.Sprintf("%d.jpg", userID.(uint))
	filepath := fmt.Sprintf("public/images/avatars/%s", filename)

	// 添加调试日志
	log.Printf("Saving avatar to: %s", filepath)

	// 确保目录存在
	avatarDir := "public/images/avatars"
	if err := os.MkdirAll(avatarDir, 0755); err != nil {
		log.Printf("Failed to create directory: %v", err)
		c.JSON(http.StatusInternalServerError, Response{
			Code:    500,
			Message: "创建目录失败",
			Data:    nil,
		})
		return
	}

	// 检查目录权限
	if info, err := os.Stat(avatarDir); err != nil {
		log.Printf("Failed to stat directory: %v", err)
	} else {
		log.Printf("Directory permissions: %v", info.Mode())
	}

	// 保存文件
	if err := c.SaveUploadedFile(file, filepath); err != nil {
		log.Printf("Failed to save file: %v", err)
		c.JSON(http.StatusInternalServerError, Response{
			Code:    500,
			Message: "保存头像失败",
			Data:    nil,
		})
		return
	}

	// 验证文件是否成功保存
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		log.Printf("File was not saved successfully: %v", err)
	} else {
		log.Printf("File was saved successfully at: %s", filepath)
	}

	// 更新用户头像信息
	var user models.User
	if result := config.DB.First(&user, userID); result.Error != nil {
		log.Printf("User not found: %v", result.Error)
		c.JSON(http.StatusNotFound, Response{
			Code:    404,
			Message: "用户不存在",
			Data:    nil,
		})
		return
	}

	// 更新数据库中的头像路径 - 使用相对路径
	avatarURL := fmt.Sprintf("/images/avatars/%s", filename)
	log.Printf("Updating avatar URL to: %s", avatarURL)

	if result := config.DB.Model(&user).Update("avatar", avatarURL); result.Error != nil {
		log.Printf("Failed to update avatar in database: %v", result.Error)
		c.JSON(http.StatusInternalServerError, Response{
			Code:    500,
			Message: "更新头像信息失败",
			Data:    nil,
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "头像上传成功",
		Data: gin.H{
			"avatar": avatarURL,
		},
	})
}

// 1. 实现用户注册接口
func Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "参数错误：用户名长度必须在2-12个字符之间，密码不少于6个字符",
			Data:    nil,
		})
		return
	}

	// 检查邮箱是否已存在
	var existingUser models.User
	if result := config.DB.Where("email = ?", req.Email).First(&existingUser); result.Error == nil {
		c.JSON(http.StatusBadRequest, Response{
			400,
			"邮箱已被注册",
			nil,
		})
		return
	}

	// 检查是否是第一个用户
	var userCount int64
	if err := config.DB.Model(&models.User{}).Count(&userCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Code:    500,
			Message: "服务器错误",
			Data:    nil,
		})
		return
	}

	// 密码加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Code:    500,
			Message: "服务器错误",
			Data:    nil,
		})
		return
	}

	// 创建用户时设置默认头像
	defaultAvatarPath := "/images/avatars/default-avatar.png"
	// 确保默认头像文件存在
	if _, err := os.Stat("public" + defaultAvatarPath); os.IsNotExist(err) {
		os.MkdirAll("public/images/avatars", 0755)
		defaultSrc := "assets/default-avatar.png"
		defaultDst := "public" + defaultAvatarPath
		if err := copyFile(defaultSrc, defaultDst); err != nil {
			log.Printf("Failed to copy default avatar: %v", err)
		}
	}

	// 创建用户，如果是第一个用户则设置为管理员
	user := models.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Role:         "user",
		Avatar:       defaultAvatarPath,
	}

	if userCount == 0 {
		user.Role = "admin"
	}

	if result := config.DB.Create(&user); result.Error != nil {
		log.Printf("Failed to create user: %v", result.Error)
		c.JSON(http.StatusInternalServerError, Response{
			Code:    500,
			Message: "创建用户失败",
			Data:    nil,
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "注册成功",
		Data:    nil,
	})
}

// 添加复制文件的辅助函数
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

// 2. 实现用户登录接口
func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "参数错误",
			Data:    nil,
		})
		return
	}

	var user models.User
	// 先尝试用邮箱查找
	result := config.DB.Where("email = ?", req.Account).First(&user)
	if result.Error != nil {
		// 如果邮箱找不到，尝试用用户名查找
		result = config.DB.Where("username = ?", req.Account).First(&user)
		if result.Error != nil {
			c.JSON(http.StatusUnauthorized, Response{
				Code:    401,
				Message: "用户名/邮箱或密码错误",
				Data:    nil,
			})
			return
		}
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, Response{
			Code:    401,
			Message: "用户名/邮箱或密码错误",
			Data:    nil,
		})
		return
	}

	// 生成 JWT token
	claims := &Claims{
		UserID: user.ID,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(config.JWTSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Code:    500,
			Message: "生成token失败",
			Data:    nil,
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "登录成功",
		Data: LoginResponse{
			Token: tokenString,
			User:  *user.ToDTO(),
		},
	})
}

// 3. 实现获取用户信息接口
func GetProfile(c *gin.Context) {
	// 打印调试信息
	log.Printf("GetProfile called - Context: %+v", c.Keys)

	userID, exists := c.Get("userID")
	if !exists {
		log.Printf("userID not found in context")
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "未授权",
			"data":    nil,
		})
		return
	}

	var user models.User
	if result := config.DB.First(&user, userID); result.Error != nil {
		log.Printf("Failed to find user: %v", result.Error)
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "用户不存在",
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    user.ToDTO(),
	})
}

// 4. 实现更新用户信息接口
func UpdateProfile(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "未授权",
		})
		return
	}

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误",
		})
		return
	}

	// 更新用户信息
	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "用户不存在",
		})
		return
	}

	// 只更新 bio 字段
	if err := config.DB.Model(&user).Updates(models.User{
		Bio: req.Bio,
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "更新失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "更新成功",
		"data":    user.ToDTO(),
	})
}

// UpdatePassword 处理修改密码的请求
func UpdatePassword(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, Response{
			Code:    401,
			Message: "未授权",
			Data:    nil,
		})
		return
	}

	var req UpdatePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "参数错误",
			Data:    nil,
		})
		return
	}

	var user models.User
	if result := config.DB.First(&user, userID); result.Error != nil {
		c.JSON(http.StatusNotFound, Response{
			Code:    403,
			Message: "用户不存在",
			Data:    nil,
		})
		return
	}

	// 验证旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.OldPassword)); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "旧密码错误",
			Data:    nil,
		})
		return
	}

	// 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Code:    500,
			Message: "服务器错误",
			Data:    nil,
		})
		return
	}

	// 更新用户密码
	user.PasswordHash = string(hashedPassword)
	if result := config.DB.Save(&user); result.Error != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Code:    500,
			Message: "更新密码失败",
			Data:    nil,
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "密码更新成功",
		Data:    nil,
	})
}

// 获取指定用户的公开资料
func GetUserProfile(c *gin.Context) {
	username := c.Param("username")

	var user models.User
	if username != "" {
		// 如果提供了用户名，根据用户名查找
		if result := config.DB.Where("username = ?", username).First(&user); result.Error != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "用户不存在",
				"data":    nil,
			})
			return
		}
	} else {
		// 如果没有提供用户名，获取当前登录用户的资料
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "未授权",
				"data":    nil,
			})
			return
		}
		if result := config.DB.First(&user, userID); result.Error != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "用户不存在",
				"data":    nil,
			})
			return
		}
	}

	// 转换为 DTO，只返回公开信息
	publicProfile := types.User{
		ID:               user.ID,
		Username:         user.Username,
		Avatar:           user.Avatar,
		Bio:              user.Bio,
		Role:             user.Role,
		Submissions:      user.Submissions,
		AcceptedProblems: user.AcceptedProblems,
		Rating:           user.Rating,
		CreatedAt:        user.CreatedAt,
		UpdatedAt:        user.UpdatedAt,
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    publicProfile,
	})
}

// 获取用户提交记录
func GetUserSubmissions(c *gin.Context) {
	username := c.Param("username")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	offset := (page - 1) * pageSize

	var user models.User
	if err := config.DB.Where("username = ?", username).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, Response{
			Code:    404,
			Message: "用户不存在",
			Data:    nil,
		})
		return
	}

	var submissions []models.Submission
	var total int64

	// 获取总数
	config.DB.Model(&models.Submission{}).Where("user_id = ?", user.ID).Count(&total)

	// 获取分页数据
	err := config.DB.Where("user_id = ?", user.ID).
		Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&submissions).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Code:    500,
			Message: "获取提交记录失败",
			Data:    nil,
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "获取成功",
		Data: gin.H{
			"submissions": submissions,
			"total":       total,
		},
	})
}

// 获取用户已解决题目
func GetUserSolvedProblems(c *gin.Context) {
	username := c.Param("username")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	offset := (page - 1) * pageSize

	var user models.User
	if err := config.DB.Where("username = ?", username).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, Response{
			Code:    404,
			Message: "用户不存在",
			Data:    nil,
		})
		return
	}

	var problems []models.Problem
	var total int64

	// 获取已解决的题目（通过 submissions 表关联）
	subQuery := config.DB.Table("submissions").
		Select("DISTINCT problem_id").
		Where("user_id = ? AND status = 'Accepted'", user.ID)

	// 获取总数
	config.DB.Model(&models.Problem{}).
		Where("id IN (?)", subQuery).
		Count(&total)

	// 获取分页数据
	err := config.DB.Where("id IN (?)", subQuery).
		Order("id ASC").
		Limit(pageSize).
		Offset(offset).
		Find(&problems).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Code:    500,
			Message: "获取题目列表失败",
			Data:    nil,
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "获取成功",
		Data: gin.H{
			"problems": problems,
			"total":    total,
		},
	})
}

// 获取用户比赛记录
func GetUserContests(c *gin.Context) {
	username := c.Param("username")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	offset := (page - 1) * pageSize

	var user models.User
	if err := config.DB.Where("username = ?", username).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, Response{
			Code:    404,
			Message: "用户不存在",
			Data:    nil,
		})
		return
	}

	var contestRecords []struct {
		models.Contest
		Rank              int   `json:"rank"`
		Score             int   `json:"score"`
		TotalParticipants int64 `json:"totalParticipants"`
	}
	var total int64

	// 获取总数
	config.DB.Table("contest_participants").
		Where("user_id = ?", user.ID).
		Count(&total)

	// 获取分页数据
	err := config.DB.Table("contests").
		Select("contests.*, cp.rank, cp.score, (SELECT COUNT(*) FROM contest_participants WHERE contest_id = contests.id) as total_participants").
		Joins("JOIN contest_participants cp ON cp.contest_id = contests.id").
		Where("cp.user_id = ?", user.ID).
		Order("contests.start_time DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&contestRecords).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Code:    500,
			Message: "获取比赛记录失败",
			Data:    nil,
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "获取成功",
		Data: gin.H{
			"contests": contestRecords,
			"total":    total,
		},
	})
}
