package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"

	"github.com/Hedgeho9X/TeachU/config"
	"github.com/Hedgeho9X/TeachU/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

func AnalyzeClass(examID uint) (models.ClassMetric, error) {
	// 从数据库中获取考试信息
	tx := config.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	var analysis models.ClassMetric
	if err := tx.Where("exam_id = ?", examID).First(&analysis).Error; err == nil {
		tx.Commit() // 找到记录后立即提交事务
		return analysis, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		tx.Rollback()
		return analysis, err // 返回实际错误
	}
	//不存在该记录，进行分析
	// 获取班级所有学生成绩
	var scores []models.Score
	if err := tx.Where("exam_id = ?", examID).Find(&scores).Error; err != nil {
		tx.Rollback()
		return analysis, err
	}
	var examTotalScore int64
	if err := tx.Model(&models.Problems{}).
		Select("SUM(total_score)").
		Where("exam_id = ?", examID).
		Row().
		Scan(&examTotalScore); err != nil {
		tx.Rollback()
		return analysis, err
	}

	// 获取班级所有学生总分
	var studentTotals []float64
	studentScores := make(map[uint]float64) // 学生ID到总分的映射

	for _, s := range scores {
		studentScores[s.StudentID] += s.Score
	}

	// 转换总分到切片
	for _, total := range studentTotals {
		studentTotals = append(studentTotals, total)
	}

	// 计算基础统计指标
	totalStudents := len(studentTotals)
	var total, max, min float64

	for i, score := range studentTotals {
		if i == 0 {
			max = score
			min = score
		}
		total += score
		if score > max {
			max = score
		}
		if score < min {
			min = score
		}
	}

	// 调用统计函数计算复杂指标
	avg := total / float64(totalStudents)
	stdDev := calculateStdDev(studentTotals, avg)
	median := calculateMedian(studentTotals)
	scoreBuckets := calculateScoreBuckets(studentTotals)

	// 构建指标对象
	analysis = models.ClassMetric{
		ExamID:           examID,
		StudentCount:     totalStudents,
		TotalScore:       examTotalScore,
		AvgTotalScore:    avg,
		MaxTotalScore:    max,
		MinTotalScore:    min,
		MedianTotalScore: median,
		StdDev:           stdDev,
		ScoreBuckets:     scoreBuckets,
	}

	// 持久化分析结果
	if err := tx.Create(&analysis).Error; err != nil {
		tx.Rollback()
		return analysis, err
	}

	tx.Commit()
	return analysis, nil
}

func AnalyzeKeypoint(examID uint) (models.KeypointMetricResp, error) {
	tx := config.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 获取知识点统计
	var keypointMetrics []models.KeypointMetric
	err := tx.Model(&models.Score{}).
		Select(`
			p.keypoint,
			SUM(scores.score) as total_score,
			AVG(scores.score) as average_score,
			COUNT(DISTINCT scores.student_id) as student_count
		`).
		Joins("JOIN problems p ON p.exam_id = scores.exam_id AND p.questions_number = scores.question_number").
		Where("scores.exam_id = ?", examID).
		Group("p.keypoint").
		Scan(&keypointMetrics).Error

	if err != nil {
		tx.Rollback()
		return models.KeypointMetricResp{}, err
	}

	// 获取知识点总分（用于计算得分率）
	var kpTotals []struct {
		Keypoint string
		Total    float64
	}
	tx.Model(&models.Problems{}).
		Select("keypoint, SUM(total_score) as total").
		Where("exam_id = ?", examID).
		Group("keypoint").
		Scan(&kpTotals)

	// 转换为map方便查找
	totalMap := make(map[string]float64)
	for _, t := range kpTotals {
		totalMap[t.Keypoint] = t.Total
		// print(t.Keypoint, t.Total, "\n")
	}

	// 计算得分率
	for i := range keypointMetrics {
		if total, exists := totalMap[keypointMetrics[i].Keypoint]; exists && total > 0 {
			keypointMetrics[i].ScoreRate = keypointMetrics[i].TotalScore / total
			keypointMetrics[i].TotalScore = total
		}
	}

	// 构建响应结构
	resp := models.KeypointMetricResp{
		ExamID:          examID,
		KeypointMetrics: keypointMetrics,
	}

	tx.Commit()
	return resp, nil
}

func AiAnalyzeClass(examID uint) (string, error) {
	// 调用豆包进行分析
	var KeypointAnalysis models.KeypointMetricResp
	var ClassMetric models.ClassMetric
	KeypointAnalysis, err := AnalyzeKeypoint(examID)
	if err != nil {
		return "", err
	}
	ClassMetric, err = AnalyzeClass(examID)
	if err != nil {
		return "", err
	}
	JsonRes := struct {
		KeypointAnalysis models.KeypointMetricResp `json:"keypoint_analysis"`
		ClassMetric      models.ClassMetric        `json:"class_metric"`
	}{
		KeypointAnalysis: KeypointAnalysis,
		ClassMetric:      ClassMetric,
	}

	// 转换为JSON字符串
	jsonData, err := json.Marshal(JsonRes)
	if err != nil {
		return "", fmt.Errorf("JSON序列化失败: %v", err)
	}

	result, err := Chat(string(jsonData), DoubaoLite, ScoreAnalyzePrompt)
	if err != nil {
		return "", err
	}
	return result, nil
}

func AnalyzeStudent(examID uint, studentID uint) (models.StudentAnalysisResponse, error) {
	resp := models.StudentAnalysisResponse{}
	resp.StudentID = studentID
	tx := config.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	var scores []models.Score
	if err := tx.Where("exam_id = ? AND student_id = ?", examID, studentID).Find(&scores).Error; err != nil {
		tx.Rollback()
		return models.StudentAnalysisResponse{}, err
	}

	if err := tx.Model(&models.Score{}).
		Select("SUM(score)").
		Where("exam_id = ? AND student_id = ?", examID, studentID).
		Row().
		Scan(&resp.TotalScore); err != nil {
		tx.Rollback()
		return models.StudentAnalysisResponse{}, err
	}

	tx.Model(&models.Score{}).
		Select(`
            scores.exam_id,
            exams.exam_name,
            SUM(score) as score,
            exams.created_at,
            (SELECT AVG(score) FROM scores s WHERE s.exam_id = scores.exam_id) as average_score
        `).
		Joins("JOIN exams ON exams.id = scores.exam_id").
		Where("scores.student_id = ? AND scores.exam_id = ?", studentID, examID).
		Group("scores.exam_id, exams.exam_name, exams.created_at").
		Order("exams.created_at DESC").
		Limit(6).
		Scan(&resp.StudentHistory)
	// 知识点得分统计
	var kpScores []models.StudentKeypoints
	err := tx.Model(&models.Score{}).
		Select(`
            p.keypoint,
            SUM(scores.score) as score,
            AVG(scores.score) as average_score
        `).
		Joins("JOIN problems p ON p.exam_id = scores.exam_id AND p.questions_number = scores.question_number").
		Where("scores.exam_id = ? AND scores.student_id = ?", examID, studentID).
		Group("p.keypoint").
		Scan(&kpScores).Error
	if err != nil {
		tx.Rollback()
		return models.StudentAnalysisResponse{}, err
	}

	// 获取知识点总分（用于计算得分率）
	var kpTotals []struct {
		Keypoint string
		Total    float64
	}
	tx.Model(&models.Problems{}).
		Select("keypoint, SUM(total_score) as total").
		Where("exam_id = ?", examID).
		Group("keypoint").
		Scan(&kpTotals)

	totalMap := make(map[string]float64)
	for _, t := range kpTotals {
		totalMap[t.Keypoint] = t.Total
	}

	// 计算得分率和掌握程度
	for i := range kpScores {
		if total, exists := totalMap[kpScores[i].Keypoint]; exists && total > 0 {
			kpScores[i].ScoreRate = kpScores[i].Score / total
			kpScores[i].MasteryLevel = getMasteryLevel(kpScores[i].ScoreRate)
			kpScores[i].IsHigh = kpScores[i].Score >= kpScores[i].AverageScore
		}
	}

	resp.StudentMetrics = kpScores

	tx.Commit()
	return resp, nil
}

func AiAnalyzeStudent(examID uint, studentID uint) (string, error) {

	analysis, err := AnalyzeStudent(examID, studentID)
	if err != nil {
		return "", err
	}

	JsonRes := struct {
		AnalysisResult models.StudentAnalysisResponse `json:"analysis"`
	}{
		AnalysisResult: analysis,
	}

	// 转换为JSON字符串
	jsonData, err := json.Marshal(JsonRes)
	if err != nil {
		return "", fmt.Errorf("JSON序列化失败: %v", err)
	}

	result, err := Chat(string(jsonData), DoubaoLite, StudentAnalyzePrompt)
	if err != nil {
		return "", err
	}
	return result, nil
}

// 新增辅助函数
func getMasteryLevel(rate float64) string {
	switch {
	case rate >= 0.8:
		return "high"
	case rate >= 0.6:
		return "medium"
	default:
		return "low"
	}
}

// 计算标准差
func calculateStdDev(scores []float64, mean float64) float64 {
	if len(scores) == 0 {
		return 0
	}
	variance := 0.0
	for _, score := range scores {
		diff := score - mean
		variance += diff * diff
	}
	return math.Sqrt(variance / float64(len(scores)))
}

// 计算中位数
func calculateMedian(scores []float64) float64 {
	if len(scores) == 0 {
		return 0
	}
	sorted := make([]float64, len(scores))
	copy(sorted, scores)
	sort.Float64s(sorted)

	mid := len(sorted) / 2
	if len(sorted)%2 == 0 {
		return (sorted[mid-1] + sorted[mid]) / 2
	}
	return sorted[mid]
}

// 计算分数段分布
func calculateScoreBuckets(scores []float64) datatypes.JSON {
	buckets := make(map[int]int) // 改用整数作为key便于排序

	// 统计原始数据
	for _, score := range scores {
		lower := int(math.Floor(score/10)) * 10
		buckets[lower]++
	}

	// 获取有序的key
	keys := make([]int, 0, len(buckets))
	for k := range buckets {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	// 构建有序map
	ordered := make(map[string]int)
	for _, k := range keys {
		key := fmt.Sprintf("%d-%d", k, k+9)
		ordered[key] = buckets[k]
	}

	jsonData, _ := json.Marshal(ordered)
	return datatypes.JSON(jsonData)
}
