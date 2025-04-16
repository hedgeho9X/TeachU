package models

import (
	"time"
)

// Class 班级模型
type Class struct {
	ID            uint      `gorm:"primaryKey;column:id"`
	ClassNumber   int       `gorm:"not null;column:class_number"`    // 班号
	GradeLevel    int       `gorm:"not null;column:grade_level"`     // 年级
	CreatedUserID uint      `gorm:"not null;column:created_user_id"` // 创建者ID
	Students      []Student `gorm:"foreignKey:ClassID"`              // 班级学生列表
	CreatedAt     time.Time `gorm:"column:created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at"`
}

// TableName 指定表名
func (Class) TableName() string {
	return "classes"
}
