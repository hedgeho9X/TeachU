package models

import "gorm.io/gorm"

type Exam struct {
	gorm.Model
	Id       uint   `gorm:"primaryKey;column:id"`
	UserId   uint   `gorm:"not null;column:created_user_id"`
	ClassId  uint   `gorm:"not null;column:class_id"`
	ExamName string `gorm:"not null;column:exam_name"` // 试题名称
	Subject  string `gorm:"not null;column:subject"`   // 科目
	// DeletedAt gorm.DeletedAt `gorm:"index"`
}

// TableName 指定表名为 exams
func (Exam) TableName() string {
	return "exams"
}

// Content   string `gorm:"type:text;column:content"`   // 解析后的内容
// Knowledge string `gorm:"type:text;column:knowledge"` // 知识点（可扩展为结构化数据）
