package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/Hedgeho9X/TeachU/internal/controllers"
	"github.com/Hedgeho9X/TeachU/internal/middlewares"
)

// RegisterAIRoutes 注册AI功能相关路由
func RegisterAIRoutes(r *gin.Engine) {

	r.GET("/ai/recommend", controllers.GetRecommmend)
	// AI路由组 - 所有接口需要JWT认证
	ai := r.Group("/ai")
	ai.Use(middlewares.JWTAuth()) // JWT身份验证

	// 聊天相关接口
	ai.POST("/chat", controllers.Chat)

	// PPT生成相关接口
	ppt := ai.Group("/ppt")
	ppt.POST("/generate", controllers.PptGenerate)
	ppt.GET("/progress", controllers.GetPPTProgress)
	ppt.GET("/list", controllers.GetPPtTemplates)

	// 图片生成接口
	ai.POST("/pic/generate", controllers.PicGenerate)

	//试题推荐接口
	ai.POST("/stu-rag", controllers.Rag4Stu)
	ai.POST("/plan-rag", controllers.Rag4Plan)
	ai.POST("/rag", controllers.Rag4Msg)
}
