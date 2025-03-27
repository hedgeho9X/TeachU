package controllers

import (
	"net/http"

	"github.com/Hedgeho9X/TeachU/models"
	"github.com/Hedgeho9X/TeachU/services"
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
