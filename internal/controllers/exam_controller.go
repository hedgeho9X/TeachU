package controllers

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"log"

	"code.sajari.com/docconv/v2" // 新增docconv导入
	"github.com/Hedgeho9X/TeachU/internal/models"
	"github.com/Hedgeho9X/TeachU/internal/services"
	"github.com/gin-gonic/gin"
)

func CreateExam(c *gin.Context) {
	var input struct {
		CreatedUserID uint                  `form:"created_user_id"` // 修改字段名
		Name          string                `form:"name" binding:"required"`
		Subject       string                `form:"subject" binding:"required"`
		File          *multipart.FileHeader `form:"file" binding:"required"`
		ClassId       uint                  `form:"class_id" binding:"required"`
	}
	userIDInterface, _ := c.Get("userID")
	userID, ok := userIDInterface.(uint)
	if !ok {
		fmt.Printf("类型转换失败: userIDInterface=%v, type=%T\n", userIDInterface, userIDInterface)
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"msg":  "无效的用户ID",
		})
		return
	}
	input.CreatedUserID = userID
	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":  0,
			"error": "参数错误: " + err.Error(),
		})
		return
	}
	// 验证文件大小和类型
	if input.File.Size > 20<<20 {
		c.JSON(http.StatusOK, gin.H{"code": 0, "error": "文件大小超过限制"})
		return
	}

	ext := filepath.Ext(input.File.Filename)
	if ext != ".docx" && ext != ".doc" && ext != ".jpg" && ext != ".png" {
		c.JSON(http.StatusOK, gin.H{"code": 0, "error": "仅支持.doc/.docx文件或jpg/png图片格式"})
		return
	}

	var keyPoint string
	var err error

	// 根据文件类型处理内容
	switch ext {
	case ".docx", ".doc":
		// 使用docconv解析Word文档
		file, err := input.File.Open()
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"code":  0,
				"error": fmt.Sprintf("文件打开失败: %v", err),
			})
			return
		}
		defer file.Close()

		// 根据文件类型设置MIME类型
		mimeType := "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
		if ext == ".doc" {
			mimeType = "application/msword"
		}

		// 添加更详细的日志
		log.Printf("[INFO] 开始解析文档，文件类型: %s, MIME类型: %s", ext, mimeType)

		// 尝试使用docconv转换
		response, err := docconv.Convert(file, mimeType, true)
		if err != nil {
			log.Printf("[ERROR] 文档解析详细错误: %v", err)

			// 重新打开文件尝试读取文件内容进行调试
			file.Seek(0, 0) // 重置文件指针到开头
			fileBytes, readErr := io.ReadAll(file)
			if readErr == nil {
				log.Printf("[DEBUG] 文件大小: %d 字节, 前100字节: %v", len(fileBytes), fileBytes[:min(100, len(fileBytes))])
			}

			c.JSON(http.StatusOK, gin.H{
				"code":  0,
				"error": fmt.Sprintf("文档解析失败: %v，请确保文件格式正确且未损坏", err),
			})
			return
		}

		var content = response.Body
		log.Printf("[INFO] 文档解析成功，内容长度: %d 字符", len(content))

		// 如果内容为空，返回错误
		if len(content) == 0 {
			c.JSON(http.StatusOK, gin.H{
				"code":  0,
				"error": "文档内容为空，无法解析",
			})
			return
		}

		keyPoint, err = services.Chat(content, services.DoubaoLite, services.AnalyzePrompt)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"code":  0,
				"error": fmt.Sprintf("AI分析失败: %v", err),
			})
			return
		}

	case ".jpg", ".png":
		pic, _ := input.File.Open()
		FileBytes, err := io.ReadAll(pic)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"code": 0, "error": "文件打开失败: " + err.Error()})
			return
		}
		baseContent, err := services.ImageToBase64(FileBytes)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"code": 0, "error": "base64转换失败: " + err.Error()})
			return
		}
		keyPoint, err = services.AIAnalyzePic(baseContent)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"code": 0, "error": "内容解析失败: " + err.Error()})
			return
		}
	}
	// 增加JSON有效性检查
	// 增加转义字符处理（新增代码）
	keyPoint = strings.ReplaceAll(keyPoint, `\\(`, "(") // 处理转义左括号
	keyPoint = strings.ReplaceAll(keyPoint, `\\)`, ")") // 处理转义右括号
	keyPoint = strings.ReplaceAll(keyPoint, `\\`, `\`)  // 修正反斜杠转义

	result, err := services.ParseQuestions(keyPoint)
	if err != nil {
		log.Printf("[ERROR] JSON解析失败: %v\n处理后的内容: %s", err, keyPoint) // 修改日志字段名
		c.JSON(http.StatusOK, gin.H{"code": 0, "error": "试题格式解析失败: " + err.Error()})
		return
	}

	// 修改原有的fmt.Print为带日志级别的记录
	log.Printf("\n\n[INFO] 试卷解析结果: %s", keyPoint)
	log.Printf("\n\n[INFO] Json解析结果: %+v", result) // 使用%+v输出结构体详情
	// 创建并保存试题
	exam := models.Exam{
		UserId:   input.CreatedUserID, // 修改字段名
		ExamName: input.Name,
		Subject:  input.Subject,
		ClassId:  input.ClassId,
	}

	if err := services.CreateExam(&exam); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 0, "error": "数据库保存失败: " + err.Error()})
		return
	}

	// 统一返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"code":    1,
		"message": "上传成功",
		"data": gin.H{
			"created_user_id": exam.UserId, // 修改字段名
			"id":              exam.ID,
			"name":            exam.ExamName,
			"subject":         exam.Subject,
			"KeyPoint":        result,
		},
	})
}

func DeleteExam(c *gin.Context) {
	examId := c.Param("id")
	if examId == "" {
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"msg":  "试题ID不能为空",
		})
		return
	}
	// 获取当前用户ID（用于权限验证）
	userIDInterface, _ := c.Get("userID")
	userID, ok := userIDInterface.(uint)
	if !ok {
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"msg":  "无效的用户ID",
		})
		return
	}

	// 调用service层删除试题
	if err := services.DeleteExam(examId, userID); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"msg":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 1,
		"msg":  "删除试题成功",
	})
}
func ListExam(c *gin.Context) {
	// 获取当前用户ID（用于权限验证）
	classID := c.Query("class_id")
	if classID == "" {
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"msg":  "班级ID不能为空",
		})
		return
	}
	userIDInterface, _ := c.Get("userID")
	userID, ok := userIDInterface.(uint)
	if !ok {
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"msg":  "无效的用户ID",
		})
		return
	}
	// 调用service层获取试题列表
	// 将字符串类型的classID转换为uint类型
	classIDUint, err := strconv.ParseUint(classID, 10, 32)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"msg":  "无效的班级ID格式",
		})
		return
	}
	exams, err := services.ListExamByClassID(uint(classIDUint), userID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"msg":  err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 1,
		"msg":  "获取试题列表成功",
		"data": exams,
	})

}
