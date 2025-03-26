package services

import (
	"fmt"

	"github.com/Hedgeho9X/TeachU/config"
	"github.com/Hedgeho9X/TeachU/models"
)

func ListStudents(classID uint) ([]models.Student, error) {
	var students []models.Student
	if err := config.DB.Where("class_id =?", classID).Find(&students).Error; err != nil {
		return nil, err
	}
	return students, nil
}
func CreateStudent(classID uint, student models.Student) error {
	var existingStudent models.Student
	if err := config.DB.Where("student_number = ?", student.StudentNumber).First(&existingStudent).Error; err == nil {
		return fmt.Errorf("学生已存在，学号: %d", student.StudentNumber)
	}
	student.ClassID = classID

	// 创建学生记录
	if err := config.DB.Create(&student).Error; err != nil {
		return fmt.Errorf("创建学生失败: %v", err)
	}

	return nil
}
func DeleteStudent(StudentNumber string) error {
	var student models.Student
	if err := config.DB.First(&student, StudentNumber).Error; err != nil {
		return fmt.Errorf("学生不存在")
	}
	if err := config.DB.Delete(&student).Error; err != nil {
		return fmt.Errorf("删除学生失败: %v", err)
	}
	return nil
}
