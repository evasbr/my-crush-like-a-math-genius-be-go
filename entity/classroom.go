package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Classroom struct {
	ID                     uuid.UUID       `gorm:"primaryKey;column:id;type:uuid"`
	Name                   string          `gorm:"column:name;type:varchar(255);not null"`
	Description            *string         `gorm:"column:description;type:text"`
	Codes                  string          `gorm:"unique;column:codes;type:varchar(50);not null"`
	CoverImg               *string         `gorm:"column:cover_img;type:text"`
	WallpaperImg           *string         `gorm:"column:wallpaper_img;type:text"`
	IsExternalInviteEnable bool            `gorm:"column:is_external_invite_enable;type:boolean;default:true;not null"`
	Status                 string          `gorm:"column:status;type:varchar(50);not null"`
	CreatedAt              time.Time       `gorm:"column:created_at;default:now()"`
	UpdatedAt              time.Time       `gorm:"column:updated_at;default:now()"`
	DeletedAt              gorm.DeletedAt  `gorm:"column:deleted_at;index"`
	Members                []ClassroomRole `gorm:"foreignKey:ClassroomID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Topics                 []Topic         `gorm:"foreignKey:ClassroomID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (Classroom) TableName() string {
	return "classrooms"
}
