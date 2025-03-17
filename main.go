package main

import (
	"encoding/csv"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.POST("/upload", func(c *gin.Context) {
		// 获取上传的文件
		file, header, err := c.Request.FormFile("file")
		if err != nil {
			c.String(http.StatusBadRequest, fmt.Sprintf("文件上传失败: %s", err.Error()))
			return
		}
		defer file.Close()

		// 打印文件信息
		fmt.Printf("上传的文件名: %s\n", header.Filename)

		// 读取 CSV 文件内容
		reader := csv.NewReader(file)
		records, err := reader.ReadAll()
		if err != nil {
			c.String(http.StatusInternalServerError, fmt.Sprintf("读取 CSV 文件失败: %s", err.Error()))
			return
		}

		// 打印 CSV 文件内容
		for _, record := range records {
			fmt.Println(record)
		}

		c.String(http.StatusOK, "文件上传并解析成功")
	})

	r.Run(":8080")
}
