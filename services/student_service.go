package services

import (
	"github.com/Hedgeho9X/TeachU/config"
	"github.com/Hedgeho9X/TeachU/models"
)

func GetStudentsByClassID(classID uint) ([]models.Student, error) {
	var students []models.Student
	if err := config.DB.Where("class_id =?", classID).Find(&students).Error; err != nil {
		return nil, err
	}
	return students, nil
}
