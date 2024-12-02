package controllers

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"goj/pkg/config"
	"goj/pkg/models"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ImportProblems 导入题目
func ImportProblems(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "获取上传文件失败",
		})
		return
	}

	// 检查文件类型
	if !strings.HasSuffix(file.Filename, ".zip") {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "只支持上传 ZIP 格式文件",
		})
		return
	}

	// 创建临时目录用于解压文件
	tmpDir, err := os.MkdirTemp("", "problem_import_*")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "创建临时目录失败",
		})
		return
	}
	defer os.RemoveAll(tmpDir)

	// 保存上传的文件
	uploadPath := filepath.Join(tmpDir, "upload.zip")
	if err := c.SaveUploadedFile(file, uploadPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "保存上传文件失败",
		})
		return
	}

	// 开启事务
	tx := config.DB.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "开启事务失败",
		})
		return
	}

	// 递归处理 ZIP 文件
	var importedCount int
	var processZipFile func(string) error

	processZipFile = func(zipPath string) error {
		// 创建临时目录解压当前 ZIP
		tmpDir, err := os.MkdirTemp("", "problem_tmp_*")
		if err != nil {
			return fmt.Errorf("创建临时目录失败: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		// 解压当前 ZIP
		if err := unzip(zipPath, tmpDir); err != nil {
			return fmt.Errorf("解压文件失败: %v", err)
		}

		// 遍历解压后的目录
		return filepath.Walk(tmpDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// 如果是 ZIP 文件，递归处理
			if !info.IsDir() && strings.HasSuffix(info.Name(), ".zip") {
				// 检查是否是题目 ZIP（通过查找 problem.json）
				tmpUnzipDir, err := os.MkdirTemp("", "problem_check_*")
				if err != nil {
					return err
				}
				defer os.RemoveAll(tmpUnzipDir)

				if err := unzip(path, tmpUnzipDir); err != nil {
					return err
				}

				// 检查是否包含 problem.json
				if _, err := os.Stat(filepath.Join(tmpUnzipDir, "problem.json")); err == nil {
					// 这是一个题目 ZIP，处理题目导入
					if err := importProblem(tx, path, &importedCount); err != nil {
						return err
					}
				} else {
					// 这可能是一个包含多个题目的 ZIP，递归处理
					if err := processZipFile(path); err != nil {
						return err
					}
				}
			}
			return nil
		})
	}

	// 处理上传的文件
	if err := processZipFile(uploadPath); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": err.Error(),
		})
		return
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "提交事务失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": fmt.Sprintf("成功导入 %d 个题目", importedCount),
	})
}

// importProblem 导入单个题目
func importProblem(tx *gorm.DB, zipPath string, importedCount *int) error {
	// 创建临时目录
	problemTmpDir, err := os.MkdirTemp("", "problem_import_*")
	if err != nil {
		return fmt.Errorf("创建题目临时目录失败: %v", err)
	}
	defer os.RemoveAll(problemTmpDir)

	// 解压题目文件
	if err := unzip(zipPath, problemTmpDir); err != nil {
		return fmt.Errorf("解压题目文件失败: %v", err)
	}

	// 读取题目信息
	problemJsonPath := filepath.Join(problemTmpDir, "problem.json")
	jsonData, err := os.ReadFile(problemJsonPath)
	if err != nil {
		return fmt.Errorf("读取题目信息失败: %v", err)
	}

	var problemInfo struct {
		ID          string   `json:"id"`
		Title       string   `json:"title"`
		Content     string   `json:"content"`
		Tags        []string `json:"tags"`
		Languages   []string `json:"languages"`
		Source      string   `json:"source"`
		Role        string   `json:"role"`
		Difficulty  int      `json:"difficulty"`
		TimeLimit   int      `json:"timeLimit"`
		MemoryLimit int      `json:"memoryLimit"`
		UseSPJ      bool     `json:"useSPJ"`
	}

	if err := json.Unmarshal(jsonData, &problemInfo); err != nil {
		return fmt.Errorf("解析题目信息失败: %v", err)
	}

	// 生成新的题目ID
	var seq struct {
		ID uint
	}

	// 创建序列表（如果不存在）
	tx.Exec(`
		CREATE TABLE IF NOT EXISTS problem_seq (
			id INT NOT NULL AUTO_INCREMENT PRIMARY KEY
		) AUTO_INCREMENT = 10001
	`)

	// 插入新记录获取ID
	if err := tx.Exec("INSERT INTO problem_seq VALUES (NULL)").Error; err != nil {
		return fmt.Errorf("生成题目ID失败: %v", err)
	}

	// 获取生成的ID
	if err := tx.Raw("SELECT LAST_INSERT_ID() as id").Scan(&seq).Error; err != nil {
		return fmt.Errorf("获取题目ID失败: %v", err)
	}

	newProblemID := fmt.Sprintf("%05d", seq.ID)

	// 创建题目记录
	problem := models.Problem{
		ID:          newProblemID,
		Title:       problemInfo.Title,
		Difficulty:  problemInfo.Difficulty,
		Role:        problemInfo.Role,
		Tag:         strings.Join(problemInfo.Tags, ","),
		Source:      problemInfo.Source,
		Languages:   strings.Join(problemInfo.Languages, ","),
		TimeLimit:   int64(problemInfo.TimeLimit),
		MemoryLimit: int64(problemInfo.MemoryLimit),
		UseSPJ:      problemInfo.UseSPJ,
	}

	// 保存到数据库
	if err := tx.Create(&problem).Error; err != nil {
		return fmt.Errorf("保存题目失败: %v", err)
	}

	// 更新 problem.json 中的 ID
	problemInfo.ID = newProblemID
	updatedJsonData, err := json.MarshalIndent(problemInfo, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化更新后的题目信息失败: %v", err)
	}

	// 写入更新后的 problem.json
	if err := os.WriteFile(problemJsonPath, updatedJsonData, 0644); err != nil {
		return fmt.Errorf("保存更新后的题目信息失败: %v", err)
	}

	// 复制题目文件到目标目录
	problemDir := filepath.Join("data", "problems", newProblemID)
	if err := copyDir(problemTmpDir, problemDir); err != nil {
		return fmt.Errorf("复制题目文件失败: %v", err)
	}

	*importedCount++
	return nil
}

// 辅助函数：解压 ZIP 文件
func unzip(src, dst string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		// 安全检查：防止 ZIP 滑动攻击
		if strings.Contains(f.Name, "..") {
			continue
		}

		fpath := filepath.Join(dst, f.Name)

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}
	return nil
}

// ExportProblem 导出单个题目
func ExportProblem(c *gin.Context) {
	problemID := c.Param("id")

	// 创建临时目录
	tmpDir, err := os.MkdirTemp("", "problem_export_*")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "创建临时目录失败",
		})
		return
	}
	defer os.RemoveAll(tmpDir)

	// 复制题目文件到临时目录
	problemDir := filepath.Join("data", "problems", problemID)
	if err := copyDir(problemDir, tmpDir); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "复制题目文件失败",
		})
		return
	}

	// 创建 ZIP 文件
	zipFile := filepath.Join(tmpDir, "problem.zip")
	if err := createZip(tmpDir, zipFile); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "创建 ZIP 文件失败",
		})
		return
	}

	// 发送文件
	c.FileAttachment(zipFile, fmt.Sprintf("problem_%s.zip", problemID))
}

// ExportBatchProblems 批量导出题目
func ExportBatchProblems(c *gin.Context) {
	var req struct {
		ProblemIDs []string `json:"problemIds" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
		})
		return
	}

	// 创建临时目录
	tmpDir, err := os.MkdirTemp("", "problems_export_*")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "创建临时目录失败",
		})
		return
	}
	defer os.RemoveAll(tmpDir)

	// 为每个题目创建 ZIP
	for _, problemID := range req.ProblemIDs {
		problemDir := filepath.Join("data", "problems", problemID)
		problemTmpDir := filepath.Join(tmpDir, problemID)

		// 复制题目文件
		if err := copyDir(problemDir, problemTmpDir); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "复制题目文件失败",
			})
			return
		}

		// 创建题目 ZIP
		zipFile := filepath.Join(tmpDir, fmt.Sprintf("problem_%s.zip", problemID))
		if err := createZip(problemTmpDir, zipFile); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "创建题目 ZIP 失败",
			})
			return
		}

		// 清理临时目录
		os.RemoveAll(problemTmpDir)
	}

	// 创建最终的 ZIP 文件
	finalZip := filepath.Join(tmpDir, "problems.zip")
	if err := createZip(tmpDir, finalZip); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "创建最终 ZIP 失败",
		})
		return
	}

	// 发送文件
	c.FileAttachment(finalZip, fmt.Sprintf("problems_%d.zip", time.Now().Unix()))
}

// ExportAllProblems 导出所有题目
func ExportAllProblems(c *gin.Context) {
	// 获取所有题目 ID
	var problems []models.Problem
	if err := config.DB.Select("id").Find(&problems).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取题目列表失败",
		})
		return
	}

	// 创建临时目录
	tmpDir, err := os.MkdirTemp("", "all_problems_export_*")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "创建临时目录失败",
		})
		return
	}
	defer os.RemoveAll(tmpDir)

	// 为每个题目创建 ZIP
	for _, problem := range problems {
		problemDir := filepath.Join("data", "problems", problem.ID)
		problemTmpDir := filepath.Join(tmpDir, problem.ID)

		// 复制题目文件
		if err := copyDir(problemDir, problemTmpDir); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "复制题目文件失败",
			})
			return
		}

		// 创建题目 ZIP
		zipFile := filepath.Join(tmpDir, fmt.Sprintf("problem_%s.zip", problem.ID))
		if err := createZip(problemTmpDir, zipFile); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "创建题目 ZIP 失败",
			})
			return
		}

		// 清理临时目录
		os.RemoveAll(problemTmpDir)
	}

	// 创建最终的 ZIP 文件
	finalZip := filepath.Join(tmpDir, "all_problems.zip")
	if err := createZip(tmpDir, finalZip); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "创建最终 ZIP 失败",
		})
		return
	}

	// 发送文件
	c.FileAttachment(finalZip, fmt.Sprintf("all_problems_%d.zip", time.Now().Unix()))
}

// 辅助函数：复制目录
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 计算目标路径
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		targetPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(targetPath, info.Mode())
		}

		// 复制文件
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		return os.WriteFile(targetPath, data, info.Mode())
	})
}

// 辅助函数：创建 ZIP 文件
func createZip(src, dst string) error {
	zipfile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer zipfile.Close()

	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 不要将目标 ZIP 文件包含在内
		if path == dst {
			return nil
		}

		// 获取相对路径
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		// 创建 ZIP 头信息
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = relPath

		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(writer, file)
		return err
	})
}
