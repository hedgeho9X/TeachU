package models

import (
	"time"
)

// User 与 teach_u 数据库中的 users 表结构对应
type User struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	PhoneNumber  string    `gorm:"size:20;unique;not null" json:"phone_number"`
	PasswordHash string    `gorm:"size:255;not null" json:"password_hash"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
}
