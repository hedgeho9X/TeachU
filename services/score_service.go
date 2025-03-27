package services

import (
	"fmt"

	"github.com/Hedgeho9X/TeachU/config"
	"github.com/Hedgeho9X/TeachU/models"
)

func CreateScores(examID uint, scores []models.ScoreInput) (int, error) {
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
		// 遍历试题并创建成绩记录
		for _, q := range s.Questions {
			// 验证试题是否存在
			var question models.Problems
			if err := tx.Where("questions_number = ? AND exam_id = ?", q.QuestionNumber, examID).First(&question, q.QuestionNumber).Error; err != nil {
				tx.Rollback()
				return 0, fmt.Errorf("本次考试中试题号%d不存在", q.QuestionNumber)
			}
			// // 验证试题是否属于当前考试

			// 创建成绩记录
			score := models.Score{
				StudentID:      student.ID,
				ExamID:         examID,
				QuestionNumber: q.QuestionNumber,
				Score:          q.Score,
			}

			if err := tx.Create(&score).Error; err != nil {
				tx.Rollback()
				return 0, fmt.Errorf("保存成绩失败: %v", err)
			}
		}
		successCount++
	}
	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return 0, fmt.Errorf("事务提交失败: %v", err)
	}
	return successCount, nil
}

func ListScores(examID uint) ([]models.StudentScoreResponse, error) {
	// 创建map暂存学生成绩
	scoreMap := make(map[uint]*models.StudentScoreResponse)

	// 查询所有成绩记录
	var scores []models.Score
	err := config.DB.
		Preload("Student").
		Where("exam_id = ?", examID).
		Order("student_id, question_number").
		Find(&scores).Error
	if err != nil {
		return nil, err
	}
	// 遍历成绩记录构建响应结构
	for _, s := range scores {
		if _, exists := scoreMap[s.StudentID]; !exists {
			scoreMap[s.StudentID] = &models.StudentScoreResponse{
				StudentInfo: models.StudentSimple{
					ID:            s.Student.ID,
					StudentNumber: s.Student.StudentNumber,
					StudentName:   s.Student.StudentName,
				},
				Questions: []models.QuestionScore{},
				Total:     0,
			}
		}

		// 添加题目成绩
		scoreMap[s.StudentID].Questions = append(scoreMap[s.StudentID].Questions, models.QuestionScore{
			QuestionNumber: s.QuestionNumber,
			Score:          s.Score,
		})

		// 累加总分
		scoreMap[s.StudentID].Total += s.Score
	}

	// 转换map为slice
	var response []models.StudentScoreResponse
	for _, v := range scoreMap {
		response = append(response, *v)
	}

	return response, nil
}
