package models

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type ClassMetric struct {
	ID               uint           `gorm:"primaryKey"`
	ExamID           uint           `gorm:"index;not null"`
	StudentCount     int            `gorm:"not null"`
	TotalScore       float64        `gorm:"type:decimal(5,2);not null"`
	AvgTotalScore    float64        `gorm:"type:decimal(5,2);not null"`
	MaxTotalScore    float64        `gorm:"type:decimal(5,2);not null"`
	MinTotalScore    float64        `gorm:"type:decimal(5,2);not null"`
	MedianTotalScore float64        `gorm:"type:decimal(5,2);not null"`
	StdDev           float64        `gorm:"type:decimal(5,2);not null"`
	ScoreBuckets     datatypes.JSON `gorm:"type:json;not null"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        gorm.DeletedAt `gorm:"index"`
}
