package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Question struct {
	ID        uuid.UUID      `gorm:"primaryKey;column:id;type:uuid"`
	TopicID   uuid.UUID      `gorm:"column:topic_id;type:uuid;not null"`
	Content   string         `gorm:"column:content;type:text;not null"`
	Level     string         `gorm:"column:level;type:varchar(50);not null"`
	TimeLimit int            `gorm:"column:time_limit;type:integer;not null"`
	CreatedAt time.Time      `gorm:"column:created_at;default:now()"`
	UpdatedAt time.Time      `gorm:"column:updated_at;default:now()"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index"`
	Topic     Topic          `gorm:"foreignKey:TopicID"`
	Answers   []Answer       `gorm:"foreignKey:QuestionID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (Question) TableName() string {
	return "questions"
}
