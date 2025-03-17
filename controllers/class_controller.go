package controllers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/Hedgeho9X/TeachU/models"
	"github.com/Hedgeho9X/TeachU/services"
	"github.com/gin-gonic/gin"
)

// ListClasses 获取班级列表
func ListClasses(c *gin.Context) {
	userIDInterface, _ := c.Get("userID")
	userID, ok := userIDInterface.(uint)
	if !ok {
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"msg":  "无效的用户ID",
		})
		return
	}
	// 调用service层获取班级列表
	classes, err := services.GetClassesByUserID(userID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"msg":  "获取班级列表失败",
		})
		return
	}

	// 从上下文获取用户名
	username, _ := c.Get("username")

	c.JSON(http.StatusOK, gin.H{
		"code":     1,
		"msg":      "获取班级列表成功",
		"data":     classes,
		"username": username,
	})
}

// CreateClass 创建班级
func CreateClass(c *gin.Context) {
	// 获取并转换用户ID
	userIDInterface, _ := c.Get("userID")
	userID, ok := userIDInterface.(uint) // 修改为 uint 类型
	if !ok {
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"msg":  "无效的用户ID",
		})
		return
	}

	// 绑定输入参数
	var input struct {
		ClassNumber int `json:"class_number" binding:"required"`
		GradeLevel  int `json:"grade_level" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"msg":  fmt.Sprintf("参数错误：%v", err),
		})
		return
	}

	// 调用service层创建班级时直接使用 userID
	if err := services.CreateClass(input.ClassNumber, input.GradeLevel, userID); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"msg":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 1,
		"msg":  "创建班级成功",
	})
}

// DeleteClass 删除班级
func DeleteClass(c *gin.Context) {
	// 获取班级ID
	classID := c.Param("id")
	if classID == "" {
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"msg":  "班级ID不能为空",
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

	// 调用service层删除班级
	if err := services.DeleteClass(classID, userID); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"msg":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 1,
		"msg":  "删除班级成功",
	})
}

func ListStudents(c *gin.Context) {
	// 获取班级ID
	classID := c.Query("class_id")
	if classID == "" {
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"msg":  "班级ID不能为空",
		})
		return
	}
	//操作services.GetStudentsByClassID获取学生列表
	classIDInt, err := strconv.Atoi(classID)
	if err != nil || classIDInt < 0 {
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"msg":  "无效的班级ID",
		})
		return
	}
	students, err := services.GetStudentsByClassID(uint(classIDInt))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"msg":  "获取学生列表失败",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 1,
		"msg":  "获取学生列表成功",
		"data": students,
	})
}

// 空接口接收前端文件
func UploadFiles(c *gin.Context) {
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		c.String(http.StatusOK, fmt.Sprintf("上传文件失败: %v", err))
		return
	}
	defer file.Close()
	c.String(http.StatusOK, "文件上传并解析成功")
}

// 批量创建学生（json）
func ImportStudents(c *gin.Context) {
	// 定义请求体结构
	type StudentInput struct {
		StudentNumber int    `json:"student_number" binding:"required"`
		StudentName   string `json:"student_name" binding:"required"`
	}

	type ImportRequest struct {
		ClassID  uint           `json:"class_id" binding:"required"` // 统一传递 ClassID
		Students []StudentInput `json:"students" binding:"required"` // 学生列表（无需重复 ClassID）
	}

	var req ImportRequest

	// 绑定 JSON 数据
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":  0,
			"error": "请求参数错误: " + err.Error(),
		})
		return
	}

	// 遍历处理学生数据
	// for _, student := range req.Students {
	// 	// 实际业务逻辑示例：
	// 	// 1. 创建学生记录，使用统一的 req.ClassID
	// 	fmt.Printf(
	// 		"创建学生: 学号=%d, 姓名=%s, 班级ID=%d\n",
	// 		student.StudentNumber,
	// 		student.StudentName,
	// 		req.ClassID,
	// 	)
	// }
	// 遍历处理学生数据
	for _, student := range req.Students {
		// 调用service层创建学生
		err := services.CreateStudent(req.ClassID, models.Student{
			StudentNumber: strconv.Itoa(student.StudentNumber),
			StudentName:   student.StudentName,
		})
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"code": 0,
				"msg":  fmt.Sprintf("创建学生失败: %v", err),
			})
			return
		}
	}

	// 返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"code":    1,
		"message": fmt.Sprintf("成功为班级 %d 添加 %d 名学生", req.ClassID, len(req.Students)),
	})
}
