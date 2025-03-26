package models

import (
	"time"
)

// Student 学生模型
type Student struct {
	ID            uint      `gorm:"primaryKey;column:id"`
	StudentName   string    `gorm:"column:student_name;size:255"`   // 学生姓名
	StudentNumber string    `gorm:"column:student_number;size:255"` // 学号
	ClassID       uint      `gorm:"column:class_id"`                // 所属班级ID
	Class         Class     `gorm:"foreignKey:ClassID"`             // 关联班级
	CreatedAt     time.Time `gorm:"column:created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at"`
}

// TableName 指定表名
func (Student) TableName() string {
	return "students"
}
