package controllers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"

	"github.com/Hedgeho9X/TeachU/services"
)

// 推荐在环境变量或配置中存储 Secret
var jwtSecret = []byte("your-secret-key")

// 自定义 Claims，存放手机号信息
type AuthClaims struct {
	PhoneNumber string `json:"phone_number"`
	jwt.RegisteredClaims
}

// Register 注册
func Register(c *gin.Context) {
	var input struct {
		PhoneNumber string `json:"phone_number"`
		Password    string `json:"password"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}
	if input.PhoneNumber == "" || input.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "手机号或密码不能为空"})
		return
	}

	user, err := services.CreateUser(input.PhoneNumber, input.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "注册失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "注册成功",
		"user":    user.PhoneNumber,
	})
}

// Login 登录
func Login(c *gin.Context) {
	var input struct {
		PhoneNumber string `json:"phone_number" binding:"required"`
		Password    string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		fmt.Println("请求参数解析失败:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	// 查找用户
	user, err := services.GetUserByPhoneNumber(input.PhoneNumber)
	if err != nil {
		fmt.Printf("用户查找失败: %v\n", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "手机号或密码错误"})
		return
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		fmt.Printf("密码验证失败: %v\n", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "手机号或密码错误"})
		return
	}

	fmt.Printf("用户 %s 登录成功，准备生成 Token\n", user.PhoneNumber)

	// 生成 Token
	token, err := services.GenerateToken(user.ID, user.PhoneNumber)
	if err != nil {
		fmt.Printf("Token 生成失败: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Token 生成失败",
		})
		return
	}
	fmt.Printf("Token 生成成功: %v\n", token)
	c.JSON(http.StatusOK, gin.H{
		"token":   token,
		"message": "登录成功",
	})
}

// Profile 受保护接口示例
func Profile(c *gin.Context) {
	userID, _ := c.Get("userID")
	username, _ := c.Get("username")

	fmt.Printf("访问个人资料，用户 ID: %v, 用户名: %v\n", userID, username)

	c.JSON(http.StatusOK, gin.H{
		"user_id":  userID,
		"username": username,
	})
}
