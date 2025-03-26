package hhh

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/volcengine/volcengine-go-sdk/service/arkruntime"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime/model"
	"github.com/volcengine/volcengine-go-sdk/volcengine"
)

// main 函数是程序的入口点
func 1() {
	// 替换为您的推理接入点ID（https://www.volcengine.com/docs/82379/1099522）
	const Model = "ep-20250311120726-h7xml"
	// 使用 API Key创建一个客户端，从环境变量中获取 API Key（https://www.volcengine.com/docs/82379/1361424）
	client := arkruntime.NewClientWithApiKey(
		os.Getenv("ARK_API_KEY"),
	)
	// 创建一个新的上下文
	goCtx := context.Background()

	// 打印创建上下文的提示信息
	fmt.Println("----- create context -----")
	// 创建一个新的上下文请求
	createCtxReq := model.CreateContextRequest{
		// 设置模型为常量 Model
		Model: Model,
		// 设置模式为会话模式
		Mode: model.ContextModeSession,
		// 设置初始消息
		Messages: []*model.ChatCompletionMessage{
			{
				// 设置角色为系统
				Role: model.ChatMessageRoleSystem,
				// 设置内容为系统消息
				Content: &model.ChatCompletionMessageContent{
					// 设置字符串值为系统消息内容
					StringValue: volcengine.String("你是李雷"),
				},
			},
		},
		// 设置 TTL 为 3600 秒
		TTL: volcengine.Int(3600),
	}

	// 发送创建上下文请求并获取响应
	createCtxRsp, err := client.CreateContext(goCtx, createCtxReq)
	// 如果发生错误，打印错误信息并返回
	if err != nil {
		fmt.Printf("create context error: %v\n", err)
		return
	}
	// 打印创建上下文的响应
	fmt.Printf("create context response: %v\n", createCtxRsp)

	// 打印非流式聊天的提示信息
	fmt.Println("----- chat round 1 (non-stream) -----")
	// 创建一个新的聊天请求
	req := model.ContextChatCompletionRequest{
		// 设置上下文ID为创建上下文的响应ID
		ContextID: createCtxRsp.ID,
		// 设置模型为常量 Model
		Model: Model,
		// 设置消息
		Messages: []*model.ChatCompletionMessage{
			{
				// 设置角色为用户
				Role: model.ChatMessageRoleUser,
				// 设置内容为用户消息
				Content: &model.ChatCompletionMessageContent{
					// 设置字符串值为用户消息内容
					StringValue: volcengine.String("我的名字是方方"),
				},
			},
		},
	}

	// 发送聊天请求并获取响应
	resp, err := client.CreateContextChatCompletion(goCtx, req)
	// 如果发生错误，打印错误信息并返回
	if err != nil {
		fmt.Printf("non-stream chat error: %v\n", err)
		return
	}
	// 打印聊天响应的内容
	fmt.Println(*resp.Choices[0].Message.Content.StringValue)

	// 打印流式聊天的提示信息
	fmt.Println("----- chat round 2 (stream) -----")
	// 创建一个新的聊天请求
	req = model.ContextChatCompletionRequest{
		// 设置上下文ID为创建上下文的响应ID
		ContextID: createCtxRsp.ID,
		// 设置模型为常量 Model
		Model: Model,
		// 设置消息
		Messages: []*model.ChatCompletionMessage{
			{
				// 设置角色为用户
				Role: model.ChatMessageRoleUser,
				// 设置内容为用户消息
				Content: &model.ChatCompletionMessageContent{
					// 设置字符串值为用户消息内容
					StringValue: volcengine.String("你是谁，我是谁？"),
				},
			},
		},
	}
	// 发送聊天请求并获取流式响应
	stream, err := client.CreateContextChatCompletionStream(goCtx, req)
	// 如果发生错误，打印错误信息并返回
	if err != nil {
		fmt.Printf("stream chat error: %v\n", err)
		return
	}
	// 延迟关闭流式响应
	defer stream.Close()

	// 循环接收流式响应
	for {
		// 接收流式响应
		recv, err := stream.Recv()
		// 如果接收到 EOF，返回
		if err == io.EOF {
			return
		}
		// 如果发生错误，打印错误信息并返回
		if err != nil {
			fmt.Printf("Stream chat error: %v\n", err)
			return
		}
		// 如果接收到的响应中有选择项
		if len(recv.Choices) > 0 {
			// 打印选择项的内容
			fmt.Print(recv.Choices[0].Delta.Content)
		}
	}
}
