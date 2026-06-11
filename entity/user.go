package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID                uuid.UUID        `gorm:"primaryKey;column:id;type:uuid"`
	FirstName         string           `gorm:"column:first_name;type:varchar(255)"`
	LastName          *string          `gorm:"column:last_name;type:varchar(255)"`
	Gender            *string          `gorm:"column:gender;type:varchar(50)"`
	Username          string           `gorm:"-"`
	Email             string           `gorm:"unique;column:email;type:varchar(255)"`
	Password          string           `gorm:"-" json:"-"`
	ProfilePictureURL *string          `gorm:"column:profile_picture_url;type:varchar(500)"`
	MetaData          *string          `gorm:"column:meta_data;type:jsonb"`
	Status            string           `gorm:"column:status;type:varchar(50)"`
	CreatedAt         time.Time        `gorm:"column:created_at;default:now()"`
	UpdatedAt         time.Time        `gorm:"column:updated_at;default:now()"`
	DeletedAt         gorm.DeletedAt   `gorm:"column:deleted_at;index"`
	UserRoles         []UserRole       `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Authentications   []Authentication `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	AttemptSessions   []AttemptSession `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	ClassroomRoles    []ClassroomRole  `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (User) TableName() string {
	return "users"
}
