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
	// 解析请求 JSON
	var input struct {
		PhoneNumber     string `json:"phone_number"`
		Password        string `json:"password"`
		PasswordConfirm string `json:"password_confirm"`
		Username        string `json:"username"`
	}
	//输入校验
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}
	if input.PhoneNumber == "" || input.Password == "" || input.PasswordConfirm == "" || input.Username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "填写信息不能为空"})
		return
	}
	if input.Password != input.PasswordConfirm {
		c.JSON(http.StatusBadRequest, gin.H{"error": "两次输入的密码不一致"})
		return
	}
	// 检查密码是否符合规则：8-20位，必须包含字母和数字
	if len(input.Password) < 8 || len(input.Password) > 20 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "密码长度必须在8-20位之间"})
		return
	}

	// 检查是否同时包含字母和数字
	hasLetter := false
	hasNumber := false
	for _, char := range input.Password {
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') {
			hasLetter = true
		}
		if char >= '0' && char <= '9' {
			hasNumber = true
		}
	}

	if !hasLetter || !hasNumber {
		c.JSON(http.StatusBadRequest, gin.H{"error": "密码必须包含字母和数字"})
		return
	}

	// 调用 service 层执行注册
	user, err := services.CreateUser(input.PhoneNumber, input.Password, input.Username)
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
	token, err := services.GenerateToken(uint(user.ID), user.PhoneNumber)
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

func ResetPassword(c *gin.Context) {
	// 解析请求 JSON
	var input struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}
	// 获取当前登录用户 ID
	userID, _ := c.Get("userID")
	// 调用 service 层执行密码校验和更新
	if err := services.ResetPasswordService(userID.(uint), input.OldPassword, input.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "密码重置成功",
	})
}
