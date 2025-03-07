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
		c.JSON(http.StatusOK, gin.H{"error": err.Error()})
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
