package services

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"os"
	"time"

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

var subjects = map[string][]string{
	"物理": {
		"密度公式应用宝典：一杯水的质量怎么算？", // 原知识点升级
		"浮力原理实验室：为什么铁船能浮在水面？",
		"压强公式实战：高跟鞋vs平底鞋谁更伤地板？", // 生活场景类比
		"光的折射魔法秀：筷子弯折的奥秘在这里！",
		"热传递三剑客：炒菜时三种传热方式都在工作", // 增加具象化描述
		"电路检修指南：如何用电压表快速定位故障？", // 新增知识点+
		"焦耳定律实验课：电热丝为什么会发红？",
		"电磁感应现象show：手摇发电机的秘密",
		"能量转化大追踪：过山车中的动能势能转换", // 增加实际案例
		"声音特性探索：为什么不同乐器音色不同？", // 新增知识点+
		"杠杆平衡特训：撬地球需要多长的棍子？",
		"惯性现象大发现：急刹车时身体前倾的真相", // 新增知识点+
	},
	"化学": {
		"元素周期表探秘：同族元素为什么性质相似？",
		"原子结构拆解：电子层就像洋葱的衣裳", // 生活化比喻
		"化合价记忆口诀升级版：一价钾钠氯氢银...",
		"实验室安全守则：这些标志你一定要认识！", // 新增图示说明+
		"过滤操作step by step：一贴二低三靠",
		"气体收集妙招：排水法vs向上排空气法",
		"燃烧条件实验：为什么纸锅烧水不会破？", // 趣味实验
		"碳的魔法世界：金刚石和石墨竟是亲戚！",
		"化肥鉴别技巧：氮磷钾肥怎么区分？", // 新增实用技能+
		"指示剂变色秀：紫甘蓝汁的酸碱彩虹", // 增加趣味性
		"质量守恒验证：镁条燃烧后的神秘增重",
		"溶液配制实操：浓度计算的万能公式", // 新增知识点+
	},
	"生物": {
		"细胞分裂live秀：有丝分裂全过程动画演示", // 增加动态描述
		"植物器官功能展：根茎叶的才艺大比拼",
		"消化系统漫游记：汉堡的24小时旅程", // 故事化叙述
		"遗传规律实战：为什么双眼皮爸妈可能生出单眼皮宝宝？",
		"生态沙盘模拟：如果草原上没有狼会怎样？", // 情景假设
		"呼吸作用实验室：澄清石灰水变浑浊的秘密",
		"免疫防线大阅兵：白细胞军团作战实录", // 拟人化手法
		"生物分类闯关：你能找到大熊猫的族谱吗？",
		"变异类型案例：太空椒的诞生记", // 新增科技关联+
		"微生物应用展：酵母菌的十八般武艺",
		"湿地保护行动：地球之肾的生态价值",   // 新增环保主题+
		"显微镜使用秘籍：从对光到找气泡的诀窍", // 新增实验技能+
	},
	// 历史、地理、道法继续向下展开...
	"历史": {
		"秦朝时空穿越：如果你是秦始皇会怎么做？", // 角色扮演式设问
		"丝路商队体验：从长安到罗马要带什么货物？",
		"五四青年日志：还原1919年的热血一天", // 沉浸式场景
		"开国大典直播：1949年的历史性时刻",
		"冷战剧场：美苏争霸的经典名场面",
		"法国大革命剧本杀：攻占巴士底狱的真相",    // 结合流行形式
		"郑和宝船探秘：古代航海黑科技有哪些？",    // 新增知识点+
		"甲午沉思录：黄海海战的启示与教训",      // 新增知识点+
		"经济全球化实验：你的早餐包含多少国家元素？", // 生活化切入
		"抗战文物说：从老照片看民族记忆",       // 新增实物关联+
		"古埃及探奇：金字塔建造的未解之谜",      // 新增世界史内容+
	},
	"地理": {
		"经纬网寻宝游戏：北纬30°的神秘地点", // 游戏化设计
		"等高线解密：3D地形与平面图的转换魔法",
		"中国地形探险：沿着胡焕庸线走一遍",        // 新增地理分界线+
		"气候类型cosplay：给不同气候配专属BGM", // 创意记忆法
		"人口问题辩论：人多好还是人少好？",
		"资源危机模拟：如果石油枯竭怎么办？",  // 情景假设
		"地球自转实验：用傅科摆证明地转偏向力", // 新增实验验证+
		"地震逃生演练：黄金12秒该怎么做？",  // 新增生存技能+
		"城市病诊断书：堵车雾霾怎么治？",
		"区域地理PK赛：秦岭淮河两岸大不同",
		"火山探秘直播：地幔物质的地下旅行",  // 新增自然现象+
		"二十四节气研学：古人如何观测太阳？", // 新增传统文化+
	},
	"道德与法治": {
		"法律诊所：帮小红分析网络谣言传播案例", // 案例教学
		"罪案现场分析：这个行为构成犯罪吗？",
		"价值观践行日记：我的24字修炼手册",
		"民族团结拼图：五十六个民族服饰展", // 可视化元素
		"消费维权剧场：遇到假货该怎么处理？",
		"道德两难选择：扶不扶老人的思辨课",  // 新增热点议题+
		"外交风云录：最近的国际大事怎么看？", // 关联时事
		"改革记忆馆：老物件里的时代变迁",
		"教育公平调研：城乡学校设施对比", // 新增实践项目+
		"网络防身术：十条必备的安全守则",
		"文化自信之旅：从故宫文创到汉服热",
		"人类命运共同体实践：全球抗疫中的中国担当", // 关联现实
	},
}

// GetRandomTopics 返回随机的6个主题
func GetRandomTopics() []string {
	// 将所有主题放入一个切片
	var allTopics []string
	for _, topics := range subjects {
		allTopics = append(allTopics, topics...)
	}

	// 打乱顺序
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(allTopics), func(i, j int) {
		allTopics[i], allTopics[j] = allTopics[j], allTopics[i]
	})

	// 取前6个
	var result []string
	if len(allTopics) >= 6 {
		result = allTopics[:6]
	} else {
		result = allTopics
	}

	return result
}
