package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/Hedgeho9X/TeachU/controllers"
	"github.com/Hedgeho9X/TeachU/middlewares"
)

// RegisterTeachingRoutes 注册教学管理相关路由
func RegisterTeachingRoutes(r *gin.Engine) {
	// 班级路由组
	class := r.Group("/classes")
	class.Use(middlewares.JWTAuth())
	class.Use(middlewares.JWTAuth())
	{
		class.GET("/list", controllers.ListClasses)    // 获取班级列表
		class.POST("/create", controllers.CreateClass) // 创建班级
		// class.PUT("/update/:id", controllers.UpdateClass)    // 更新班级信息
		class.DELETE("/delete/:id", controllers.DeleteClass) // 删除班级
	}

	// 学生路由组
	student := r.Group("/students")
	student.Use(middlewares.JWTAuth())
	// {
	student.GET("/list/", controllers.ListStudents)  // 按班级查询学生
	student.POST("/upload", controllers.UploadFiles) //空接口接收前端文件
	// 	student.POST("/create", controllers.CreateStudent)        // 添加学生
	student.POST("/batch-import", controllers.ImportStudents) // 批量导入
	// 	student.PUT("/update/:id", controllers.UpdateStudent)     // 更新学生信息
	// 	student.DELETE("/delete/:id", controllers.DeleteStudent)  // 删除学生
	// }
	//接收文件
}
