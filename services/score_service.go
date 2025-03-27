package services

import (
	"fmt"

	"github.com/Hedgeho9X/TeachU/config"
	"github.com/Hedgeho9X/TeachU/models"
)

func CreateScores(examID uint, scores []struct {
	StudentNumber  string  `json:"student_number"`
	QuestionNumber int     `json:"question_number"`
	Score          float64 `json:"score"`
}) (int, error) {
	// 开启事务
	tx := config.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var successCount int
	for _, s := range scores {
		// 根据学号获取学生ID
		var student models.Student
		if err := tx.Where("student_number = ?", s.StudentNumber).First(&student).Error; err != nil {
			tx.Rollback()
			return 0, fmt.Errorf("学号%s不存在", s.StudentNumber)
		}

		// 验证考试是否存在
		var exam models.Exam
		if err := tx.First(&exam, examID).Error; err != nil {
			tx.Rollback()
			return 0, fmt.Errorf("考试ID%d不存在", examID)
		}

		// 创建成绩记录
		score := models.Score{
			StudentID:      student.ID,
			ExamID:         examID,
			QuestionNumber: s.QuestionNumber,
			Score:          s.Score,
		}

		if err := tx.Create(&score).Error; err != nil {
			tx.Rollback()
			return 0, fmt.Errorf("保存成绩失败: %v", err)
		}
		successCount++
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return 0, fmt.Errorf("事务提交失败: %v", err)
	}

	return successCount, nil
}

func ListScores(examID uint) ([]models.Score, error) {
	var scores []models.Score

	// 使用预加载获取成绩记录及关联的学生信息
	if err := config.DB.Preload("Student").Where("exam_id = ?", examID).Find(&scores).Error; err != nil {
		return nil, fmt.Errorf("获取成绩记录失败: %v", err)
	}

	return scores, nil
}
