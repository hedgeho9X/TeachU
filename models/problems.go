package models

import "gorm.io/gorm"

type Problems struct {
	gorm.Model
	Id              uint    `gorm:"primaryKey;column:id"`
	ExamId          uint    `gorm:"not null;column:exam_id"`          //试题
	Keypoint        string  `gorm:"not null;column:keypoint"`         //知识点
	QuestionsNumber uint    `gorm:"not null;column:questions_number"` // 题号
	TotalScore      float64 `gorm:"not null;column:total_score"`      // 总分
	Content         string  `gorm:"not null;column:content"`          //内容
}

// TableName 指定表名为 exams
func (Problems) TableName() string {
	return "problems"
}

// Content   string `gorm:"type:text;column:content"`   // 解析后的内容
// Knowledge string `gorm:"type:text;column:knowledge"` // 知识点（可扩展为结构化数据）
