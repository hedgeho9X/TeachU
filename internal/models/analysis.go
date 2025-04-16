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
	Keypoint     string  `json:"keypoint"`      // 知识点
	Score        float64 `json:"score"`         // 学生得分
	AverageScore float64 `json:"average_score"` // 平均分
	TotalScore   float64 `json:"total_score"`   //总分
	ScoreRate    float64 `json:"score_rate"`    //  该知识点得分率
	IsHigh       bool    `json:"is_high"`       // 是否高于平均分
	MasteryLevel string  `json:"mastery_level"`
}
type StudentHistory struct {
	ExamID       uint      `json:"exam_id"`
	ExamName     string    `json:"exam_name"`
	Score        float64   `json:"score"`
	AverageScore float64   `json:"average_score"`
	Time         time.Time `json:"time"`
}

type StuRankResp struct {
	ExamID      uint         `json:"exam_id"`
	StuRank     []Rank       `json:"rank_list"`    // 添加JSON标签
	ImproveRank []RankChange `json:"top_progress"` // 修正拼写并添加标签
	DeclineRank []RankChange `json:"top_regress"`  // 修正拼写并添加标签
}

type Rank struct {
	ExamID      uint    `json:"exam_id"`
	StudentID   uint    `json:"student_id"`
	StudentName string  `json:"student_name"`
	Rank        int     `json:"rank"`
	TotalScore  float64 `json:"total_score"`
}

type RankChange struct {
	StudentName string `json:"student_name"`
	Rank        int    `json:"current_rank"`
	Change      uint   `json:"change_rank"` // 变化绝对值
}
