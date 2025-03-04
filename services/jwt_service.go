package services

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"

	"github.com/Hedgeho9X/TeachU/config"
	"github.com/Hedgeho9X/TeachU/models"
)

func GenerateToken(userID uint, username string) (string, error) {
	fmt.Printf("正在为用户生成 Token: ID=%d, 用户名=%s\n", userID, username)

	claims := models.Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
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

	fmt.Printf("Token 生成成功: %s\n", tokenString)
	return tokenString, nil
}
