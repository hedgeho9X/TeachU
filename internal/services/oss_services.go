package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
)

// GeneratePresignedURL 为指定的 OSS 对象生成带处理参数的预签名 URL
func GeneratePresignedURL(bucketName, objectName, region, endpoint, processParams string, expiration time.Duration) (string, error) {
	// 加载默认配置并设置凭证提供者和区域
	cfg := oss.LoadDefaultConfig().
		WithCredentialsProvider(credentials.NewEnvironmentVariableCredentialsProvider()).
		WithRegion(region).
		WithEndpoint(endpoint).
		WithUseCName(true) // 假设总是使用 CName (自定义域名)

	// 创建OSS客户端
	client := oss.NewClient(cfg)

	// 生成GetObject的预签名URL
	result, err := client.Presign(context.TODO(), &oss.GetObjectRequest{
		Bucket: oss.Ptr(bucketName),
		Key:    oss.Ptr(objectName),
		// 设置文档处理参数
		Process: oss.Ptr(processParams),
	}, oss.PresignExpires(expiration)) // 使用传入的有效期

	if err != nil {
		// 返回错误而不是 Fatalf
		return "", fmt.Errorf("生成预签名 URL 失败: %w", err)
	}

	// 打印日志信息（可选，如果函数仅用于获取 URL，可以移除日志）
	log.Printf("request method:%v\n", result.Method)
	log.Printf("request expiration:%v\n", result.Expiration)
	log.Printf("request url:%v\n", result.URL)

	if len(result.SignedHeaders) > 0 {
		log.Printf("signed headers:\n")
		for k, v := range result.SignedHeaders {
			log.Printf("%v: %v\n", k, v)
		}
	}

	// 返回生成的 URL 和 nil 错误
	return result.URL, nil
}

// FileRead 获取指定 OSS 对象的带预览参数的预签名 URL
func FileRead(objectName string) (ReadUrl string, err error) {

	bucketName := "teachu"
	region := "cn-chengdu"
	endpoint := "https://oss.hedgeho9.cn" // 自定义域名
	// 文档处理参数
	processParams := "doc/preview,export_1,print_1"
	// URL 有效期
	expiration := 10 * time.Hour

	// 调用新封装的函数
	presignedURL, err := GeneratePresignedURL(bucketName, objectName, region, endpoint, processParams, expiration)
	if err != nil {
		// 直接返回错误，让调用者处理
		return "", fmt.Errorf("调用 GeneratePresignedURL 失败: %w", err)
	}

	// 返回获取到的 URL
	return presignedURL, nil
}

// UploadFileToOSS 将本地文件上传到指定的阿里云 OSS 存储空间
func UploadFileToOSS(localFilePath, ossObjectName string) (*oss.PutObjectResult, error) { // 根据 SDK 源码修正返回类型
	region := "cn-chengdu"
	bucketName := "teachu"
	// 加载默认配置并设置凭证提供者和区域
	cfg := oss.LoadDefaultConfig().
		WithCredentialsProvider(credentials.NewEnvironmentVariableCredentialsProvider()).
		WithRegion(region)

	// 创建OSS客户端
	client := oss.NewClient(cfg)

	// 创建上传对象的请求
	putRequest := &oss.PutObjectRequest{
		Bucket:       oss.Ptr(bucketName),      // 存储空间名称
		Key:          oss.Ptr(ossObjectName),   // 对象名称
		StorageClass: oss.StorageClassStandard, // 指定对象的存储类型为标准存储
		Acl:          oss.ObjectACLPublicRead,  // 指定对象的访问权限为公共读
	}

	// 执行上传对象的请求
	result, err := client.PutObjectFromFile(context.TODO(), putRequest, localFilePath)
	if err != nil {
		// 返回错误而不是 Fatalf
		return nil, fmt.Errorf("上传文件 '%s' 到 OSS '%s/%s' 失败: %w", localFilePath, bucketName, ossObjectName, err)
	}

	// 打印上传对象的结果 (可选，如果函数仅用于上传，可以移除日志)
	log.Printf("上传文件 '%s' 到 OSS '%s/%s' 成功. ETag: %s\n", localFilePath, bucketName, ossObjectName, *result.ETag)

	// 返回上传结果和 nil 错误
	return result, nil
}
