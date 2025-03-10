package services

import (
	"context"
	"fmt"
	"io"

	ark "github.com/sashabaranov/go-openai"
)

type StreamChunk struct {
	Data  string
	Error error
}

func GetAIStream(ctx context.Context, message string) <-chan StreamChunk {
	ch := make(chan StreamChunk)
	fmt.Printf("[服务层] 创建流通道，消息: %s\n", message)

	go func() {
		defer func() {
			close(ch)
			fmt.Println("[服务层] 流通道已关闭")
		}()

		// 初始化客户端
		fmt.Println("[服务层] 初始化AI客户端")
		config := ark.DefaultConfig("7c150e05-83d5-4fd4-8dc3-a48e9e154487")
		config.BaseURL = "https://ark.cn-beijing.volces.com/api/v3"
		client := ark.NewClientWithConfig(config)

		// 创建流
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
		if err != nil {
			fmt.Printf("[服务层] 创建流失败: %v\n", err)
			ch <- StreamChunk{Error: err}
			return
		}
		defer stream.Close()
		fmt.Println("[服务层] 流已成功创建")

		// 接收数据
		fmt.Println("[服务层] 开始接收数据块")
		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				fmt.Println("[服务层] 收到流结束信号")
				return
			}
			if err != nil {
				fmt.Printf("[服务层] 接收错误: %v\n", err)
				ch <- StreamChunk{Error: err}
				return
			}

			// 处理有效数据
			if len(resp.Choices) > 0 {
				// 从响应中提取内容
				content := resp.Choices[0].Delta.Content
				fmt.Printf("[服务层] 收到内容: %s\n", content)

				select {
				case <-ctx.Done():
					fmt.Println("[服务层] 上下文已取消")
					return
				case ch <- StreamChunk{Data: content}:
				}
			}
		}
	}()

	return ch
}
