package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func RagRecommend(examID uint, studentID uint, prompt string) (map[string]interface{}, error) {

	// analysis, err := AnalyzeStudent(examID, studentID)
	// if err != nil {
	// 	return nil, err
	// }

	// JsonRes := struct {
	// 	AnalysisResult models.StudentAnalysisResponse `json:"analysis"`
	// }{
	// 	AnalysisResult: analysis,
	// }

	// // 转换为JSON字符串
	// jsonData, err := json.Marshal(JsonRes)
	// if err != nil {
	// 	return nil, fmt.Errorf("JSON序列化失败: %v", jsonData)
	// }

	// question, err := Chat("", DoubaoLite, "推荐一些历史试题")
	// if err != nil {
	// 	return nil, err
	// }
	// println(question)
	// 2. 创建请求体
	reqBody := fmt.Sprintf(`{"question": "%s"}`, prompt)

	// 3. 创建HTTP请求
	req, err := http.NewRequest(
		"POST",
		"http://118.145.201.37:8000/recommend",
		bytes.NewBufferString(reqBody),
	)
	if err != nil {
		return nil, fmt.Errorf("请求创建失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// 4. 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("服务请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 5. 处理响应
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("推荐服务返回错误状态码: %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("响应解析失败: %w", err)
	}

	return result, nil
}
