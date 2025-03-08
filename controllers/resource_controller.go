package controllers

import (
	"net/http"

	"github.com/Hedgeho9X/TeachU/services"
	"github.com/gin-gonic/gin"
)

func SearchResources(c *gin.Context) {
	// 从URL查询参数获取值
	pageNum := c.Query("page_num")
	pageSize := c.Query("page_size")
	q := c.Query("q")
	subject := c.Query("subject")
	grade := c.Query("grade")

	if q == "" && subject == "" && grade == "" {
		c.JSON(http.StatusOK, gin.H{"code": 0, "error": "查找内容不能为空"})
		return
	}
	// 调用 service 层执行查找
	resources, err := services.SearchResources(pageNum, pageSize, q, subject, grade)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 0, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 1,
		"data": resources,
	})
}
