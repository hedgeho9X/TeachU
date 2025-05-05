package controllers

import (
	"fmt"
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"

	"github.com/Hedgeho9X/TeachU/internal/models"
	"github.com/Hedgeho9X/TeachU/internal/services"
)

// AuthClaims 定义了 JWT 的自定义 Claims，包含用户 ID。
type AuthClaims struct {
	UserID uint64 `json:"user_id"`
	jwt.RegisteredClaims
}

// Register 处理用户注册请求。
// 它接收用户的手机号、密码、确认密码、用户名、学科、邮箱和验证码，
// 进行校验后创建新用户。
func Register(c *gin.Context) {
	// 解析请求 JSON
	var input struct {
		PhoneNumber     string `json:"phone_number"`
		Password        string `json:"password"`
		PasswordConfirm string `json:"password_confirm"`
		Username        string `json:"username"`
		Subject         string `json:"subject"`
		Email           string `json:"email"`
		Code            string `json:"code"`
	}
	//输入校验
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 0, "error": "请求参数错误"})
		return
	}
	// 检查必填字段
	fieldNames := map[string]string{
		"phone_number":     "手机号",
		"password":         "密码",
		"password_confirm": "确认密码",
		"username":         "用户名",
		"email":            "邮箱",
		"code":             "验证码",
		"subject":          "学科",
	}
	requiredFields := map[string]string{
		"phone_number":     input.PhoneNumber,
		"password":         input.Password,
		"password_confirm": input.PasswordConfirm,
		"username":         input.Username,
		"email":            input.Email,
		"code":             input.Code,
		"subject":          input.Subject,
	}
	for field, value := range requiredFields {
		if value == "" {
			c.JSON(http.StatusOK, gin.H{
				"code":  0,
				"error": fmt.Sprintf("%s不能为空", fieldNames[field]),
			})
			return
		}
	}
	// 检查手机号格式
	if ok, _ := regexp.MatchString(`^1[3456789]\d{9}$`, input.PhoneNumber); !ok {
		c.JSON(http.StatusOK, gin.H{"code": 0, "error": "手机号格式错误"})
		return
	}

	// 检查学科必须为语数外物化生政史地
	if ok, _ := regexp.MatchString(`^(语文|数学|英语|物理|化学|生物|政治|历史|地理)$`, input.Subject); !ok {
		c.JSON(http.StatusOK, gin.H{"code": 0, "error": "学科信息有误"})
		return
	}

	// 检查密码是否相同
	if input.Password != input.PasswordConfirm {
		c.JSON(http.StatusOK, gin.H{"code": 0, "error": "两次输入的密码不一致"})
		return
	}
	// 检查密码是否符合规则：8-20位，必须包含字母和数字
	if len(input.Password) < 8 || len(input.Password) > 20 {
		c.JSON(http.StatusOK, gin.H{"code": 0, "error": "密码长度必须在8-20位之间"})
		return
	}
	//检查邮箱正则
	emailPattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, err := regexp.MatchString(emailPattern, input.Email)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":  0,
			"error": err},
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
	//验证邮箱验证码
	email := services.NewEmailVerificationService(services.EmailConfig{
		From:      "rjl7@qq.com",
		FromAlias: "EduSpark",
		Password:  "nqwryufsseyxfaei",
		Host:      "smtp.qq.com",
		Port:      465,
	})
	if !email.VerifyCode(input.Email, input.Code) {
		c.JSON(http.StatusOK, gin.H{
			"code":  0,
			"error": "验证码验证失败"},
		)
		return
	}

	userObj := &models.User{
		PhoneNumber: input.PhoneNumber,
		Username:    input.Username,
		Email:       input.Email,
		Subject:     input.Subject,
	}
	_, err = services.CreateUser(userObj, input.Password)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 0, "error": "注册失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    1,
		"message": "注册成功",
	})
}

// Login 处理用户通过手机号和密码登录的请求。
// 验证成功后返回 JWT Token。
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
		c.JSON(http.StatusOK, gin.H{"code": 0, "error": "该账号未注册"})
		return
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 0, "error": "手机号或密码错误"})
		return
	}

	// 生成 Token
	token, err := services.GenerateToken(uint(user.ID))
	if err != nil {
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

// Email2Login 处理用户通过邮箱和验证码登录的请求。
// 验证成功后返回 JWT Token。
func Email2Login(c *gin.Context) {
	var input struct {
		Email string `json:"email" binding:"required"`
		Code  string `json:"code" binding:"required"`
	}
	//解析参数
	if err := c.ShouldBind(&input); err != nil {
		fmt.Println("请求参数解析失败:", err)
		c.JSON(http.StatusOK, gin.H{"code": 0, "error": "请求参数错误"})
		return
	}
	//校验参数
	if input.Email == "" || input.Code == "" {
		c.JSON(http.StatusOK, gin.H{"code": 0, "error": "邮箱或验证码不能为空"})
		return
	}
	//查询用户是否已注册
	user, err := services.GetUserByEmail(input.Email)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":  0,
			"error": "该邮箱未注册"},
		)
		return
	}
	//验证邮箱验证码
	email := services.NewEmailVerificationService(services.EmailConfig{
		From:      "rjl7@qq.com",
		FromAlias: "EduSpark",
		Password:  "nqwryufsseyxfaei",
		Host:      "smtp.qq.com",
		Port:      465,
	})
	if !email.VerifyCode(input.Email, input.Code) {
		c.JSON(http.StatusOK, gin.H{
			"code":  0,
			"error": "验证码验证失败.",
		})
		return // 验证码失败时应返回
	}
	//验证成功，返回token
	token, err := services.GenerateToken(uint(user.ID))
	if err != nil {
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

// ResetPassword 处理已登录用户重置密码的请求。
// 需要提供旧密码和新密码。
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
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{ // 使用正确的状态码
			"code":  0,
			"error": "用户未登录",
		})
		return
	}
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

// ForgetPassword 处理用户通过邮箱验证码忘记密码并重置密码的请求。
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
		From:      "rjl7@qq.com",
		FromAlias: "EduSpark",
		Password:  "nqwryufsseyxfaei",
		Host:      "smtp.qq.com",
		Port:      465,
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

// SendCode 处理发送邮箱验证码的请求。
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
		From:      "rjl7@qq.com",
		FromAlias: "EduSpark",
		Password:  "nqwryufsseyxfaei",
		Host:      "smtp.qq.com",
		Port:      465,
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

// VerifyCode 处理验证邮箱验证码的请求。
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
		From:      "rjl7@qq.com",
		FromAlias: "EduSpark",
		Password:  "nqwryufsseyxfaei",
		Host:      "smtp.qq.com",
		Port:      465,
	})
	// 调用 service 层验证验证码
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
