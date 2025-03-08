package services

import (
	"strconv"

	"github.com/Hedgeho9X/TeachU/config"
	"github.com/Hedgeho9X/TeachU/models"
)

func SearchResources(pageNum, pageSize, q, subject, grade string) ([]models.Resource, error) {
	var resources []models.Resource
	query := config.DB.Table("teaching_resources")

	// 构建查询条件
	if subject != "" {
		query = query.Where("subject = ?", subject)
	}

	if grade != "" {
		query = query.Where("grade = ?", grade)
	}

	if q != "" {
		query = query.Where("object_key LIKE ? OR file_name LIKE ?", "%"+q+"%", "%"+q+"%")
	}

	// 处理分页
	offset, limit := 0, 10 // 默认值
	if pageNum != "" && pageSize != "" {
		offset, _ = strconv.Atoi(pageNum)
		limit, _ = strconv.Atoi(pageSize)
		if offset > 0 {
			offset = (offset - 1) * limit
		}
	}

	// 执行查询
	err := query.Offset(offset).Limit(limit).Find(&resources).Error
	if err != nil {
		return nil, err
	}

	return resources, nil
}
