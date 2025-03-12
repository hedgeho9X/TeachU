package controllers

import (
	"fmt"
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"

	"github.com/Hedgeho9X/TeachU/services"
)

// 推荐在环境变量或配置中存储 Secret
// var jwtSecret = []byte("your-secret-key")

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
		Email           string `json:"email"`
		Code            string `json:"code"`
	}
	//输入校验
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 0, "error": "请求参数错误"})
		return
	}
	if input.PhoneNumber == "" || input.Password == "" || input.PasswordConfirm == "" || input.Username == "" {
		c.JSON(http.StatusOK, gin.H{"code": 0, "error": "填写信息不能为空"})
		return
	}
	if input.Password != input.PasswordConfirm {
		c.JSON(http.StatusOK, gin.H{"code": 0, "error": "两次输入的密码不一致"})
		return
	}
	// 检查密码是否符合规则：8-20位，必须包含字母和数字
	if len(input.Password) < 8 || len(input.Password) > 20 {
		c.JSON(http.StatusOK, gin.H{"code": 0, "error": "密码长度必须在8-20位之间"})
		return
	}
	emailPattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, err := regexp.MatchString(emailPattern, input.Email)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":  0,
			"error": "邮箱格式校验失败"},
		)
		return
	}
	if !matched {
		c.JSON(http.StatusOK, gin.H{
			"code":  0,
			"error": "邮箱格式不正确"},
		)
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
		c.JSON(http.StatusOK, gin.H{"code": 0, "error": "密码必须包含字母和数字"})
		return
	}
	email := services.NewEmailVerificationService(services.EmailConfig{
		From:      "rjl7@qq.com", // 改为小写
		FromAlias: "EduSpark",
		Password:  "nqwryufsseyxfaei",
		Host:      "smtp.qq.com",
		Port:      465, // 改为 SSL 端口
	})
	if !email.VerifyCode(input.Email, input.Code) {
		c.JSON(http.StatusOK, gin.H{
			"code":  0,
			"error": "验证码验证失败"},
		)
		return
	}
	// 调用 service 层执行注册
	_, err = services.CreateUser(input.PhoneNumber, input.Password, input.Username, input.Email)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 0, "error": "注册失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    1,
		"message": "注册成功",
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
		c.JSON(http.StatusOK, gin.H{"code": 0, "error": "请求参数错误"})
		return
	}

	// 查找用户
	user, err := services.GetUserByPhoneNumber(input.PhoneNumber)
	if err != nil {
		fmt.Printf("用户查找失败: %v\n", err)
		c.JSON(http.StatusOK, gin.H{"code": 0, "error": "该账号未注册"})
		return
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		fmt.Printf("密码验证失败: %v\n", err)
		c.JSON(http.StatusOK, gin.H{"code": 0, "error": "手机号或密码错误"})
		return
	}

	fmt.Printf("用户 %s 登录成功，准备生成 Token\n", user.PhoneNumber)

	// 生成 Token
	token, err := services.GenerateToken(uint(user.ID), user.PhoneNumber)
	if err != nil {
		fmt.Printf("Token 生成失败: %v\n", err)
		c.JSON(http.StatusOK, gin.H{
			"code":  0,
			"error": "Token 生成失败",
		})
		return
	}
	fmt.Printf("Token 生成成功: %v\n", token)
	c.JSON(http.StatusOK, gin.H{
		"code":    1,
		"token":   token,
		"message": "登录成功",
	})
}

func ResetPassword(c *gin.Context) {
	// 解析请求 JSON
	var input struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":  0,
			"error": "请求参数错误"},
		)
		return
	}
	// 获取当前登录用户 ID
	userID, _ := c.Get("userID")
	// 调用 service 层执行密码校验和更新
	if err := services.ResetPasswordService(userID.(uint), input.OldPassword, input.NewPassword); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":  0,
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    1,
		"message": "密码重置成功",
	})
}

func SendCode(c *gin.Context) {
	// 解析请求 JSON
	var input struct {
		Email string `json:"email"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":  0,
			"error": "请求参数错误"},
		)
		return
	}
	//校验
	if input.Email == "" {
		c.JSON(http.StatusOK, gin.H{
			"code":  0,
			"error": "邮箱不能为空"},
		)
		return
	}
	//正则表达校验邮箱
	emailPattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, err := regexp.MatchString(emailPattern, input.Email)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":  0,
			"error": "邮箱格式校验失败"},
		)
		return
	}
	if !matched {
		c.JSON(http.StatusOK, gin.H{
			"code":  0,
			"error": "邮箱格式不正确"},
		)
		return
	}
	// 调用 service 层发送验证码
	email := services.NewEmailVerificationService(services.EmailConfig{
		From:      "rjl7@qq.com", // 改为小写
		FromAlias: "EduSpark",
		Password:  "nqwryufsseyxfaei",
		Host:      "smtp.qq.com",
		Port:      465, // 改为 SSL 端口
	})
	if _, err := email.SendCode(input.Email); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":  0,
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    1,
		"message": "验证码发送成功",
	})
}

func VerifyCode(c *gin.Context) {
	// 解析请求 JSON
	var input struct {
		Email string `json:"email"`
		Code  string `json:"code"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":  0,
			"error": "请求参数错误"},
		)
		return
	}
	emailPattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, err := regexp.MatchString(emailPattern, input.Email)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":  0,
			"error": "邮箱格式校验失败"},
		)
		return
	}
	if !matched {
		c.JSON(http.StatusOK, gin.H{
			"code":  0,
			"error": "邮箱格式不正确"},
		)
		return
	}
	email := services.NewEmailVerificationService(services.EmailConfig{
		From:      "rjl7@qq.com", // 改为小写
		FromAlias: "EduSpark",
		Password:  "nqwryufsseyxfaei",
		Host:      "smtp.qq.com",
		Port:      465, // 改为 SSL 端口
	})
	// 调用 service 层验证验证码
	// 验证邮箱验证码
	Ok := email.VerifyCode(input.Email, input.Code)
	if !Ok {
		c.JSON(http.StatusOK, gin.H{
			"code":  0,
			"error": fmt.Sprintf("验证码验证失败."),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    1,
		"message": "验证码验证成功",
	})
}

// ForgetPassword 通过邮箱验证码重置密码
func ForgetPassword(c *gin.Context) {
	// 解析请求 JSON
	var input struct {
		Email       string `json:"email"`
		Code        string `json:"code"`
		NewPassword string `json:"new_password"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":  0,
			"error": "请求参数错误"},
		)
		return
	}

	// 校验邮箱格式
	emailPattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, err := regexp.MatchString(emailPattern, input.Email)
	if err != nil || !matched {
		c.JSON(http.StatusOK, gin.H{
			"code":  0,
			"error": "邮箱格式不正确"},
		)
		return
	}

	// 校验密码格式
	if len(input.NewPassword) < 8 || len(input.NewPassword) > 20 {
		c.JSON(http.StatusOK, gin.H{
			"code":  0,
			"error": "密码长度必须在8-20位之间"},
		)
		return
	}

	// 检查是否同时包含字母和数字
	hasLetter := false
	hasNumber := false
	for _, char := range input.NewPassword {
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') {
			hasLetter = true
		}
		if char >= '0' && char <= '9' {
			hasNumber = true
		}
	}

	if !hasLetter || !hasNumber {
		c.JSON(http.StatusOK, gin.H{
			"code":  0,
			"error": "密码必须包含字母和数字"},
		)
		return
	}

	// 验证邮箱验证码
	email := services.NewEmailVerificationService(services.EmailConfig{
		From:      "rjl7@qq.com", // 改为小写
		FromAlias: "EduSpark",
		Password:  "nqwryufsseyxfaei",
		Host:      "smtp.qq.com",
		Port:      465, // 改为 SSL 端口
	})
	if !email.VerifyCode(input.Email, input.Code) {
		c.JSON(http.StatusOK, gin.H{
			"code":  0,
			"error": "验证码验证失败"},
		)
		return
	}

	// 根据邮箱查找用户
	user, err := services.GetUserByEmail(input.Email)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":  0,
			"error": "该邮箱未注册"},
		)
		return
	}

	// 更新用户密码
	if err := services.UpdatePassword(user.ID, input.NewPassword); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":  0,
			"error": "密码重置失败: " + err.Error()},
		)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    1,
		"message": "密码重置成功"},
	)
}
