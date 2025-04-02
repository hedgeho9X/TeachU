package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/Hedgeho9X/TeachU/controllers"
	"github.com/Hedgeho9X/TeachU/middlewares"
)

// RegisterAuthRoutes 注册认证相关路由
func RegisterAnalysisRoutes(r *gin.Engine) {
	// 创建 auth 组
	analysis := r.Group("/analysis")
	analysis.Use(middlewares.JWTAuth())
	analysis.GET("/class", controllers.AnalyzeExam)
	analysis.GET("/class-ai", controllers.AiAnalyzeClass)
	analysis.GET("/student", controllers.AnalyzeStudent)
	analysis.GET("/student-ai", controllers.AiAnalyzeStudent)
	analysis.GET("/rank", controllers.GetRank)
}
