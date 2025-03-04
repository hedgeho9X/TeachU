package services

import (
	"errors"

	"github.com/Hedgeho9X/TeachU/config"
	"github.com/Hedgeho9X/TeachU/models"
	"golang.org/x/crypto/bcrypt"
)

// CreateUser 根据手机号和明文密码创建新用户
func CreateUser(phoneNumber, plainPassword string) (*models.User, error) {
	// 1. 检查是否已存在相同的手机号
	var existingUser models.User
	result := config.DB.Where("phone_number = ?", phoneNumber).First(&existingUser)
	if result.Error == nil {
		return nil, errors.New("该手机号已被注册")
	}

	// 2. 对密码进行哈希加密
	hash, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// 3. 插入数据库
	user := models.User{
		PhoneNumber:  phoneNumber,
		PasswordHash: string(hash),
	}
	if err := config.DB.Create(&user).Error; err != nil {
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
