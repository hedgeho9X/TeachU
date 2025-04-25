package services

import (
	"context"
	"crypto/tls" // 添加 tls 包
	"errors"
	"fmt"
	"math/rand"
	"net/smtp"
	"os"
	"strings"
	"time"

	"github.com/Hedgeho9X/TeachU/internal/config"
	"github.com/Hedgeho9X/TeachU/internal/models"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

// CreateUser 根据用户对象和明文密码创建新用户
func CreateUser(user *models.User, plainPassword string) (*models.User, error) {
	// 检查手机号是否已被注册
	if err := config.DB.Where("phone_number = ?", user.PhoneNumber).First(&models.User{}).Error; err == nil {
		return nil, errors.New("该手机号已被注册")
	}

	// 检查邮箱是否已被注册
	if err := config.DB.Where("email = ?", user.Email).First(&models.User{}).Error; err == nil {
		return nil, errors.New("该邮箱已被注册")
	}

	// 对密码进行哈希加密
	hash, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user.PasswordHash = string(hash)

	// 插入数据库
	if err := config.DB.Create(user).Error; err != nil {
		return nil, err
	}

	return user, nil
}

func GetUsernameByUserID(userID uint) (string, error) {
	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		return "", err
	}
	return user.Username, nil
}
func GetUserByuserID(userID uint) (*models.User, error) {
	var user models.User
	err := config.DB.Where("id =?", userID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByPhoneNumber 根据手机号获取用户
func GetUserByPhoneNumber(phoneNumber string) (*models.User, error) {
	var user models.User
	err := config.DB.Where("phone_number = ?", phoneNumber).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}
func GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	err := config.DB.Where("email =?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// ResetPasswordService 重置密码（需要旧密码验证）
func ResetPasswordService(userID uint, oldPassword, newPassword string) error {
	// 1. 查找用户
	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		return errors.New("找不到该用户")
	}
	// 2. 验证旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return errors.New("旧密码错误")
	}
	// 3. 哈希新密码
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	// 4. 更新密码
	user.PasswordHash = string(hash)
	if err := config.DB.Save(&user).Error; err != nil {
		return err
	}
	return nil
}

// UpdatePassword 直接更新密码（不需要旧密码验证，用于忘记密码功能）
func UpdatePassword(userID uint64, newPassword string) error {
	// 1. 查找用户
	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		return errors.New("数据库：找不到该用户")
	}

	// 2. 检查密码是否符合规则：8-20位，必须包含字母和数字
	if len(newPassword) < 8 || len(newPassword) > 20 {
		return errors.New("密码长度必须在8-20位之间")
	}

	// 检查是否同时包含字母和数字
	hasLetter := false
	hasNumber := false
	for _, char := range newPassword {
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') {
			hasLetter = true
		}
		if char >= '0' && char <= '9' {
			hasNumber = true
		}
	}

	if !hasLetter || !hasNumber {
		return errors.New("密码必须包含字母和数字")
	}

	// 3. 哈希新密码
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// 4. 更新密码
	user.PasswordHash = string(hash)
	if err := config.DB.Save(&user).Error; err != nil {
		return err
	}

	return nil
}

// VerificationService 验证码服务接口
// 使用方法:
// 1. 实现该接口的具体验证服务（如EmailVerificationService）
// 2. SendCode方法用于发送验证码到指定接收者，返回生成的验证码和可能的错误
// 3. VerifyCode方法用于验证用户输入的验证码是否正确
type VerificationService interface {
	// SendCode 发送验证码到指定接收者
	// recipient: 接收者（如邮箱地址或手机号）
	// 返回值: 生成的验证码和可能的错误
	SendCode(recipient string) (string, error)

	// VerifyCode 验证用户输入的验证码
	// recipient: 接收者（如邮箱地址或手机号）
	// code: 用户输入的验证码
	// 返回值: 验证是否成功
	VerifyCode(recipient string, code string) bool
}

// EmailVerificationService 邮件验证码服务实现
type EmailVerificationService struct {
	emailConfig EmailConfig
}

// EmailConfig 邮件配置
type EmailConfig struct {
	From      string
	FromAlias string
	Password  string
	Host      string
	Port      int
}

// NewEmailVerificationService 创建新的邮件验证码服务
func NewEmailVerificationService(config EmailConfig) *EmailVerificationService {
	return &EmailVerificationService{
		emailConfig: config,
	}
}

// SendCode 发送验证码
func (s *EmailVerificationService) SendCode(recipient string) (string, error) {
	// 生成验证码
	code := generateRandomCode()

	// 读取HTML模板
	htmlContent, err := os.ReadFile("internal/services/VerifyCode.html")
	if err != nil {
		return "", fmt.Errorf("读取模板失败: %v", err)
	}

	// 替换验证码和支持邮箱
	emailContent := strings.Replace(string(htmlContent), "{验证码}", code, 1)
	emailContent = strings.Replace(emailContent, "{支持邮箱}", s.emailConfig.From, 1)

	// 构建邮件内容
	email := EmailInfo{
		From:        s.emailConfig.From,
		FromAlias:   s.emailConfig.FromAlias,
		Password:    s.emailConfig.Password,
		Host:        s.emailConfig.Host,
		Port:        s.emailConfig.Port,
		To:          []string{recipient},
		Subject:     "EduSpark 验证码",
		Content:     emailContent,
		ContentType: "html", // 设置为html类型
	}

	// 保存验证码到Redis
	ctx := context.Background()
	err = config.RDB.Set(ctx, "verify:"+recipient, code, 5*time.Minute).Err()
	if err != nil {
		return "", fmt.Errorf("保存验证码失败: %v", err)
	}

	// 发送邮件
	if err := smtpSend(email); err != nil {
		return "", fmt.Errorf("发送邮件失败: %v", err)
	}

	return code, nil
}

// VerifyCode 验证验证码
func (s *EmailVerificationService) VerifyCode(recipient string, code string) bool {
	ctx := context.Background()
	storedCode, err := config.RDB.Get(ctx, "verify:"+recipient).Result()
	if err == redis.Nil {
		return false // 验证码不存在或已过期
	}
	if err != nil {
		return false // Redis错误
	}
	return storedCode == code
}

// generateRandomCode 生成6位随机验证码
func generateRandomCode() string {
	rand.Seed(time.Now().UnixNano())
	code := rand.Intn(1000000)
	return fmt.Sprintf("%06d", code)
}

// smtpSend 发送邮件
// EmailInfo 邮件信息结构体
type EmailInfo struct {
	From        string   // 发件人邮箱
	FromAlias   string   // 发件人别名
	Password    string   // 邮箱密码
	Host        string   // SMTP服务器地址
	Port        int      // SMTP端口
	To          []string // 收件人列表
	Subject     string   // 邮件主题
	Content     string   // 邮件内容
	ContentType string   // 内容类型
}

func smtpSend(email EmailInfo) error {
	// 设置 PlainAuth
	auth := smtp.PlainAuth("", email.From, email.Password, email.Host)

	// 创建 tls 配置
	tlsconfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         email.Host,
	}

	// 连接到 SMTP 服务器
	conn, err := tls.Dial("tcp", fmt.Sprintf("%s:%d", email.Host, email.Port), tlsconfig)
	if err != nil {
		return fmt.Errorf("TLS 连接失败: %v", err)
	}
	defer conn.Close()

	// 创建 SMTP 客户端
	client, err := smtp.NewClient(conn, email.Host)
	if err != nil {
		return fmt.Errorf("SMTP 客户端创建失败: %v", err)
	}
	defer client.Quit()

	// 使用 auth 进行认证
	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("认证失败: %v", err)
	}

	// 设置发件人
	if err = client.Mail(email.From); err != nil {
		return fmt.Errorf("发件人设置失败: %v", err)
	}

	// 设置收件人
	for _, to := range email.To {
		if err = client.Rcpt(to); err != nil {
			return fmt.Errorf("收件人设置失败: %v", err)
		}
	}

	// 写入邮件内容
	wc, err := client.Data()
	if err != nil {
		return fmt.Errorf("数据写入失败: %v", err)
	}
	defer wc.Close()

	contentType := "Content-Type: text/plain; charset=UTF-8"
	if email.ContentType == "html" {
		contentType = "Content-Type: text/html; charset=UTF-8"
	}

	msg := "To: " + strings.Join(email.To, ",") + "\r\n" +
		"From: " + email.FromAlias + "<" + email.From + ">\r\n" +
		"Subject: " + email.Subject + "\r\n" +
		contentType + "\r\n\r\n" +
		email.Content

	if _, err = wc.Write([]byte(msg)); err != nil {
		return fmt.Errorf("消息发送失败: %v", err)
	}

	return nil
}
