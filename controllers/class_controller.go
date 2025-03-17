package controllers

import (
	"fmt"
	"net/http"
	"strconv"

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
