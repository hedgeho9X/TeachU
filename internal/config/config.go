package config

import (
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() {
	dsn := "root:Jerry2589:@tcp(47.109.64.142:3306)/teach_u?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}

	DB = db
	fmt.Println("数据库连接成功")
}

const (
	JWTSecret     = "86d6kZJNV11ok8U1lZNSxni+WfhJWzBbUTMzaLVDCSs=" // 在生产环境中应该使用环境变量
	XunFei_ID     = "ed5b171f"
	XunFei_Secret = "ZDUxMjllYWRlNjZhOWM3MjNlOWNhMTdi"
)
