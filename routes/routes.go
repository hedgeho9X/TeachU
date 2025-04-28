package routes

import (
	"github.com/gin-gonic/gin"
)

// SetupRouter 初始化并配置 Gin 引擎，设置所有应用程序路由
func SetupRouter() *gin.Engine {
	r := gin.Default()

	// 基础路由
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Welcome to TeachU ",
		})
	})

	// 注册各模块路由
	// 注册认证相关路由 (登录、注册等)
	RegisterAuthRoutes(r)
	// 注册考试管理相关路由
	RegisterExamRoutes(r)
	// 注册题目管理相关路由
	RegisterProblemsRoutes(r)
	// 注册 AI 相关路由 (AI 分析等)
	RegisterAIRoutes(r)
	// 注册资源管理相关路由
	RegisterResourceRoutes(r)
	// 注册教学管理相关路由 (班级、学生管理等)
	RegisterTeachingRoutes(r)
	// 注册数据分析相关路由
	RegisterAnalysisRoutes(r)
	return r
}
