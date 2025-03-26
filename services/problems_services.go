package services

import (
	"errors"
	"github.com/Hedgeho9X/TeachU/config"
	"github.com/Hedgeho9X/TeachU/models"
)

func CreatProblems(problems *models.Problems) error {
	var exam models.Exam
	//检查试题是否存在
	if err := config.DB.First(&exam, problems.ExamId).Error; err != nil {
		return errors.New("试题不存在")
	}
	return config.DB.Create(problems).Error
}
