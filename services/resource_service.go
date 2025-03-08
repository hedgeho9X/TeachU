package services

import (
	"context"
	"strconv"
	"time"

	"github.com/Hedgeho9X/TeachU/config"
	"github.com/Hedgeho9X/TeachU/models"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
)

func SearchResources(pageNum, pageSize, q, subject, grade string) ([]models.Resource, int64, error) {
	var resources []models.Resource
	var total int64
	query := config.DB

	// 构建查询条件
	if subject == "全部" {
		subject = ""
	}
	if grade == "全部" {
		grade = ""
	}
	if q == "" {
		q = ""
	}
	if subject != "" {
		query = query.Where("subject = ?", subject)
	}

	if grade != "" {
		query = query.Where("grade = ?", grade)
	}

	if q != "" {
		query = query.Where("object_key LIKE ? OR file_name LIKE ?", "%"+q+"%", "%"+q+"%")
	}

	// 获取总数
	err := query.Model(&models.Resource{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
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
	err = query.Offset(offset).Limit(limit).Find(&resources).Error
	if err != nil {
		return nil, 0, err
	}

	return resources, total, nil
}

func GetResourceUseExternal(objectName string) (string, error) {
	region := "cn-chengdu"
	bucketName := "teachu"
	// objectName = ""
	// 加载默认配置并设置凭证提供者和区域
	cfg := oss.LoadDefaultConfig().
		WithCredentialsProvider(credentials.NewEnvironmentVariableCredentialsProvider()).
		WithRegion(region).
		WithEndpoint("http://oss.hedgeho9.cn").
		WithUseCName(true)
	// 创建OSS客户端
	client := oss.NewClient(cfg)
	// 生成GetObject的预签名URL
	result, err := client.Presign(context.TODO(), &oss.GetObjectRequest{
		Bucket: oss.Ptr(bucketName),
		Key:    oss.Ptr(objectName),
		//RequestPayer: oss.Ptr("requester"), // 指定请求者身份
	},
		oss.PresignExpires(10*time.Minute),
	)
	if err != nil {
		return "", err
	}
	// log.Printf("request method:%v\n", result.Method)
	// log.Printf("request expiration:%v\n", result.Expiration)
	// log.Printf("request url:%v\n", result.URL)
	// if len(result.SignedHeaders) > 0 {
	// 	//当返回结果包含预签名头时，使用预签名URL发送GET请求时也包含相应的请求头，以免出现不一致，导致请求失败和预签名错误
	// 	log.Printf("signed headers:\n")
	// 	for k, v := range result.SignedHeaders {
	// 		log.Printf("%v: %v\n", k, v)
	// 	}
	// }

	return result.URL, nil
}
