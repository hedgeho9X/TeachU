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
	// 获取班级所有学生成绩
	var scores []models.Score
	if err := tx.Where("exam_id = ?", examID).Find(&scores).Error; err != nil {
		tx.Rollback()
		return analysis, err
	}
	var examTotalScore float64
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
	for _, total := range studentScores { // 修改这里：应该遍历studentScores而不是studentTotals
		studentTotals = append(studentTotals, total)
	}

	// 计算基础统计指标
	totalStudents := len(studentTotals)
	var total, max, min float64

	// 防止空数组导致NaN
	if totalStudents == 0 {
		tx.Rollback()
		return analysis, errors.New("没有找到学生成绩数据")
	}

	for i, score := range studentTotals {
		// fmt.Printf("学生 %d 的总分: %f\n", i+1, score)  // 移除或注释掉调试输出
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

	// 计算得分率（修改为平均得分率）
	for i := range keypointMetrics {
		if total, exists := totalMap[keypointMetrics[i].Keypoint]; exists && total > 0 {
			averageScore := keypointMetrics[i].TotalScore / float64(keypointMetrics[i].StudentCount)
			keypointMetrics[i].ScoreRate = averageScore / total
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
		Where("scores.student_id = ?", studentID).
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
			kpScores[i].TotalScore = total
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

// 进步退步同学
func GetStuRank(examID uint) (models.StuRankResp, error) {
	tx := config.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 获取当前考试所有排名数据
	var currentRanks []models.Rank
	if err := tx.Where("exam_id = ?", examID).
		Order("`rank` ASC").
		Find(&currentRanks).Error; err != nil {
		tx.Rollback()
		return models.StuRankResp{}, fmt.Errorf("获取排名失败: %v", err)
	}

	// 获取所有学生姓名
	studentNameMap := make(map[uint]string)
	var studentIDs []uint
	for _, rank := range currentRanks {
		studentIDs = append(studentIDs, rank.StudentID)
	}

	var students []models.Student
	if err := tx.Where("id IN ?", studentIDs).Find(&students).Error; err != nil {
		tx.Rollback()
		return models.StuRankResp{}, fmt.Errorf("获取学生信息失败: %v", err)
	}

	for _, student := range students {
		studentNameMap[student.ID] = student.StudentName
	}

	// 构建响应数据
	resp := models.StuRankResp{
		ExamID: examID,
	}

	// 转换全部排名数据
	for _, rank := range currentRanks {
		resp.StuRank = append(resp.StuRank, models.Rank{
			ExamID:      rank.ExamID,
			StudentID:   rank.StudentID,
			StudentName: studentNameMap[rank.StudentID],
			Rank:        rank.Rank,
			TotalScore:  rank.TotalScore,
		})
	}

	// 如果rank表中没有记录，则根据score表生成rank
	if len(currentRanks) == 0 {
		// 获取所有学生总分
		var studentTotals []struct {
			StudentID uint
			Total     float64
		}
		if err := tx.Model(&models.Score{}).
			Select("student_id, SUM(score) as total").
			Where("exam_id = ?", examID).
			Group("student_id").
			Scan(&studentTotals).Error; err != nil {
			tx.Rollback()
			return models.StuRankResp{}, fmt.Errorf("获取学生总分失败: %v", err)
		}

		// 按总分排序
		sort.Slice(studentTotals, func(i, j int) bool {
			return studentTotals[i].Total > studentTotals[j].Total
		})

		// 生成排名记录
		for i, st := range studentTotals {
			// 获取学生姓名
			var student models.Student
			if err := tx.First(&student, st.StudentID).Error; err != nil {
				tx.Rollback()
				return models.StuRankResp{}, fmt.Errorf("获取学生信息失败: %v", err)
			}

			currentRanks = append(currentRanks, models.Rank{
				ExamID:      examID,
				StudentID:   st.StudentID,
				StudentName: student.StudentName, // 直接使用查询结果
				Rank:        i + 1,
				TotalScore:  st.Total,
			})

			// 更新姓名映射表
			studentNameMap[st.StudentID] = student.StudentName // 新增
		}

		// 保存生成的排名前确保student_name存在
		if err := tx.CreateInBatches(currentRanks, 100).Error; err != nil {
			tx.Rollback()
			return models.StuRankResp{}, fmt.Errorf("保存排名失败: %v", err)
		}
	}

	// 获取前一次考试ID
	var prevExamID uint
	tx.Model(&models.Exam{}).
		Select("id").
		Where("class_id = (SELECT class_id FROM exams WHERE id = ?) AND created_at < (SELECT created_at FROM exams WHERE id = ?)", examID, examID).
		Order("created_at DESC").
		Limit(1).
		Scan(&prevExamID)

	// 获取历史排名数据
	prevRankMap := make(map[uint]int)
	if prevExamID != 0 {
		var prevRanks []models.Rank
		if err := tx.Where("exam_id = ?", prevExamID).Find(&prevRanks).Error; err == nil {
			for _, r := range prevRanks {
				prevRankMap[r.StudentID] = r.Rank
			}
		}
	}

	// 提取进步和退步数据
	for _, rank := range currentRanks {
		if prevRank, exists := prevRankMap[rank.StudentID]; exists {
			rankChange := prevRank - rank.Rank
			change := models.RankChange{
				StudentName: studentNameMap[rank.StudentID],
				Rank:        rank.Rank,
				Change:      uint(math.Abs(float64(rankChange))),
			}

			if rankChange > 0 {
				resp.ImproveRank = append(resp.ImproveRank, change)
			} else if rankChange < 0 {
				resp.DeclineRank = append(resp.DeclineRank, change)
			}
		}
	}

	// 按进步/退步名次变化排序（从大到小）
	sort.Slice(resp.ImproveRank, func(i, j int) bool {
		return resp.ImproveRank[i].Change > resp.ImproveRank[j].Change
	})
	sort.Slice(resp.DeclineRank, func(i, j int) bool {
		return resp.DeclineRank[i].Change > resp.DeclineRank[j].Change
	})

	// 限制最多10条记录
	if len(resp.ImproveRank) > 10 {
		resp.ImproveRank = resp.ImproveRank[:10]
	}
	if len(resp.DeclineRank) > 10 {
		resp.DeclineRank = resp.DeclineRank[:10]
	}

	tx.Commit()
	return resp, nil
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
	// 防止除以零或NaN结果
	if len(scores) > 0 {
		return math.Sqrt(variance / float64(len(scores)))
	}
	return 0
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
