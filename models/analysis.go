package models

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type ClassMetric struct {
	ID               uint           `gorm:"primaryKey"`
	ExamID           uint           `gorm:"index;not null"`
	StudentCount     int            `gorm:"not null"`
	TotalScore       float64        `gorm:"type:int;not null"`
	AvgTotalScore    float64        `gorm:"type:decimal(5,2);not null"`
	MaxTotalScore    float64        `gorm:"type:decimal(5,2);not null"`
	MinTotalScore    float64        `gorm:"type:decimal(5,2);not null"`
	MedianTotalScore float64        `gorm:"type:decimal(5,2);not null"`
	StdDev           float64        `gorm:"type:decimal(5,2);not null"`
	ScoreBuckets     datatypes.JSON `gorm:"type:json;not null"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        gorm.DeletedAt `gorm:"index"`
}

type KeypointMetricResp struct {
	ExamID          uint             `json:"exam_id"`
	KeypointMetrics []KeypointMetric `json:"keypoints"` // 修改字段名和类型
}

type KeypointMetric struct { // 重命名结构体
	Keypoint     string  `json:"keypoint"`
	TotalScore   float64 `json:"total_score"`   // 类型改为float64
	AverageScore float64 `json:"average_score"` // 字段名对齐SQL结果
	StudentCount int     `json:"student_count"` // 新增字段
	ScoreRate    float64 `json:"score_rate"`    // 保留得分率字段
}

type StudentAnalysisResponse struct {
	StudentID      uint               `json:"student_id"`
	TotalScore     float64            `json:"total_score"`
	StudentMetrics []StudentKeypoints `json:"student_keypoints"`
	StudentHistory []StudentHistory   `json:"student_history"`
}

type StudentKeypoints struct {
	Keypoint     string  `json:"keypoint"`
	Score        float64 `json:"score"`
	AverageScore float64 `json:"average_score"`
	ScoreRate    float64 `json:"score_rate"` // 新增字段
	IsHigh       bool    `json:"is_high"`
	MasteryLevel string  `json:"mastery_level"`
}
type StudentHistory struct {
	ExamID       uint      `json:"exam_id"`
	Score        float64   `json:"score"`
	AverageScore float64   `json:"average_score"`
	Time         time.Time `json:"time"`
}
