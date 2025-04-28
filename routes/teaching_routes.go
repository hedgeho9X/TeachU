package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/Hedgeho9X/TeachU/internal/controllers"
	"github.com/Hedgeho9X/TeachU/internal/middlewares"
)

// RegisterTeachingRoutes 注册教学管理相关路由
func RegisterTeachingRoutes(r *gin.Engine) {
	r.Use(middlewares.JWTAuth())
	// 添加全局中间件处理 Content-Type
	r.Use(func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		c.Next()
	})

	// 班级路由组
	class := r.Group("/classes")
	{
		class.GET("/list", controllers.ListClasses)
		class.POST("/create", controllers.CreateClass)
		class.DELETE("/delete/:id", controllers.DeleteClass)
	}

	// 学生路由组
	student := r.Group("/students")
	{
		student.GET("/search", controllers.ListStudents)
		student.POST("/batch-import", controllers.ImportStudents)
		student.DELETE("/delete/:student_number", controllers.DeleteStudent)
	}
}
