package main

import (
	"context"
	"fmt"
	"os"

	"github.com/volcengine/volcengine-go-sdk/service/arkruntime"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime/model"
	"github.com/volcengine/volcengine-go-sdk/volcengine"
)

func main() {
	// 请确保您已将 API Key 存储在环境变量 ARK_API_KEY 中
	// 初始化Ark客户端，从环境变量中读取您的API Key
	client := arkruntime.NewClientWithApiKey(
		// 从环境变量中获取您的 API Key。此为默认方式，您可根据需要进行修改
		os.Getenv("ARK_API_KEY"),
		// 此为默认路径，您可根据业务所在地域进行配置
		arkruntime.WithBaseUrl("https://ark.cn-beijing.volces.com/api/v3"),
	)

	ctx := context.Background()

	fmt.Println("----- standard request -----")
	req := model.CreateChatCompletionRequest{
		// 指定您创建的方舟推理接入点 ID，此处已帮您修改为您的推理接入点 ID
		Model: "ep-20250311120726-h7xml",
		Messages: []*model.ChatCompletionMessage{
			{
				Role: model.ChatMessageRoleSystem,
				Content: &model.ChatCompletionMessageContent{
					StringValue: volcengine.String("识别试卷 ，给出题号，题号对应的知识点，以及对应的题目内容的json，格式为{\n\"question_id\" :\"\",\n\"key\":\"\",\n\"chapter\":\"\"\n\"content\":\" \",\n}-,Json数组格式发送，只保留json内容"),
				},
			},
			{
				Role: model.ChatMessageRoleUser,
				Content: &model.ChatCompletionMessageContent{
					StringValue: volcengine.String("【例1】（23-24高二下·安徽合肥·期末）若质点A运动的位移S（单位：\"m\" ）与时间t（单位：\"s\" ）之间的函数关系是S(t)=-2/t(t≥1），那么该质点在t= 3\"s\" 时的瞬时速度和从t=1\"s\" 到t=3\"s\" 这两秒内的平均速度分别为（    ）\nA．-2/3,2/9\tB．2/3,2/9\tC．2/9,-2/3\tD．2/9,2/3\n【例2】（23-24高二下·福建龙岩·期中）若函数f(x)=x^2-x，则函数f(x)从x=1到x=3的平均变化率为（    ）\nA．6\tB．3\tC．2\tD．1\n"),
				},
			},
		},
	}

	resp, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		fmt.Printf("standard chat error: %v\n", err)
		return
	}
	fmt.Println(*resp.Choices[0].Message.Content.StringValue)

	//fmt.Println("----- streaming request -----")
	//
	//req = model.CreateChatCompletionRequest{
	//	// 指定您创建的方舟推理接入点 ID，此处已帮您修改为您的推理接入点 ID
	//	Model: "ep-20250311120726-h7xml",
	//	Messages: []*model.ChatCompletionMessage{
	//		{
	//			Role: model.ChatMessageRoleSystem,
	//			Content: &model.ChatCompletionMessageContent{
	//				StringValue: volcengine.String("识别试卷 ，给出知识点，题号，题目内容的json."),
	//			},
	//		},
	//		{
	//			Role: model.ChatMessageRoleUser,
	//			Content: &model.ChatCompletionMessageContent{
	//				StringValue: volcengine.String("1．数列的概念\n数列的定义\n一般地，把按照确定的顺序排列的一列数称为数列.数列中的每一个数叫做这个数列的项，数列的第一\n个位置上的数叫做这个数列的第1项，常用符号 表示，第二个位置上的数叫做这个数列的第2项，用 表示  第n个位置上的数叫做这个数列的第n项，用 表示.其中第1项也叫做首项. \n1．数列的概念\n数列的定义\n一般地，把按照确定的顺序排列的一列数称为数列.数列中的每一个数叫做这个数列的项，数列的第一\n个位置上的数叫做这个数列的第1项，常用符号 表示，第二个位置上的数叫做这个数列的第2项，用 表示  第n个位置上的数叫做这个数列的第n项，用 表示.其中第1项也叫做首项. \n？"),
	//			},
	//		},
	//	},
	//}
	stream, err := client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		fmt.Printf("stream chat error: %v\n", err)
		return
	}
	defer stream.Close()

	//for {
	//	recv, err := stream.Recv()
	//	if err == io.EOF {
	//		return
	//	}
	//	if err != nil {
	//		fmt.Printf("Stream chat error: %v\n", err)
	//		return
	//	}
	//
	//	if len(recv.Choices) > 0 {
	//		fmt.Print(recv.Choices[0].Delta.Content)
	//	}
	//}
}
