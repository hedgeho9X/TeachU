package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Hedgeho9X/TeachU/internal/config"
	"github.com/Hedgeho9X/TeachU/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// Rag4Plan 处理基于通用消息的 RAG 推荐请求（用于计划生成场景）
func Rag4Plan(c *gin.Context) {
	// 绑定请求参数
	var input struct {
		Msg string `json:"msg"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"msg":  "参数错误",
			"data": err.Error(),
		})
		return
	}
	prompt, _ := services.Chat(input.Msg, services.DoubaoLite, services.RagAiPrompt)
	result, _ := services.RagRecommend(prompt)
	c.JSON(http.StatusOK, gin.H{
		"code": 1,
		"msg":  "RAG生成成功",
		"data": result,
	})
}

// Rag4Msg 处理基于通用消息的 RAG 推荐请求（用于消息回复场景）
func Rag4Msg(c *gin.Context) {
	// 绑定请求参数
	var input struct {
		Msg string `json:"msg"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"msg":  "参数错误",
			"data": err.Error(),
		})
		return
	}
	prompt, _ := services.Chat(input.Msg, services.DoubaoLite, services.RagAiPrompt)
	result, _ := services.RagRecommend(prompt)
	c.JSON(http.StatusOK, gin.H{
		"code": 1,
		"msg":  "RAG生成成功",
		"data": result,
	})
}

// Rag4Stu 处理针对特定学生的 RAG 推荐请求
// 它首先检查 Redis 缓存，如果未命中，则获取学生分析数据，
// 调用 RAG 服务生成推荐，将结果存入 Redis 缓存，并返回给客户端。
func Rag4Stu(c *gin.Context) {
	var input struct {
		ExamID    uint `json:"exam_id"`
		StudentID uint `json:"student_id"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "参数绑定失败", "error": err.Error()})
		return
	}

	ctx := context.Background()
	// 定义 Redis 键
	redisKey := fmt.Sprintf("RAG-Recommend:examID:%d:studentID:%d", input.ExamID, input.StudentID)

	// 尝试从 Redis 获取缓存
	recommendJSON, err := config.RDB.Get(ctx, redisKey).Result()
	if err == nil { // 缓存命中
		log.Printf("缓存命中: %s", redisKey)
		// 将 JSON 字符串反序列化回 map 或结构体以便返回
		var recommendResult map[string]interface{} // 或者使用更具体的结构体
		if unmarshalErr := json.Unmarshal([]byte(recommendJSON), &recommendResult); unmarshalErr != nil {
			log.Printf("Redis 缓存反序列化失败: %v", unmarshalErr)
			// 即使反序列化失败，也可以考虑直接返回原始 JSON 字符串
			c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "success (cached, raw)", "recommend": recommendJSON})
		} else {
			c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "success (cached)", "recommend": recommendResult})
		}
		return
	} else if err != redis.Nil { // Redis 查询出错（非缓存未命中）
		log.Printf("Rag4Stu函数->Redis查询错误: %v", err)
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "查询缓存失败", "error": err.Error()})
		return
	}

	// --- 缓存未命中，执行 RAG 生成 ---
	log.Printf("缓存未命中: %s, 开始生成 RAG 推荐", redisKey)

	//  获取学生分析数据字符串
	analysisStr, err := services.StuAnalysis2String(input.ExamID, input.StudentID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "获取学生分析数据失败", "error": err.Error()})
		return
	}

	//  调用 AI 生成 RAG 查询 Prompt
	ragQueryPrompt, err := services.Chat(analysisStr, services.DoubaoLite, services.StudentRagPrompt)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "生成 RAG 查询失败", "error": err.Error()})
		return
	}

	//  调用 RAG 服务获取推荐结果 (返回的是 map[string]interface{} 或类似结构)
	recommendMap, err := services.RagRecommend(ragQueryPrompt)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "RAG 推荐生成失败", "error": err.Error()})
		return
	}

	//  将推荐结果 map 序列化为 JSON 字符串以便存入 Redis
	recommendJSONBytes, err := json.Marshal(recommendMap)
	if err != nil {
		log.Printf("RAG 推荐结果序列化失败: %v", err)
		// 即使序列化失败，仍然返回结果，但不进行缓存
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "success (serialization failed)", "recommend": recommendMap})
		return
	}
	recommendJSON = string(recommendJSONBytes) // 转换回 string 类型

	// 将 JSON 字符串存入 Redis
	err = config.RDB.Set(ctx, redisKey, recommendJSON, 1*time.Hour).Err()
	if err != nil {
		log.Printf("Redis 存储 RAG 推荐失败: %v", err)
		// 缓存失败不应阻止成功响应，记录日志
	} else {
		log.Printf("成功将 RAG 推荐存入 Redis: %s", redisKey)
	}

	//  返回原始的推荐结果 map 给客户端
	log.Printf("RAG 推荐生成成功 (ExamID: %d, StudentID: %d)", input.ExamID, input.StudentID)
	c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "success", "recommend": recommendMap})
}
