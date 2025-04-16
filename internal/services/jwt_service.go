package services

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"

	"github.com/Hedgeho9X/TeachU/internal/config"
	"github.com/Hedgeho9X/TeachU/internal/models"
)

func GenerateToken(userID uint) (string, error) {
	// fmt.Printf("正在为用户生成 Token: ID=%d, 用户名=%s\n", userID, username)

	// 生成Token时补充用户信息
	user, _ := GetUserByuserID(userID) // 需要实现该函数
	claims := models.Claims{
		UserID:      userID,
		Username:    user.Username,
		PhoneNumber: user.PhoneNumber, // 新增电话
		Email:       user.Email,       // 新增邮箱
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "TeachU",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(config.JWTSecret))

	if err != nil {
		fmt.Printf("Token 生成错误: %v\n", err)
		return "", err
	}

	// fmt.Printf("Token 生成成功: %s\n", tokenString)
	return tokenString, nil
}
