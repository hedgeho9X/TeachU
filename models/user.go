package models

import (
	"time"
)

// 在用户模型中添加 Username 字段
type User struct {
	ID           uint64 `gorm:"primaryKey"`
	PhoneNumber  string `gorm:"uniqueIndex:idx_phone;size:20;not null"` // 指定索引长度为20
	Username     string `gorm:"size:50"`
	PasswordHash string `gorm:"not null;size:255"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
