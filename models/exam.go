package models

import "time"

type Exam struct {
	ID        uint      `gorm:"primaryKey;column:id" json:"id"`
	UserId    uint      `json:"created_user_id"`
	ClassId   uint      `json:"class_id"`
	ExamName  string    `json:"exam_name"`
	Subject   string    `json:"subject"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt time.Time `gorm:"column:deleted_at" json:"deleted_at,omitempty"`
}

// TableName 指定表名为 exams
func (Exam) TableName() string {
	return "exams"
}

// Content   string `gorm:"type:text;column:content"`   // 解析后的内容
// Knowledge string `gorm:"type:text;column:knowledge"` // 知识点（可扩展为结构化数据）
