package controllers

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/Hedgeho9X/TeachU/config"
	"github.com/Hedgeho9X/TeachU/services"
	"github.com/gin-gonic/gin"
)

func Chat(c *gin.Context) {
	// 绑定请求参数
	var input struct {
		UserMessage string `json:"message"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 0, "error": err.Error()})
		return
	}

	// 设置流式响应头
	c.Writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Flush()

	// 创建可取消上下文
	ctx, cancel := context.WithCancel(c)
	defer cancel()

	// 监听客户端断开
	clientGone := c.Writer.CloseNotify()

	// 获取AI响应流
	stream := services.GetAIStream(ctx, input.UserMessage)

	// 流式写入响应
	for {
		select {
		case <-clientGone:
			fmt.Println("客户端断开连接")
			return

		case chunk, ok := <-stream:
			if !ok {
				c.Writer.Flush()
				return
			}
			if chunk.Error != nil {
				c.Writer.Write([]byte(chunk.Error.Error()))
				c.Writer.Flush()
				return
			}

			if _, err := c.Writer.Write([]byte(chunk.Data)); err != nil {
				fmt.Println("写入失败:", err)
				return
			}
			c.Writer.Flush()
		}
	}
}

func PptGenerate(c *gin.Context) {
	var input struct {
		Query      string `json:"query" `
		TemplateID string `json:"template_id"`
		Author     string `json:"author"`
		Search     bool   `json:"search"`
		Language   string `json:"language"`
		IsFigure   bool   `json:"is_figure"`
		AIImage    string `json:"ai_image"`
	}
	//解析json参数，注意AIImage仅在IsFigure为true时才有用

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"msg":  "参数错误",
			"data": err.Error(),
		})
		return
	}

	// 业务逻辑验证
	if input.Query == "" {
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"msg":  "用户提示不能为空",
		})
		return
	}

	if input.IsFigure && input.AIImage == "" {
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"msg":  "启用图片生成时必须指定AIImage类型",
		})
		return
	}

	auth := &services.AuthConfig{
		AppID:  config.XunFei_ID,
		Secret: config.XunFei_Secret,
	}

	// // 示例1：查询PPT主题列表
	// themeReq := services.PPTThemeRequest{
	// 	Style:    "商务",
	// 	PageSize: 10,
	// }
	// themeRes, _ := services.GetPPTThemeList(auth, themeReq)
	// fmt.Printf("主题列表响应: %+v\n", themeRes)

	// 生成PPT
	pptReq := services.GeneratePPTRequest{
		Query:      "用于教学，要求:" + input.Query,
		TemplateID: input.TemplateID,
		Author:     input.Author,
		Search:     input.Search,
		Language:   input.Language,
		IsFigure:   input.IsFigure,
		AIImage:    input.AIImage,
		IsCardNote: true,
	}
	pptRes, _ := services.GeneratePPT(auth, pptReq)
	c.JSON(http.StatusOK, gin.H{
		"code": 1,
		"msg":  "PPT生成响应",
		"data": pptRes,
	})
	// fmt.Printf("PPT生成响应: %+v\n", pptRes)
}

func GetPPTProgress(c *gin.Context) {
	auth := &services.AuthConfig{
		AppID:  config.XunFei_ID,
		Secret: config.XunFei_Secret,
	}
	//获取并解析query参数
	sid := c.Query("sid")
	if sid == "" {
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"msg":  "sid不能为空",
		})
		return
	}
	//获取ppt生成进度
	progressRes, err := services.GetPPTProgress(auth, sid)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"msg":  "查询生成进度失败",
			"data": err.Error(),
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 1,
		"msg":  "查询生成进度成功",
		"data": progressRes,
	})
}

func GetPPtTemplates(c *gin.Context) {
	// 从query参数中获取输入并进行URL解码
	style := c.Query("style")
	color := c.Query("color")
	industry := c.Query("industry")
	pageNum := c.Query("page_num")
	pageSize := c.Query("page_size")

	// 去除空格
	// style = strings.TrimSpace(style)
	// color = strings.TrimSpace(color)
	// industry = strings.TrimSpace(industry)
	// pageNum = strings.TrimSpace(pageNum)
	// pageSize = strings.TrimSpace(pageSize)
	// fmt.Print(style, color, industry, pageNum, pageSize)
	// // 从query参数中获取输入
	// style := c.Query("style")        // 风格类型
	// color := c.Query("color")        // 颜色类型
	// industry := c.Query("industry")  // 行业类型
	// pageNum := c.Query("page_num")   // 页数
	// pageSize := c.Query("page_size") // 每页数量

	// 转换页码和每页数量为整数
	pageNumInt := 1
	if num, err := strconv.Atoi(pageNum); err == nil && num > 0 {
		pageNumInt = num
	}

	pageSizeInt := 10
	if size, err := strconv.Atoi(pageSize); err == nil && size > 0 {
		pageSizeInt = size
	}
	fmt.Printf("解码后参数: style=[%s], color=[%s], industry=[%s], page_num=[%s], page_size=[%s]\n",
		style, color, industry, pageNum, pageSize)
	auth := &services.AuthConfig{
		AppID:  config.XunFei_ID,
		Secret: config.XunFei_Secret,
	}

	// 验证风格参数
	validStyles := map[string]bool{
		"简约": true, "卡通": true, "商务": true, "创意": true,
		"国风": true, "清新": true, "扁平": true, "插画": true, "节日": true,
	}
	if style != "" && !validStyles[style] {
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"msg":  "无效的风格类型",
		})
		return
	}

	// 验证颜色参数
	validColors := map[string]bool{
		"蓝色": true, "绿色": true, "红色": true, "紫色": true, "黑色": true,
		"灰色": true, "黄色": true, "粉色": true, "橙色": true,
	}
	if color != "" && !validColors[color] {
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"msg":  "无效的颜色类型",
		})
		return
	}

	// 验证行业参数
	validIndustries := map[string]bool{
		"教育培训": true, "学院": true}
	if industry != "" && !validIndustries[industry] {
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"msg":  "无效的行业类型",
		})
		return
	}

	// 查询PPT主题列表
	themeReq := services.PPTThemeRequest{
		Style:    style,
		Color:    color,
		Industry: industry,
		PageNum:  pageNumInt,
		PageSize: pageSizeInt,
	}

	themeRes, err := services.GetPPTThemeList(auth, themeReq)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"msg":  "获取模板列表失败",
			"data": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 1,
		"msg":  "获取模板列表成功",
		"data": themeRes,
	})
}

func PicGenerate(c *gin.Context) {
	var input struct {
		Prompt string `json:"prompt"`
		Width  int    `json:"width"`
		Height int    `json:"height"`
		Seed   int    `json:"seed"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"msg":  "参数错误",
			"data": err.Error(),
		})
		return
	}
	if input.Prompt == "" {
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"msg":  "提示词不能为空",
		})
		return
	}

	// 调用图像生成服务
	pics, err := services.PicGenerate(services.PicGenerateRequest{
		Prompt: input.Prompt,
		Width:  input.Width,
		Height: input.Height,
		Seed:   input.Seed,
	})
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"msg":  "图像生成失败",
			"data": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 1,
		"msg":  "图像生成成功",
		"data": pics,
	})
}

// 返回随机的6个主题
func GetRecommmend(c *gin.Context) {
	// 将所有主题放入一个切片
	var allTopics []string
	for _, topics := range services.Subjects {
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
	c.JSON(http.StatusOK, gin.H{
		"code": 1,
		"msg":  "获取推荐主题成功",
		"data": result,
	})
}
