package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Topic struct {
	ID            uuid.UUID              `gorm:"primaryKey;column:id;type:uuid"`
	LevelSettings map[string]interface{} `gorm:"serializer:json;column:level_settings;type:jsonb"`
	MaxAttempts   int                    `gorm:"column:max_attempts;type:integer;not null"`
	CreatedAt     time.Time              `gorm:"column:created_at;default:now()"`
	UpdatedAt     time.Time              `gorm:"column:updated_at;default:now()"`
	DeletedAt     gorm.DeletedAt         `gorm:"column:deleted_at;index"`
	Questions     []Question             `gorm:"foreignKey:TopicID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (Topic) TableName() string {
	return "topics"
}
