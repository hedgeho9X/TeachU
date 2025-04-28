package controllers

import (
	"net/http"
	"strconv"

	"github.com/Hedgeho9X/TeachU/internal/services"
	"github.com/gin-gonic/gin"
)

func AiAnalyzeClass(c *gin.Context) {
	ExamIdStr := c.Query("exam_id")
	if ExamIdStr == "" {
		c.JSON(http.StatusOK, gin.H{"error": "exam_id为空"})
		return
	}
	// 调用services层分析考试
	ExamIdUint, err := strconv.ParseUint(ExamIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": "0",
			"msg":  "exam_id必须为数字",
		})
		return
	}
	// ExamCla
	AnalysisContent, err := services.AiAnalyzeClass(uint(ExamIdUint))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 0, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "success", "analysis_content": AnalysisContent})
}

func AiAnalyzeStudent(c *gin.Context) {
	ExamIdStr := c.Query("exam_id")
	StudentIDstr := c.Query("student_id")
	if ExamIdStr == "" || StudentIDstr == "" {
		c.JSON(http.StatusOK, gin.H{"error": "exam_id或student_id为空"})
		return
	}
	// 调用services层分析考试
	ExamIdUint, _ := strconv.ParseUint(ExamIdStr, 10, 64)
	StudentIDUint, err := strconv.ParseUint(StudentIDstr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": "0",
			"msg":  "exam_id必须为数字",
		})
		return
	}
	AnalysisContent, err := services.AiAnalyzeStudent(uint(ExamIdUint), uint(StudentIDUint))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 0, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "success", "analysis": AnalysisContent})
}
