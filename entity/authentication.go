package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Authentication struct {
	ID             uuid.UUID      `gorm:"primaryKey;column:id;type:uuid"`
	UserID         uuid.UUID      `gorm:"column:user_id;type:uuid"`
	Method         string         `gorm:"column:method;type:varchar(50)"`
	ProviderUserID string         `gorm:"column:provider_user_id;type:varchar(255)"`
	Password       *string        `gorm:"column:password;type:varchar(255)"`
	IsVerified     bool           `gorm:"column:is_verified;type:boolean;default:false"`
	VerifiedAt     *time.Time     `gorm:"column:verified_at"`
	CreatedAt      time.Time      `gorm:"column:created_at;default:now()"`
	UpdatedAt      time.Time      `gorm:"column:updated_at;default:now()"`
	DeletedAt      gorm.DeletedAt `gorm:"column:deleted_at;index"`
	User           User           `gorm:"foreignKey:UserID"`
}

func (Authentication) TableName() string {
	return "authentications"
}

type AuthMethod string

const (
	MethodLocalEmail    AuthMethod = "local_email"
	MethodLocalUsername AuthMethod = "local_username"
	MethodGoogle        AuthMethod = "google"
	MethodGithub        AuthMethod = "github"
)
