package controllers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/Hedgeho9X/TeachU/services"
	"github.com/gin-gonic/gin"
)

func UploadScore(c *gin.Context) {
	// userIDInterface, _ := c.Get("userID")
	// userID, ok := userIDInterface.(uint)
	// if !ok {
	// 	c.JSON(http.StatusOK, gin.H{"code": "0", "error": "无效的userID"})
	// 	return
	// }

	var input struct {
		ExamID uint `json:"exam_id" binding:"required"`
		Scores []struct {
			StudentNumber  string  `json:"student_number" binding:"required"`
			StudentName    string  `json:"student_name"`
			QuestionNumber int     `json:"question_id" binding:"required"`
			Score          float64 `json:"score" binding:"required,gte=0"`
		} `json:"scores" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": "0", "error": "输入有误: " + err.Error()})
		return
	}

	// 处理成绩数据（这里需要调用service层保存数据）
	// 示例处理逻辑：
	var successCount int
	// 将输入的成绩数据转换为服务层期望的格式
	scores := make([]struct {
		StudentNumber  string  `json:"student_number"`
		QuestionNumber int     `json:"question_number"`
		Score          float64 `json:"score"`
	}, len(input.Scores))

	for i, s := range input.Scores {
		scores[i] = struct {
			StudentNumber  string  `json:"student_number"`
			QuestionNumber int     `json:"question_number"`
			Score          float64 `json:"score"`
		}{
			StudentNumber:  s.StudentNumber,
			QuestionNumber: s.QuestionNumber,
			Score:          s.Score,
		}
	}

	successCount, err := services.CreateScores(input.ExamID, scores)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": "0", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 1,
		"msg":  fmt.Sprintf("成功保存%d条成绩数据", successCount),
		"data": gin.H{
			"exam_id": input.ExamID,
			"total":   len(input.Scores),
			"success": successCount,
		},
	})
}

func ListScore(c *gin.Context) {
	examID := c.Query("exam_id")
	if examID == "" {
		c.JSON(http.StatusOK, gin.H{
			"code": "0",
			"msg":  "exam_id不能为空",
		})
		return

		ExamIDUint, err := strconv.ParseUint(examID, 10, 64)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"code": "0",
				"msg":  "exam_id必须为数字",
			})
			return
		}
	}
	scores, err := services.ListScores(uint(ExamIDUint))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": "0",
			"msg":  err.Error(),
		})
		return
		c.JSON(http.StatusOK, gin.H{
			"code": "1",
			"msg":  "获取成绩成功",
			"data": scores,
		})
	}

}
