package services

import (
	"fmt"
	"sort"

	"github.com/Hedgeho9X/TeachU/internal/config"
	"github.com/Hedgeho9X/TeachU/internal/models"
)

func CreateScores(examID uint, scoreInputs []models.ScoreInput) (int, error) {
	tx := config.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var successCount int
	var scoresToSave []models.Score
	studentScoreMap := make(map[uint]float64)
	studentNameMap := make(map[uint]string) // 新增：存储学生ID与姓名的映射

	// 验证考试存在性
	var exam models.Exam
	if err := tx.First(&exam, examID).Error; err != nil {
		tx.Rollback()
		return 0, fmt.Errorf("考试ID%d不存在", examID)
	}

	// 新增：创建存在性检查的map
	existingRecords := make(map[string]struct{})

	// 查询当前考试已存在的记录
	var existingScores []models.Score
	if err := tx.Where("exam_id = ?", examID).Find(&existingScores).Error; err != nil {
		tx.Rollback()
		return 0, fmt.Errorf("查询已有记录失败: %v", err)
	}

	// 填充存在性检查map
	for _, es := range existingScores {
		key := fmt.Sprintf("%d-%d-%d", es.StudentID, es.ExamID, es.QuestionNumber)
		existingRecords[key] = struct{}{}
	}

	for _, s := range scoreInputs {
		// 学生验证
		var student models.Student
		if err := tx.Where("student_number = ?", s.StudentNumber).First(&student).Error; err != nil {
			tx.Rollback()
			return 0, fmt.Errorf("学号%s不存在", s.StudentNumber)
		}

		studentNameMap[student.ID] = student.StudentName // 新增：记录学生姓名

		// 遍历试题
		for _, q := range s.Questions {
			// 试题验证
			var problem models.Problems
			if err := tx.Where("exam_id = ? AND questions_number = ?", examID, q.QuestionNumber).
				First(&problem).Error; err != nil {
				tx.Rollback()
				return 0, fmt.Errorf("试题%d不存在", q.QuestionNumber)
			}

			// 新增：存在性检查
			recordKey := fmt.Sprintf("%d-%d-%d", student.ID, examID, q.QuestionNumber)
			if _, exists := existingRecords[recordKey]; exists {
				continue // 跳过已存在的记录
			}

			// 记录需要保存的成绩
			scoresToSave = append(scoresToSave, models.Score{
				StudentID:      student.ID,
				ExamID:         examID,
				QuestionNumber: q.QuestionNumber,
				Score:          q.Score,
			})

			// 将新增记录加入存在性检查map
			existingRecords[recordKey] = struct{}{}

			// 累加总分
			studentScoreMap[student.ID] += q.Score
		}
		successCount++
	}

	// 批量保存成绩
	if err := tx.CreateInBatches(scoresToSave, 100).Error; err != nil {
		tx.Rollback()
		return 0, fmt.Errorf("保存成绩失败: %v", err)
	}

	// 生成排名数据
	type studentTotal struct {
		ID    uint
		Total float64
	}
	var totals []studentTotal
	for k, v := range studentScoreMap {
		totals = append(totals, studentTotal{ID: k, Total: v})
	}

	// 按总分排序
	sort.Slice(totals, func(i, j int) bool {
		return totals[i].Total > totals[j].Total
	})

	// 生成排名记录
	var ranks []models.Rank
	for i, t := range totals {
		ranks = append(ranks, models.Rank{
			ExamID:      examID,
			StudentID:   t.ID,
			StudentName: studentNameMap[t.ID], // 新增：使用映射表获取姓名
			Rank:        i + 1,
			TotalScore:  t.Total,
		})
	}

	// 保存排名
	if err := tx.CreateInBatches(ranks, 100).Error; err != nil {
		tx.Rollback()
		return 0, fmt.Errorf("保存排名失败: %v", err)
	}

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
	sort.Slice(response, func(i, j int) bool {
		return response[i].Total > response[j].Total
	})
	return response, nil
}
