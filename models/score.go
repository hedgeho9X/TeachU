package models

import (
	"time"
)

// Score 成绩模型
type Score struct {
	ID             uint      `gorm:"primaryKey;column:id"`
	StudentID      uint      `gorm:"column:student_id;not null"`              // 学生ID
	Student        Student   `gorm:"foreignKey:StudentID"`                    // 关联学生
	ExamID         uint      `gorm:"column:exam_id;not null"`                 // 考试ID
	Exam           Exam      `gorm:"foreignKey:ExamID"`                       // 关联考试
	QuestionNumber int       `gorm:"column:question_number;not null"`         // 题号
	Score          float64   `gorm:"column:score;type:decimal(5,2);not null"` // 得分
	CreatedAt      time.Time `gorm:"column:created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at"`
	DeletedAt      time.Time `gorm:"column:deleted_at"`
}

// TableName 指定表名
func (Score) TableName() string {
	return "scores"
}
