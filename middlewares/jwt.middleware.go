package middlewares

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"

	"github.com/Hedgeho9X/TeachU/config"
	"github.com/Hedgeho9X/TeachU/models"
)

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Println("开始 JWT 认证...")

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			fmt.Println("未找到 Authorization 头")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "缺少认证信息",
			})
			return
		}

		// 检查 Bearer 格式
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			fmt.Println("Authorization 格式错误")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "认证格式无效",
			})
			return
		}

		// 解析 Token
		tokenString := parts[1]
		fmt.Printf("收到 Token: %s\n", tokenString)

		claims := &models.Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("意外的签名方法: %v", token.Header["alg"])
			}
			return []byte(config.JWTSecret), nil
		})

		if err != nil {
			fmt.Printf("Token 解析错误: %v\n", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "无效的 Token",
			})
			return
		}

		if !token.Valid {
			fmt.Println("Token 无效")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Token 已过期或无效",
			})
			return
		}

		// Token 验证成功，保存用户信息到上下文
		fmt.Printf("Token 验证成功，用户 ID: %d, 用户名: %s\n", claims.UserID, claims.Username)
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Next()
	}
}
