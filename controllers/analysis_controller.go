package controllers

import (
	"net/http"
	"strconv"

	"github.com/Hedgeho9X/TeachU/services"
	"github.com/gin-gonic/gin"
)

func AnalyzeExam(c *gin.Context) {
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
	analysis, err := services.AnalyzeClass(uint(ExamIdUint))
	KeypointAnalysis, err := services.AnalyzeKeypoint(uint(ExamIdUint))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"class_analysis": analysis, "keypoint_analysis": KeypointAnalysis})
}

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
	}
	c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "success", "analysis_content": AnalysisContent})
}

func AnalyzeStudent(c *gin.Context) {
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
	analysis, err := services.AnalyzeStudent(uint(ExamIdUint), uint(StudentIDUint))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"student": analysis})
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

func RagRecommend(c *gin.Context) {
	var input struct {
		ExamID    uint   `json:"exam_id"`
		StudentID uint   `json:"student_id"`
		Prompt    string `json:"prompt"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusOK, gin.H{"error": err.Error()})
		return
	}
	recommend, err := services.RagRecommend(input.ExamID, input.StudentID, input.Prompt)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 0, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "success", "recommend": recommend})
}
