package routes

import (
	"github.com/gin-gonic/gin"
)

// SetupRouter 设置所有路由
func SetupRouter() *gin.Engine {
	r := gin.Default()

	// 基础路由
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Welcome to TeachU API",
		})
	})

	// 注册各模块路由
	RegisterAuthRoutes(r)
	RegisterExamRoutes(r)
	RegisterProblemsRoutes(r)
	RegisterAIRoutes(r)
	RegisterResourceRoutes(r)
	RegisterTeachingRoutes(r)

	return r
}
