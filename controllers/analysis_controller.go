package controllers

import (
	"net/http"
	"strconv"

	"github.com/Hedgeho9X/TeachU/services"
	"github.com/gin-gonic/gin"
)

func AnalyzeExam(c *gin.Context) {
	examIDstr := c.Query("exam_id")
	if examIDstr == "" {
		c.JSON(http.StatusOK, gin.H{"error": "exam_id为空"})
		return
	}
	// 调用services层分析考试
	examIdUint, err := strconv.ParseUint(examIDstr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": "0",
			"msg":  "exam_id必须为数字",
		})
		return
	}
	analysis, err := services.AnalyzeClass(uint(examIdUint))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"analysis": analysis})
}
