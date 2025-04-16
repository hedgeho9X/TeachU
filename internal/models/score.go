package models

import (
	"time"

	"gorm.io/gorm" // 添加这个导入
)

// Score 成绩模型
type Score struct {
	ID             uint           `gorm:"primaryKey;column:id"`
	StudentID      uint           `gorm:"column:student_id;not null"`
	Student        Student        `gorm:"foreignKey:StudentID"`
	ExamID         uint           `gorm:"column:exam_id;not null"`
	Exam           Exam           `gorm:"foreignKey:ExamID"`
	QuestionNumber int            `gorm:"column:question_number;not null"`
	Score          float64        `gorm:"column:score;type:decimal(5,2);column:score;not null"`
	CreatedAt      time.Time      `gorm:"column:created_at"`
	UpdatedAt      time.Time      `gorm:"column:updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"column:deleted_at;index"` // 修改这里
}

// TableName 指定表名
func (Score) TableName() string {
	return "scores"
}

// ScoreInput 成绩输入的数据传输对象
type ScoreInput struct {
	StudentNumber string          `json:"student_number"`
	Questions     []QuestionScore `json:"questions"`
}

// QuestionScore 题目成绩
type QuestionScore struct {
	QuestionNumber int     `json:"question_number"`
	Score          float64 `json:"score"`
}

type StudentSimple struct {
	ID            uint   `json:"student_id"`
	StudentNumber string `json:"student_number"`
	StudentName   string `json:"student_name"`
}

type StudentScoreResponse struct {
	StudentInfo StudentSimple   `json:"student_info"`
	Questions   []QuestionScore `json:"questions"`
	Total       float64         `json:"total"`
}
