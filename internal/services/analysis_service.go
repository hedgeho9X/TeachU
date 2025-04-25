package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv" // 新增导入 strconv 包
	"time"

	"github.com/Hedgeho9X/TeachU/internal/config"
	"github.com/Hedgeho9X/TeachU/internal/models"
	"github.com/redis/go-redis/v9"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// ClassHistoryAnalysis 获取指定班级和科目的历史考试分析数据
// Subject: 科目名称
// ClassID: 班级 ID
// 返回包含历次考试分析结果的响应结构和可能发生的错误
func ClassHistoryAnalysis(Subject string, ClassID uint) (models.ClassHistoryAnalysisResponse, error) {
	var response models.ClassHistoryAnalysisResponse
	var exams []models.Exam

	// 1. 查询该班级该科目的所有考试记录，按时间倒序排列
	err := config.DB.Where("class_id = ? AND subject = ?", ClassID, Subject).Order("created_at DESC").Find(&exams).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 没有找到考试记录是正常情况，返回空列表
			return response, nil
		}
		// 其他数据库错误
		return response, fmt.Errorf("查询考试记录失败: %v", err)
	}

	// 2. 遍历每次考试，获取分析数据
	for _, exam := range exams {
		// 调用已有的分析函数获取单次考试的班级指标和知识点指标
		classMetric, err := AnalyzeClass(exam.ID)
		if err != nil {
			// 如果某次考试分析失败，可以记录日志并跳过，或者直接返回错误
			fmt.Printf("分析班级指标失败 (ExamID: %d): %v\n", exam.ID, err)
			continue // 跳过这次考试
			// 或者 return response, fmt.Errorf("分析班级指标失败 (ExamID: %d): %v", exam.ID, err)
		}

		keypointMetricResp, err := AnalyzeKeypoint(exam.ID)
		if err != nil {
			fmt.Printf("分析知识点指标失败 (ExamID: %d): %v\n", exam.ID, err)
			continue // 跳过这次考试
			// 或者 return response, fmt.Errorf("分析知识点指标失败 (ExamID: %d): %v", exam.ID, err)
		}

		// 3. 组装单次考试的历史记录
		historyEntry := models.ClassHistory{
			ExamID:             exam.ID,
			ExamName:           exam.ExamName,
			ExamTime:           exam.CreatedAt, // 使用考试创建时间作为考试时间
			ClassMetric:        classMetric,
			KeypointMetricResp: keypointMetricResp,
		}
		response.ClassHistory = append(response.ClassHistory, historyEntry)
	}

	return response, nil
}

// AnalyzeClass 分析指定考试的班级整体指标。
// 如果已有分析结果，则直接返回；否则，计算各项统计指标（总分、平均分、最高分、最低分、中位数、标准差、分数段分布）并存入数据库。
// examID: 需要分析的考试 ID。
// 返回计算出的班级指标和可能发生的错误。
func AnalyzeClass(examID uint) (models.ClassMetric, error) {
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

	var scores []models.Score
	if err := tx.Where("exam_id = ?", examID).Find(&scores).Error; err != nil {
		tx.Rollback()
		return analysis, err
	}

	var examTotalScore float64
	if err := tx.Model(&models.Problems{}).
		Select("COALESCE(SUM(total_score), 0) as total_score"). // 使用 COALESCE 处理 NULL
		Where("exam_id = ?", examID).
		Row().
		Scan(&examTotalScore); err != nil {
		tx.Rollback()
		return analysis, err
	}

	var studentTotals []float64
	studentScores := make(map[uint]float64) // 学生ID到总分的映射

	for _, s := range scores {
		studentScores[s.StudentID] += s.Score
	}

	for _, total := range studentScores {
		studentTotals = append(studentTotals, total)
	}

	totalStudents := len(studentTotals)
	var total, max, min float64

	if totalStudents == 0 {
		tx.Rollback()
		return analysis, errors.New("没有找到该班级成绩数据")
	}

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

	avg := total / float64(totalStudents)
	stdDev := calculateStdDev(studentTotals, avg)
	median := calculateMedian(studentTotals)
	scoreBuckets := calculateScoreBuckets(studentTotals)

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

	if err := tx.Create(&analysis).Error; err != nil {
		tx.Rollback()
		return analysis, err
	}

	tx.Commit()
	return analysis, nil
}

// AnalyzeKeypoint 分析指定考试的知识点掌握情况。
// 计算每个知识点的总分值、学生平均得分、参与学生数以及平均得分率（保留两位小数）。
// examID: 需要分析的考试 ID。
// 返回包含各知识点指标的响应结构和可能发生的错误。
func AnalyzeKeypoint(examID uint) (models.KeypointMetricResp, error) {
	tx := config.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 临时结构体用于接收数据库查询结果
	var queryResults []struct {
		Keypoint     string
		SumScore     float64 // SUM(scores.score) - 学生在该知识点上的总得分
		AvgScorePerQ float64 // AVG(scores.score) - 该知识点下每题的平均得分 (暂不直接使用)
		StudentCount int     // COUNT(DISTINCT scores.student_id) - 参与学生数
	}
	err := tx.Model(&models.Score{}).
		Select(`
			p.keypoint,
			SUM(scores.score) as sum_score,
			AVG(scores.score) as avg_score_per_q,
			COUNT(DISTINCT scores.student_id) as student_count
		`).
		Joins("JOIN problems p ON p.exam_id = scores.exam_id AND p.questions_number = scores.question_number").
		Where("scores.exam_id = ?", examID).
		Group("p.keypoint").
		Scan(&queryResults).Error

	if err != nil {
		tx.Rollback()
		return models.KeypointMetricResp{}, err
	}

	// 查询每个知识点的总分值
	var kpTotals []struct {
		Keypoint   string
		TotalValue float64 // SUM(total_score) - 知识点的总分值
	}
	tx.Model(&models.Problems{}).
		Select("keypoint, SUM(total_score) as total_value").
		Where("exam_id = ?", examID).
		Group("keypoint").
		Scan(&kpTotals)

	// 将知识点总分值存入 map 以便快速查找
	totalValueMap := make(map[string]float64)
	for _, t := range kpTotals {
		totalValueMap[t.Keypoint] = t.TotalValue
	}

	// 组装最终的 KeypointMetric 列表
	var keypointMetrics []models.KeypointMetric
	for _, qr := range queryResults {
		totalValue, exists := totalValueMap[qr.Keypoint]
		if !exists || totalValue <= 0 || qr.StudentCount <= 0 {
			// 如果知识点没有总分值或没有学生参与，则跳过
			continue
		}

		// 计算学生在该知识点上的平均得分
		averageScorePerStudent := qr.SumScore / float64(qr.StudentCount)

		// 计算平均得分率
		rawScoreRate := averageScorePerStudent / totalValue

		// 格式化得分率为两位小数
		formattedRateStr := fmt.Sprintf("%.2f", rawScoreRate)
		formattedRate, err := strconv.ParseFloat(formattedRateStr, 64)
		if err != nil {
			// 处理转换错误，例如记录日志或使用原始值
			fmt.Printf("警告: 转换得分率失败 (Keypoint: %s): %v, 使用原始值\n", qr.Keypoint, err)
			formattedRate = rawScoreRate // 出错时使用原始值
		}

		metric := models.KeypointMetric{
			Keypoint:     qr.Keypoint,
			TotalScore:   totalValue,             // 存储知识点的总分值
			AverageScore: averageScorePerStudent, // 存储学生在该知识点上的平均得分
			StudentCount: qr.StudentCount,        // 存储参与学生数
			ScoreRate:    formattedRate,          // 存储格式化后的平均得分率
		}
		keypointMetrics = append(keypointMetrics, metric)
	}

	resp := models.KeypointMetricResp{
		ExamID:          examID,
		KeypointMetrics: keypointMetrics,
	}

	tx.Commit()
	return resp, nil
}

// AiAnalyzeClass 调用 AI 模型（如豆包）对班级考试进行分析。
// 首先尝试从 Redis 缓存获取结果，如果缓存未命中，则调用 AnalyzeClass 和 AnalyzeKeypoint 获取数据，
// 将数据组合成 JSON 发送给 AI 模型，并将 AI 返回的分析结果存入 Redis 缓存。
// examID: 需要分析的考试 ID。
// 返回 AI 生成的分析文本和可能发生的错误。
func AiAnalyzeClass(examID uint) (string, error) {
	ctx := context.Background()
	redisKey := fmt.Sprintf("examID:%d", examID)
	AnalysisResult, err := config.RDB.Get(ctx, redisKey).Result()
	if err == nil { // 缓存命中
		return AnalysisResult, nil
	} else if err != redis.Nil { // Redis 其他错误
		return "", fmt.Errorf("redis获取缓存失败: %v", err)
	}

	// 缓存未命中，执行分析
	KeypointAnalysis, err := AnalyzeKeypoint(examID)
	if err != nil {
		return "", err
	}
	ClassMetric, err := AnalyzeClass(examID)
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

	jsonData, err := json.Marshal(JsonRes)
	if err != nil {
		return "", fmt.Errorf("JSON序列化失败: %v", err)
	}

	result, err := Chat(string(jsonData), DoubaoLite, ScoreAnalyzePrompt)
	if err != nil {
		return "", err
	}

	// 存入 Redis 缓存
	err = config.RDB.Set(ctx, redisKey, result, 5*time.Hour).Err()
	if err != nil {
		// 即使缓存失败，也应返回 AI 分析结果，但记录错误
		fmt.Printf("redis保存分析内容失败: %v\n", err)
		// return "", fmt.Errorf("redis保存分析内容失败: %v", err) // 根据需求决定是否因缓存失败而报错
	}
	return result, nil
}

// AnalyzeStudent 分析指定学生在某次考试中的表现。
// 计算学生的总分、历史成绩趋势（最近6次）、各知识点的得分、得分率和掌握程度。
// examID: 需要分析的考试 ID。
// studentID: 需要分析的学生 ID。
// 返回包含学生各项分析指标的响应结构和可能发生的错误。
func AnalyzeStudent(examID uint, studentID uint) (models.StudentAnalysisResponse, error) {
	resp := models.StudentAnalysisResponse{}
	resp.StudentID = studentID
	tx := config.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 检查学生是否存在该次考试成绩
	var scoreCount int64
	if err := tx.Model(&models.Score{}).Where("exam_id = ? AND student_id = ?", examID, studentID).Count(&scoreCount).Error; err != nil {
		tx.Rollback()
		return models.StudentAnalysisResponse{}, err
	}
	if scoreCount == 0 {
		tx.Rollback()
		return models.StudentAnalysisResponse{}, fmt.Errorf("未找到学生 %d 在考试 %d 中的成绩记录", studentID, examID)
	}

	if err := tx.Model(&models.Score{}).
		Select("SUM(score)").
		Where("exam_id = ? AND student_id = ?", examID, studentID).
		Row().
		Scan(&resp.TotalScore); err != nil {
		tx.Rollback()
		return models.StudentAnalysisResponse{}, err
	}

	// 查询学生历史成绩（最近6次）
	// 注意: ranks 表可能需要先生成才有数据，否则 average_score 可能为 NULL 或 0
	tx.Model(&models.Score{}).
		Select(`
            scores.exam_id,
            exams.exam_name,
            SUM(scores.score) as score,
            exams.created_at as time,
            (SELECT AVG(r.total_score) FROM ranks r WHERE r.exam_id = scores.exam_id) as average_score
        `).
		Joins("JOIN exams ON exams.id = scores.exam_id").
		Where("scores.student_id = ?", studentID).
		Group("scores.exam_id, exams.exam_name, exams.created_at").
		Order("exams.created_at DESC").
		Limit(6).
		Scan(&resp.StudentHistory)

	// 查询学生在本次考试中各知识点的得分情况以及班级在该知识点上的平均分
	var kpScores []struct {
		Keypoint     string
		Score        float64 // 学生在该知识点的得分
		AverageScore float64 // 班级在该知识点的平均分 (AVG(scores.score) per keypoint)
	}
	// 使用子查询计算班级知识点平均分
	err := tx.Model(&models.Score{}).
		Select(`
            p.keypoint,
            SUM(CASE WHEN scores.student_id = ? THEN scores.score ELSE 0 END) as score,
            AVG(scores.score) as average_score
        `, studentID). // 将 studentID 作为参数传入
		Joins("JOIN problems p ON p.exam_id = scores.exam_id AND p.questions_number = scores.question_number").
		Where("scores.exam_id = ?", examID).
		Group("p.keypoint").
		Scan(&kpScores).Error

	if err != nil {
		tx.Rollback()
		return models.StudentAnalysisResponse{}, fmt.Errorf("查询学生知识点得分失败: %v", err)
	}

	// 查询知识点总分
	var kpTotals []struct {
		Keypoint   string
		TotalValue float64 // 知识点总分值
	}
	tx.Model(&models.Problems{}).
		Select("keypoint, SUM(total_score) as total_value").
		Where("exam_id = ?", examID).
		Group("keypoint").
		Scan(&kpTotals)

	totalValueMap := make(map[string]float64)
	for _, t := range kpTotals {
		totalValueMap[t.Keypoint] = t.TotalValue
	}

	// 组装 StudentKeypoints
	resp.StudentMetrics = make([]models.StudentKeypoints, 0, len(kpScores))
	for _, kps := range kpScores {
		totalValue, exists := totalValueMap[kps.Keypoint]
		if !exists || totalValue <= 0 {
			continue // 跳过没有总分或总分为0的知识点
		}

		scoreRate := 0.0
		if totalValue > 0 {
			scoreRate = kps.Score / totalValue
		}

		// 格式化得分率为两位小数
		formattedRateStr := fmt.Sprintf("%.2f", scoreRate)
		formattedRate, _ := strconv.ParseFloat(formattedRateStr, 64) // 忽略错误，出错时默认为0.0

		studentKp := models.StudentKeypoints{
			Keypoint:     kps.Keypoint,
			Score:        kps.Score,        // 学生在该知识点的得分
			AverageScore: kps.AverageScore, // 班级在该知识点的平均分
			TotalScore:   totalValue,       // 知识点总分值
			ScoreRate:    formattedRate,    // 学生在该知识点的得分率
			MasteryLevel: getMasteryLevel(scoreRate),
			IsHigh:       kps.Score >= kps.AverageScore, // 与班级平均分比较
		}
		resp.StudentMetrics = append(resp.StudentMetrics, studentKp)
		// 移除之前的打印语句
		// fmt.Println(i, kpScores[i].Score, kpScores[i].AverageScore)
	}

	tx.Commit()
	return resp, nil
}

// AiAnalyzeStudent 调用 AI 模型（如豆包）对学生在某次考试中的表现进行分析。
// 首先尝试从 Redis 缓存获取结果，如果缓存未命中，则调用 AnalyzeStudent 获取数据，
// 将数据组合成 JSON 发送给 AI 模型，并将 AI 返回的分析结果存入 Redis 缓存。
// examID: 需要分析的考试 ID。
// studentID: 需要分析的学生 ID。
// 返回 AI 生成的分析文本和可能发生的错误。
func AiAnalyzeStudent(examID uint, studentID uint) (string, error) {
	ctx := context.Background()
	redisKey := fmt.Sprintf("examID:%d AND studentID:%d", examID, studentID)
	AnalysisResult, err := config.RDB.Get(ctx, redisKey).Result()
	if err == nil { // 缓存命中
		return AnalysisResult, nil
	} else if err != redis.Nil { // Redis 其他错误
		return "", fmt.Errorf("redis获取缓存失败: %v", err)
	}

	analysis, err := AnalyzeStudent(examID, studentID)
	if err != nil {
		return "", err
	}

	JsonRes := struct {
		AnalysisResult models.StudentAnalysisResponse `json:"analysis"`
	}{
		AnalysisResult: analysis,
	}

	jsonData, err := json.Marshal(JsonRes)
	if err != nil {
		return "", fmt.Errorf("JSON序列化失败: %v", err)
	}

	result, err := Chat(string(jsonData), DoubaoLite, StudentAnalyzePrompt)
	if err != nil {
		return "", err
	}

	// 存入 Redis 缓存
	err = config.RDB.Set(ctx, redisKey, result, 5*time.Hour).Err()
	if err != nil {
		// 即使缓存失败，也应返回 AI 分析结果，但记录错误
		fmt.Printf("redis保存分析内容失败: %v\n", err)
		// return "", fmt.Errorf("redis保存分析内容失败: %v", err) // 根据需求决定是否因缓存失败而报错
	}
	return result, nil
}

// getMasteryLevel 根据得分率判断知识点掌握程度。
// rate: 得分率 (0.0 - 1.0)。
// 返回掌握程度字符串 ("high", "medium", "low")。
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

// GetStuRank 获取指定考试的学生排名信息，包括全班排名、进步最大学生和退步最大学生。
// 如果 ranks 表中没有当前考试的排名数据，则会根据 scores 表计算并生成排名数据存入 ranks 表。
// examID: 需要获取排名的考试 ID。
// 返回包含排名信息的响应结构和可能发生的错误。
func GetStuRank(examID uint) (models.StuRankResp, error) {
	tx := config.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var currentRanks []models.Rank
	if err := tx.Where("exam_id = ?", examID).
		Order("`rank` ASC").
		Find(&currentRanks).Error; err != nil {
		tx.Rollback()
		return models.StuRankResp{}, fmt.Errorf("获取排名失败: %v", err)
	}

	studentNameMap := make(map[uint]string)
	var studentIDs []uint
	if len(currentRanks) > 0 {
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
	}

	resp := models.StuRankResp{
		ExamID: examID,
	}

	// 如果 rank 表中没有记录，则根据 score 表生成 rank
	if len(currentRanks) == 0 {
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

		if len(studentTotals) == 0 {
			tx.Commit() // 没有成绩数据，直接返回空结果
			return resp, nil
		}

		sort.Slice(studentTotals, func(i, j int) bool {
			return studentTotals[i].Total > studentTotals[j].Total
		})

		rankDataToSave := make([]models.Rank, 0, len(studentTotals))
		studentIDs = make([]uint, 0, len(studentTotals)) // 重置 studentIDs
		for _, st := range studentTotals {
			studentIDs = append(studentIDs, st.StudentID)
		}

		// 批量获取学生姓名
		var students []models.Student
		if err := tx.Where("id IN ?", studentIDs).Find(&students).Error; err != nil {
			tx.Rollback()
			return models.StuRankResp{}, fmt.Errorf("获取学生信息失败: %v", err)
		}
		tempStudentNameMap := make(map[uint]string)
		for _, student := range students {
			tempStudentNameMap[student.ID] = student.StudentName
			studentNameMap[student.ID] = student.StudentName // 更新全局 map
		}

		for i, st := range studentTotals {
			studentName, ok := tempStudentNameMap[st.StudentID]
			if !ok {
				// 理论上不应该发生，因为上面已经查询了所有 studentID
				tx.Rollback()
				return models.StuRankResp{}, fmt.Errorf("未能找到学生ID %d 的姓名", st.StudentID)
			}
			rank := models.Rank{
				ExamID:      examID,
				StudentID:   st.StudentID,
				StudentName: studentName,
				Rank:        i + 1,
				TotalScore:  st.Total,
			}
			rankDataToSave = append(rankDataToSave, rank)
			currentRanks = append(currentRanks, rank) // 同时填充 currentRanks 供后续使用
		}

		if err := tx.CreateInBatches(rankDataToSave, 100).Error; err != nil {
			tx.Rollback()
			return models.StuRankResp{}, fmt.Errorf("保存排名失败: %v", err)
		}
	}

	// 填充响应中的全班排名
	for _, rank := range currentRanks {
		resp.StuRank = append(resp.StuRank, models.Rank{
			ExamID:      rank.ExamID,
			StudentID:   rank.StudentID,
			StudentName: studentNameMap[rank.StudentID], // 使用已填充的 map
			Rank:        rank.Rank,
			TotalScore:  rank.TotalScore,
		})
	}

	var prevExamID uint
	tx.Model(&models.Exam{}).
		Select("id").
		Where("class_id = (SELECT class_id FROM exams WHERE id = ?) AND created_at < (SELECT created_at FROM exams WHERE id = ?)", examID, examID).
		Order("created_at DESC").
		Limit(1).
		Scan(&prevExamID)

	prevRankMap := make(map[uint]int)
	if prevExamID != 0 {
		var prevRanks []models.Rank
		// 优化：只查询需要的字段
		if err := tx.Model(&models.Rank{}).Select("student_id", "`rank`").Where("exam_id = ?", prevExamID).Find(&prevRanks).Error; err == nil {
			for _, r := range prevRanks {
				prevRankMap[r.StudentID] = r.Rank
			}
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			// 记录错误，但不中断流程
			fmt.Printf("获取上次考试 %d 排名失败: %v\n", prevExamID, err)
		}
	}

	for _, rank := range currentRanks {
		studentName, nameExists := studentNameMap[rank.StudentID]
		if !nameExists {
			// 理论上不应该发生
			continue
		}
		if prevRank, exists := prevRankMap[rank.StudentID]; exists {
			rankChange := prevRank - rank.Rank // 正数表示进步，负数表示退步
			change := models.RankChange{
				StudentName: studentName,
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

	sort.Slice(resp.ImproveRank, func(i, j int) bool {
		return resp.ImproveRank[i].Change > resp.ImproveRank[j].Change
	})
	sort.Slice(resp.DeclineRank, func(i, j int) bool {
		return resp.DeclineRank[i].Change > resp.DeclineRank[j].Change
	})

	if len(resp.ImproveRank) > 10 {
		resp.ImproveRank = resp.ImproveRank[:10]
	}
	if len(resp.DeclineRank) > 10 {
		resp.DeclineRank = resp.DeclineRank[:10]
	}

	tx.Commit()
	return resp, nil
}

// calculateStdDev 计算一组分数的标准差。
// scores: 分数切片。
// mean: 分数的平均值。
// 返回标准差。如果分数切片为空，返回 0。
func calculateStdDev(scores []float64, mean float64) float64 {
	n := len(scores)
	if n == 0 {
		return 0
	}
	variance := 0.0
	for _, score := range scores {
		diff := score - mean
		variance += diff * diff
	}
	return math.Sqrt(variance / float64(n))
}

// calculateMedian 计算一组分数的中位数。
// scores: 分数切片。
// 返回中位数。如果分数切片为空，返回 0。
func calculateMedian(scores []float64) float64 {
	n := len(scores)
	if n == 0 {
		return 0
	}
	// 复制一份以避免修改原切片
	sorted := make([]float64, n)
	copy(sorted, scores)
	sort.Float64s(sorted)

	mid := n / 2
	if n%2 == 0 { // 偶数个
		return (sorted[mid-1] + sorted[mid]) / 2
	}
	// 奇数个
	return sorted[mid]
}

// calculateScoreBuckets 计算分数段分布。
// 将分数按每 10 分一个区间进行分组统计。
// scores: 分数切片。
// 返回一个 JSON 对象，键为分数段（如 "90-99"），值为该分数段的人数。
func calculateScoreBuckets(scores []float64) datatypes.JSON {
	buckets := make(map[int]int)

	for _, score := range scores {
		// 处理边界情况，例如满分
		var lower int
		if score == 100 { // 假设满分100，特殊处理
			lower = 90 // 归入 90-99 或单独处理为 100
		} else {
			lower = int(math.Floor(score/10)) * 10
		}
		// 确保 lower 不会小于 0
		if lower < 0 {
			lower = 0
		}
		buckets[lower]++
	}

	keys := make([]int, 0, len(buckets))
	for k := range buckets {
		keys = append(keys, k)
	}
	sort.Ints(keys) // 按分数段升序排序

	ordered := make(map[string]int) // 使用 map[string]int 以便 JSON 序列化
	for _, k := range keys {
		upper := k + 9
		// 特殊处理 100 分的情况，如果需要显示为 "100" 而不是 "100-109"
		// if k == 100 { key = "100" } else { key = fmt.Sprintf("%d-%d", k, upper) }
		key := fmt.Sprintf("%d-%d", k, upper)
		ordered[key] = buckets[k]
	}

	jsonData, err := json.Marshal(ordered)
	if err != nil {
		// 处理序列化错误，可以返回空 JSON 或记录日志
		fmt.Printf("无法序列化分数段: %v\n", err)
		return datatypes.JSON("[]")
	}
	return datatypes.JSON(jsonData)
}
