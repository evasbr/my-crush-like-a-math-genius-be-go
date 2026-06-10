package entity

import (
	"time"

	"github.com/google/uuid"
)

type AttemptDetail struct {
	ID               uuid.UUID      `gorm:"primaryKey;column:id;type:uuid"`
	AttemptSessionID uuid.UUID      `gorm:"column:attempt_session_id;type:uuid;not null"`
	QuestionID       uuid.UUID      `gorm:"column:question_id;type:uuid;not null"`
	AnswerID         *uuid.UUID     `gorm:"column:answer_id;type:uuid"`
	IsCorrect        *bool          `gorm:"column:is_correct;type:boolean"`
	AnsweredAt       *time.Time     `gorm:"column:answered_at"`
	AttemptSession   AttemptSession `gorm:"foreignKey:AttemptSessionID"`
	Question         Question       `gorm:"foreignKey:QuestionID"`
	Answer           *Answer        `gorm:"foreignKey:AnswerID"`
}

func (AttemptDetail) TableName() string {
	return "attempt_details"
}
