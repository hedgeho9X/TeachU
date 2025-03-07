package controllers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Hedgeho9X/TeachU/services"
	"github.com/gin-gonic/gin"
)

func Chat(c *gin.Context) {
	// 绑定请求参数
	var input struct {
		UserMessage string `json:"message"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 设置流式响应头
	c.Writer.Header().Set("Content-Type", "application/json; charset=utf-8")
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

		// 在流式写入部分添加JSON格式化
		case chunk, ok := <-stream:
			if !ok {
				// 发送流结束标记
				c.Writer.Write([]byte("data: {\"type\":\"done\"}\n\n"))
				c.Writer.Flush()
				return
			}
			if chunk.Error != nil {
				// 格式化错误信息为JSON
				errJson := fmt.Sprintf(`{"type":"error","message":%q}`, chunk.Error.Error())
				c.Writer.Write([]byte("data: " + errJson + "\n\n"))
				c.Writer.Flush()
				return
			}

			// 格式化正常数据为JSON
			jsonData := fmt.Sprintf(`{"type":"data","content":%q}`, chunk.Data)
			if _, err := c.Writer.Write([]byte("data: " + jsonData + "\n\n")); err != nil {
				fmt.Println("写入失败:", err)
				return
			}
			c.Writer.Flush()
		}
	}
}
