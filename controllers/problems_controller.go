package controllers

import (
	"net/http"

	"github.com/Hedgeho9X/TeachU/models"
	"github.com/Hedgeho9X/TeachU/services"
	"github.com/gin-gonic/gin"
)

func CreatProblem(c *gin.Context) {
	var input struct {
		ExamId          uint    `json:"exam_id" binding:"required"`
		Keypoint        string  `json:"keypoint" binding:"required"`
		QuestionsNumber uint    `json:"question_number" binding:"required"`
		TotalScore      float64 `json:"total_score" binding:"required"`
		Content         string  `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":  0,
			"error": "参数错误:" + err.Error(),
		})
		return
	}
	problems := &models.Problems{
		ExamId:          input.ExamId,
		Keypoint:        input.Keypoint,
		QuestionsNumber: input.QuestionsNumber,
		TotalScore:      input.TotalScore,
		Content:         input.Content,
	}
	if err := services.CreatProblems(problems); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":  0,
			"error": "数据库保存失败:" + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    1,
		"message": "上传成功",
		"data": gin.H{
			"id":              problems.Id,
			"ExamId":          problems.ExamId,
			"Keypoint":        problems.Keypoint,
			"QuestionsNumber": problems.QuestionsNumber,
			"TotalScore":      problems.TotalScore,
			"Content":         problems.Content,
		},
	})
}
