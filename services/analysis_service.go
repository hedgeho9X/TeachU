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

	// 获取班级所有学生总分
	var studentTotals []float64
	studentScores := make(map[uint]float64) // 学生ID到总分的映射

	for _, s := range scores {
		studentScores[s.StudentID] += s.Score
	}

	// 转换总分到切片
	for _, total := range studentScores {
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
		TotalScore:       total,
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
		if lower >= 150 {
			buckets[150]++
		} else {
			buckets[lower]++
		}
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
