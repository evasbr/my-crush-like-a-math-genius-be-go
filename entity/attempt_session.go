package entity

import (
	"time"

	"github.com/google/uuid"
)

type AttemptSession struct {
	ID                 uuid.UUID              `gorm:"primaryKey;column:id;type:uuid"`
	UserID             uuid.UUID              `gorm:"column:user_id;type:uuid;not null"`
	TopicID            uuid.UUID              `gorm:"column:topic_id;type:uuid;not null"`
	SelectedLevel      string                 `gorm:"column:selected_level;type:varchar(50);not null"`
	RequestedQuestions int                    `gorm:"column:requested_questions;type:integer;not null"`
	Score              *int                   `gorm:"column:score;type:integer"`
	Status             string                 `gorm:"column:status;type:varchar(50);not null"`
	StartedAt          time.Time              `gorm:"column:started_at;default:now();not null"`
	FinishedAt         *time.Time             `gorm:"column:finished_at"`
	MetaData           map[string]interface{} `gorm:"serializer:json;column:meta_data;type:jsonb"`
	User               User                   `gorm:"foreignKey:UserID"`
	Topic              Topic                  `gorm:"foreignKey:TopicID"`
	AttemptDetails     []AttemptDetail        `gorm:"foreignKey:AttemptSessionID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (AttemptSession) TableName() string {
	return "attempt_sessions"
}
