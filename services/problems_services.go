package services

import (
	"errors"
	"github.com/Hedgeho9X/TeachU/config"
	"github.com/Hedgeho9X/TeachU/models"
	"gorm.io/gorm"
)

func CreatProblems(problems []models.Problems) error {
	var exam models.Exam
	//检查试题是否存在
	if err := config.DB.First(&exam, problems[0].ExamId).Error; err != nil {
		return errors.New("试题不存在")
	}
	//return config.DB.Create(problems).Error
	return config.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&problems).Error; err != nil {
			return err
		}
		return nil
	})
}
