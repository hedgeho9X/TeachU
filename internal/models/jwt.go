package models

import (
	"github.com/golang-jwt/jwt/v4"
)

type Claims struct {
	UserID      uint   `json:"user_id"`
	Username    string `json:"username"`
	PhoneNumber string `json:"phone_number"` // 增加json标签
	Email       string `json:"email"`         // 增加json标签
	jwt.RegisteredClaims
}
