package config

import (
	"github.com/redis/go-redis/v9"
)

var RDB *redis.Client

func InitRedis() {
	RDB = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // 如果没有设置密码，使用空字符串
		DB:       0,  // 使用默认数据库
	})
}
