package controllers

import (
	"fmt"
	"net/http"

	"github.com/Hedgeho9X/TeachU/internal/config"

	"github.com/Hedgeho9X/TeachU/internal/models"
	"github.com/Hedgeho9X/TeachU/internal/services"
	"github.com/gin-gonic/gin"
)

type ProblemDetail struct {
	Keypoint        string  `json:"keypoint" binding:"required"`
	QuestionsNumber uint    `json:"questions_number" binding:"required"`
	TotalScore      float64 `json:"total_score" binding:"required"`
	Content         string  `json:"content" binding:"required"`
}

func CreatProblem(c *gin.Context) {
	var input struct {
		ExamId   uint            `json:"exam_id" binding:"required"`
		Problems []ProblemDetail `json:"problems" binding:"required,dive"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":  0,
			"error": "参数错误:" + err.Error(),
		})
		return
	}
	var problems []models.Problems
	for _, detail := range input.Problems {
		problems = append(problems, models.Problems{
			ExamId:          input.ExamId,
			Keypoint:        detail.Keypoint,
			QuestionsNumber: detail.QuestionsNumber,
			TotalScore:      detail.TotalScore,
			Content:         detail.Content,
		})
		if detail.Keypoint == "" || detail.QuestionsNumber == 0 || detail.TotalScore == 0 || detail.Content == "" {
			c.JSON(http.StatusOK, gin.H{
				"code": 0,
				"msg":  "传入内容含空信息",
			})
			return
		}
		var existingProblem models.Problems
		if err := config.DB.Where("exam_id = ? AND questions_number = ?",
			input.ExamId, detail.QuestionsNumber).First(&existingProblem).Error; err == nil {
			c.JSON(http.StatusOK, gin.H{
				"code": 0,
				"msg":  fmt.Sprintf("试卷 %d 中已存在题号为 %d 的试题", input.ExamId, detail.QuestionsNumber),
			})
			return
		}
	}
	if err := services.CreatProblems(problems); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":  0,
			"error": "数据库保存失败:" + err.Error(),
		})
		return
	}
	var responseProblems []models.Problems
	for _, p := range problems {
		responseProblems = append(responseProblems, models.Problems{
			Keypoint:        p.Keypoint,
			QuestionsNumber: p.QuestionsNumber,
			TotalScore:      p.TotalScore,
			Content:         p.Content,
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    1,
		"message": "上传成功",
		"data": gin.H{
			"id":       input.ExamId,
			"problems": responseProblems,
		},
	})
}
