package controllers

import (
	"fmt"
	"net/http"

	"github.com/Hedgeho9X/TeachU/services"
	"github.com/gin-gonic/gin"
)

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
	classes, username, err := services.GetClassesByUserID(userID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"msg":  "获取班级列表失败",
		})
		return
	}

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

// // UpdateClass 更新班级信息
// func UpdateClass(c *gin.Context) {
// 	// TODO: 实现更新班级信息逻辑
// }

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
