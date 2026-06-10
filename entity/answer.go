package entity

import (
	"github.com/google/uuid"
)

type Answer struct {
	ID         uuid.UUID `gorm:"primaryKey;column:id;type:uuid"`
	QuestionID uuid.UUID `gorm:"column:question_id;type:uuid;not null"`
	Content    string    `gorm:"column:content;type:text;not null"`
	IsCorrect  bool      `gorm:"column:is_correct;type:boolean;not null"`
	Question   Question  `gorm:"foreignKey:QuestionID"`
}

func (Answer) TableName() string {
	return "answers"
}
