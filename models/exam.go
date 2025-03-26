package models

type Exam struct {
	ID        uint   `gorm:"primaryKey;column:id" json:"id"` // 确保主键定义
	UserId    uint   `json:"created_user_id"`
	ClassId   uint   `json:"class_id"`
	ExamName  string `json:"exam_name"`
	Subject   string `json:"subject"`
	CreatedAt string `gorm:"column:created_at"`
	// DeletedAt gorm.DeletedAt `gorm:"index"`
}

// TableName 指定表名为 exams
func (Exam) TableName() string {
	return "exams"
}

// Content   string `gorm:"type:text;column:content"`   // 解析后的内容
// Knowledge string `gorm:"type:text;column:knowledge"` // 知识点（可扩展为结构化数据）
