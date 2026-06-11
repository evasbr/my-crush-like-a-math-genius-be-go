package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LevelSetting struct {
	Level      string `json:"level"`
	TrueScore  int    `json:"true_score"`
	FalseScore int    `json:"false_score"`
}

type Topic struct {
	ID            uuid.UUID        `gorm:"primaryKey;column:id;type:uuid"`
	ClassroomID        uuid.UUID        `gorm:"column:classroom_id;type:uuid;not null"`
	Name               string           `gorm:"column:name;type:varchar(255);not null"`
	Description        *string          `gorm:"column:description;type:text"`
	FemaleNormalImg    string           `gorm:"column:female_normal_img;type:text;not null"`
	MaleNormalImg      string           `gorm:"column:male_normal_img;type:text;not null"`
	FemaleDatingImg    string           `gorm:"column:female_dating_img;type:text;not null"`
	MaleDatingImg      string           `gorm:"column:male_dating_img;type:text;not null"`
	FemaleNormalDialog string           `gorm:"column:female_normal_dialog;type:text;not null"`
	MaleNormalDialog   string           `gorm:"column:male_normal_dialog;type:text;not null"`
	FemaleDatingDialog string           `gorm:"column:female_dating_dialog;type:text;not null"`
	MaleDatingDialog   string           `gorm:"column:male_dating_dialog;type:text;not null"`
	Status             string           `gorm:"column:status;type:varchar(50);not null"`
	LevelSettings []LevelSetting   `gorm:"serializer:json;column:level_settings;type:jsonb"`
	MaxAttempts   int              `gorm:"column:max_attempts;type:integer;not null"`
	CreatedAt     time.Time        `gorm:"column:created_at;default:now()"`
	UpdatedAt     time.Time        `gorm:"column:updated_at;default:now()"`
	DeletedAt     gorm.DeletedAt   `gorm:"column:deleted_at;index"`
	Questions     []Question       `gorm:"foreignKey:TopicID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Classroom     *Classroom       `gorm:"foreignKey:ClassroomID"`
}

func (Topic) TableName() string {
	return "topics"
}
