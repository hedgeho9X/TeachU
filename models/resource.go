package models

type Resource struct {
	ID        uint   `gorm:"primaryKey;autoIncrement"`
	ObjectKey string `gorm:"size:1000;not null"`
	FileName  string `gorm:"size:255;not null"`
	Grade     string `gorm:"size:50;not null;index:idx_subject_grade"`
	Subject   string `gorm:"size:50;not null;index:idx_subject_grade"`
	FileSize  string `gorm:"size:20;not null"`
	FileType  string `gorm:"size:20;not null"`
}
