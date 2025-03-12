package models

// EmailConfig 邮件服务配置
type EmailConfig struct {
	From      string
	FromAlias string
	Password  string
	Host      string
	Port      int
}