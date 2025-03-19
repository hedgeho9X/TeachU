package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/Hedgeho9X/TeachU/controllers"
	"github.com/Hedgeho9X/TeachU/middlewares"
)

// RegisterTeachingRoutes 注册教学管理相关路由
func RegisterTeachingRoutes(r *gin.Engine) {
	// 添加全局中间件处理 Content-Type
	r.Use(func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		c.Next()
	})

	// 班级路由组
	class := r.Group("/classes")
	class.Use(middlewares.JWTAuth()) // 移除重复的 JWTAuth
	{
		class.GET("/list", controllers.ListClasses)
		class.POST("/create", controllers.CreateClass)
		class.DELETE("/delete/:id", controllers.DeleteClass)
	}

	// 学生路由组
	student := r.Group("/students")
	student.Use(middlewares.JWTAuth())
	{ // 恢复大括号
		student.GET("/search", controllers.ListStudents)
		student.POST("/batch-import", controllers.ImportStudents)
		student.DELETE("/delete/:student_number", controllers.DeleteStudent)
	}
}
