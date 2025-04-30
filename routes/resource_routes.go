package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/Hedgeho9X/TeachU/internal/controllers"
)

// RegisterResourceRoutes 注册资源相关路由
func RegisterResourceRoutes(r *gin.Engine) {
	// 资源路由组
	resource := r.Group("/resource")
	resource.GET("/list", controllers.SearchResources)
	resource.GET("/object", controllers.GetResource)

}
