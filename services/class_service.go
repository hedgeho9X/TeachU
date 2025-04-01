package services

import (
	"errors"
	"fmt"

	"github.com/Hedgeho9X/TeachU/config"
	"github.com/Hedgeho9X/TeachU/models"
)

// GetClassesByUserID 获取用户创建的班级列表
// 新增响应结构体
type ClassResp struct {
	models.Class
	StudentCount int `json:"student_count"`
}

func GetClassesByUserID(userID uint) ([]ClassResp, error) {
	var resp []ClassResp

	// 使用左连接查询并统计学生人数
	err := config.DB.Model(&models.Class{}).
		Select("classes.*, COUNT(students.id) as student_count").
		Joins("LEFT JOIN students ON students.class_id = classes.id").
		Where("classes.created_user_id = ?", userID).
		Group("classes.id").
		Find(&resp).Error

	return resp, err
}

// CreateClassInput 创建班级的输入参数

// CreateClass 创建班级
func CreateClass(ClassNumber, GradeLevel int, userID uint) error {
	// 将输入参数转换为班级模型
	// 检查是否存在相同年级和班级的记录
	var existingClass models.Class
	err := config.DB.Where("created_user_id = ? AND grade_level = ? AND class_number = ?", userID, GradeLevel, ClassNumber).First(&existingClass).Error
	if err == nil {
		return errors.New("该班级已存在")
	}
	class := models.Class{
		ClassNumber:   ClassNumber,
		GradeLevel:    GradeLevel,
		CreatedUserID: userID,
	}
	// 创建班级记录

	if err = config.DB.Create(&class).Error; err != nil {
		return fmt.Errorf("创建班级失败: %v", err)
	}

	return nil
}

// DeleteClass 删除班级
func DeleteClass(classID string, userID uint) error {
	var class models.Class

	// 查找班级是否存在
	if err := config.DB.First(&class, classID).Error; err != nil {
		return errors.New("班级不存在")
	}

	// 验证是否为班级创建者
	if class.CreatedUserID != userID {
		return errors.New("无权删除该班级")
	}

	// 检查班级是否还有学生
	// var studentCount int64
	// if err := config.DB.Model(&models.Student{}).Where("class_id = ?", classID).Count(&studentCount).Error; err != nil {
	// 	return err
	// }
	// if studentCount > 0 {
	// 	return errors.New("班级中还有学生，无法删除")
	// }

	// 开启事务
	tx := config.DB.Begin()

	// 删除班级下的所有学生
	if err := tx.Where("class_id = ?", classID).Delete(&models.Student{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("删除学生失败: %v", err)
	}

	// 删除班级
	if err := tx.Delete(&class).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("删除班级失败: %v", err)
	}

	return tx.Commit().Error
}
