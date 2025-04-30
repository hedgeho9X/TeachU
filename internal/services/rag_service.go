package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Hedgeho9X/TeachU/internal/models"
)

// StuAnalysis2String将学生学情json转换为string
func StuAnalysis2String(examID uint, studentID uint) (string, error) {
	analysis, err := AnalyzeStudent(examID, studentID)
	if err != nil {
		return "", err
	}

	JsonRes := struct {
		AnalysisResult models.StudentAnalysisResponse `json:"analysis"`
	}{
		AnalysisResult: analysis,
	}

	// 转换为JSON字符串
	jsonData, err := json.Marshal(JsonRes)
	if err != nil {
		return "", fmt.Errorf("JSON序列化失败: %v", jsonData)
	}
	return string(jsonData), nil
}

func RagRecommend(prompt string) (map[string]interface{}, error) {
	// 1. 调用RAG服务
	reqBody := fmt.Sprintf(`{"question": "%s"}`, prompt)

	// 2. 创建HTTP请求->server2
	req, err := http.NewRequest(
		"POST",
		"http://118.145.201.37:8000/recommend",
		bytes.NewBufferString(reqBody),
	)
	if err != nil {
		return nil, fmt.Errorf("请求创建失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// 3. 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("服务请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 4. 处理响应
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("推荐服务返回错误状态码: %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("响应解析失败: %w", err)
	}

	return result, nil
}
