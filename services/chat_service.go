package services

import (
	"context"
	"fmt"
	"io"
	"os"

	ark "github.com/sashabaranov/go-openai"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime/model"
	"github.com/volcengine/volcengine-go-sdk/volcengine"
)

type StreamChunk struct {
	Data  string
	Error error
}

// GetAIStream 函数接收上下文和消息，返回一个只读的StreamChunk通道
func GetAIStream(ctx context.Context, message string) <-chan StreamChunk {
	// 创建一个StreamChunk类型的通道用于传输数据
	ch := make(chan StreamChunk)
	fmt.Printf("[服务层] 创建流通道，消息: %s\n", message)

	// 启动一个goroutine来处理异步流式数据
	go func() {
		// 确保在函数返回时关闭通道
		defer func() {
			close(ch)
			fmt.Println("[服务层] 流通道已关闭")
		}()

		// 初始化AI客户端配置
		fmt.Println("[服务层] 初始化AI客户端")
		config := ark.DefaultConfig("7c150e05-83d5-4fd4-8dc3-a48e9e154487")
		config.BaseURL = "https://ark.cn-beijing.volces.com/api/v3"
		client := ark.NewClientWithConfig(config)

		// 创建聊天完成流，设置模型和消息内容
		fmt.Println("[服务层] 创建聊天完成流")
		stream, err := client.CreateChatCompletionStream(
			ctx,
			ark.ChatCompletionRequest{
				Model: "ep-20250307105111-pwznn",
				Messages: []ark.ChatCompletionMessage{
					{Role: ark.ChatMessageRoleSystem, Content: SystemPrompt},
					{Role: ark.ChatMessageRoleUser, Content: message},
				},
				Stream: true,
			},
		)
		// 处理流创建错误
		if err != nil {
			fmt.Printf("[服务层] 创建流失败: %v\n", err)
			ch <- StreamChunk{Error: err}
			return
		}
		// 确保在函数返回时关闭流
		defer stream.Close()
		fmt.Println("[服务层] 流已成功创建")

		// 循环接收流数据
		fmt.Println("[服务层] 开始接收数据块")
		for {
			// 从流中接收响应
			resp, err := stream.Recv()
			// 检查是否到达流末尾
			if err == io.EOF {
				fmt.Println("[服务层] 收到流结束信号")
				return
			}
			// 处理接收错误
			if err != nil {
				fmt.Printf("[服务层] 接收错误: %v\n", err)
				ch <- StreamChunk{Error: err}
				return
			}

			// 处理有效的响应数据
			if len(resp.Choices) > 0 {
				// 提取响应中的内容
				content := resp.Choices[0].Delta.Content
				fmt.Printf("[服务层] 收到内容: %s\n", content)
				// 使用select处理数据发送和上下文取消
				select {
				case <-ctx.Done():
					fmt.Println("[服务层] 上下文已取消")
					return
				case ch <- StreamChunk{Data: content}:
					// 成功发送数据到通道
				}
			}
		}
	}()

	// 返回只读通道
	return ch
}

func Chat(text string, Model string, prompt string) (string, error) {
	client := arkruntime.NewClientWithApiKey(
		os.Getenv("ARK_API_KEY"),
		arkruntime.WithBaseUrl("https://ark.cn-beijing.volces.com/api/v3"),
	)

	ctx := context.Background()

	fmt.Println("----- standard request -----")
	req := model.CreateChatCompletionRequest{
		// 指定您创建的方舟推理接入点 ID，此处已帮您修改为您的推理接入点 ID
		// Model: "ep-20250311120726-h7xml",
		Model: Model,
		Messages: []*model.ChatCompletionMessage{
			{
				Role: model.ChatMessageRoleSystem,
				Content: &model.ChatCompletionMessageContent{
					StringValue: volcengine.String(prompt),
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
	return *resp.Choices[0].Message.Content.StringValue, nil
}
