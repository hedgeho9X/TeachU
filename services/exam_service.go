package services

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/Hedgeho9X/TeachU/config"
	"github.com/Hedgeho9X/TeachU/models"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime/model"
	"github.com/volcengine/volcengine-go-sdk/volcengine"
	"gorm.io/gorm"
)

// 解析base64编码
func ImageToBase64(fileBytes []byte) (string, error) {
	// 将字节数据直接编码为 Base64
	base64String := base64.StdEncoding.EncodeToString(fileBytes)
	return base64String, nil
}

// Ai解析文本
func AIAnalyzeText(text string) (string, error) {
	client := arkruntime.NewClientWithApiKey(
		os.Getenv("ARK_API_KEY"),
		arkruntime.WithBaseUrl("https://ark.cn-beijing.volces.com/api/v3"),
	)

	ctx := context.Background()

	fmt.Println("----- AIAnalyzeText request -----")
	req := model.CreateChatCompletionRequest{
		// 指定您创建的方舟推理接入点 ID，此处已帮您修改为您的推理接入点 ID
		Model: "ep-20250311120726-h7xml",
		Messages: []*model.ChatCompletionMessage{
			{
				Role: model.ChatMessageRoleSystem,
				Content: &model.ChatCompletionMessageContent{
					StringValue: volcengine.String("识别试卷 ，给出题号，题号对应的知识点，以及对应的题目内容的json，格式为{\n\"question_id\" :\"\",\n\"key\":\"\",\n\"content\":\" \",\n}-,Json数组格式发送，只保留json内容"),
				},
			},
			{
				Role: model.ChatMessageRoleUser,
				Content: &model.ChatCompletionMessageContent{
					StringValue: volcengine.String(text),
				},
			},
		},
	}

	resp, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		fmt.Printf("standard chat error: %v\n", err)
		return "", err
	}
	string := *resp.Choices[0].Message.Content.StringValue
	fmt.Println(string)
	return *resp.Choices[0].Message.Content.StringValue, nil
}

// AI解析图片
func AIAnalyzePic(base64String string) (string, error) {

	client := arkruntime.NewClientWithApiKey(
		// 从环境变量中获取您的 API Key。此为默认方式，您可根据需要进行修改
		os.Getenv("ARK_API_KEY"),
		// 此为默认路径，您可根据业务所在地域进行配置
		arkruntime.WithBaseUrl("https://ark.cn-beijing.volces.com/api/v3"),
	)
	ctx := context.Background()
	req := model.CreateChatCompletionRequest{
		// 指定您创建的方舟推理接入点 ID，此处已帮您修改为您的推理接入点 ID
		Model: "ep-20250323114418-qv2jv",
		Messages: []*model.ChatCompletionMessage{
			{
				Role: model.ChatMessageRoleUser,
				Content: &model.ChatCompletionMessageContent{
					ListValue: []*model.ChatCompletionMessageContentPart{
						{
							Type: model.ChatCompletionMessageContentPartTypeText,
							Text: AnalyzePrompt,
						},
						{
							Type: model.ChatCompletionMessageContentPartTypeImageURL,
							ImageURL: &model.ChatMessageImageURL{
								URL: fmt.Sprintf("data:image/png;base64,%s", base64String),
							},
						},
					},
				},
			},
		},
	}

	resp, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		fmt.Printf("standard chat error: %v\n", err)
		return "", err
	}
	// fmt.Println(*resp.Choices[0].Message.Content.StringValue)
	return *resp.Choices[0].Message.Content.StringValue, nil

}

// 数据库存储Exam
func CreateExam(exam *models.Exam) error {
	// 检查班级是否存在
	var class models.Class
	if err := config.DB.First(&class, exam.ClassId).Error; err != nil {
		return errors.New("班级不存在")
	}
	// 检查用户是否为班级创建者
	if exam.UserId != class.CreatedUserID {
		return errors.New("您不是班级创建者，无法创建试题")
	}
	return config.DB.Create(exam).Error
}

// DeleteExam 级联删除考试及相关数据
func DeleteExam(examId string, userID uint) error {
	var exam models.Exam

	// 验证考试存在性和权限
	if err := config.DB.First(&exam, examId).Error; err != nil {
		return errors.New("试题不存在")
	}
	if exam.UserId != userID {
		return errors.New("无权删除该试题")
	}

	// 直接删除考试记录，依赖数据库级联删除
	if err := config.DB.Unscoped().Delete(&exam).Error; err != nil {
		return fmt.Errorf("删除失败: %v", err)
	}
	return nil
}

func GetExamByID(db *gorm.DB, id string) (*models.Exam, error) {
	var exam models.Exam
	result := db.First(&exam, "id = ?", id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &exam, nil
}

func ListExamByClassID(classID uint, userID uint) ([]models.Exam, error) {
	var exams []models.Exam
	// 检查班级是否存在
	var class models.Class
	if err := config.DB.First(&class, classID).Error; err != nil {
		return nil, errors.New("班级不存在")
	}
	// 检查用户是否为班级创建者
	if userID != class.CreatedUserID {
		return nil, errors.New("您不是班级创建者，无法查看考试")
	}
	// 查询班级下的所有考试
	result := config.DB.Where("class_id =?", classID).Find(&exams)
	if result.Error != nil {
		return nil, result.Error
	}
	return exams, nil
}

// 结构体定义
type Question struct {
	QuestionID string `json:"question_id"`
	Key        string `json:"key"`
	Content    string `json:"content"`
}

// ParseQuestions 解析试题JSON到结构体切片
func ParseQuestions(jsonStr string) ([]Question, error) {
	var questions []Question
	if err := json.Unmarshal([]byte(jsonStr), &questions); err != nil {
		return nil, fmt.Errorf("JSON解析失败: %w", err)
	}
	return questions, nil
}
