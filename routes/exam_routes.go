package routes

import (
	"github.com/Hedgeho9X/TeachU/internal/controllers"
	"github.com/Hedgeho9X/TeachU/internal/middlewares"
	"github.com/gin-gonic/gin"
)

// RegisterExamRoutes 注册考试和成绩相关的路由
func RegisterExamRoutes(r *gin.Engine) {
	r.Use(func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		c.Next()
	})

	// 考试路由组
	exam := r.Group("/exam")
	// 对考试路由组应用 JWT 认证中间件
	exam.Use(middlewares.JWTAuth())
	// 创建考试路由
	exam.POST("/create", controllers.CreateExam)
	// 删除考试路由
	exam.DELETE("/:id", controllers.DeleteExam)
	// 获取考试列表路由
	exam.GET("/list", controllers.ListExam)

	// 成绩路由组
	score := r.Group("/exam/score")
	// 上传成绩路由
	score.POST("/upload", controllers.UploadScore)
	// 获取成绩列表路由
	score.GET("/list", controllers.ListScore)
}
