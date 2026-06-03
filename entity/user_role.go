package entity

import (
	"time"

	"github.com/google/uuid"
)

type UserRole struct {
	ID        uuid.UUID `gorm:"primaryKey;column:id;type:uuid"`
	UserID    uuid.UUID `gorm:"column:user_id;type:uuid"`
	RoleID    uuid.UUID `gorm:"column:role_id;type:uuid"`
	CreatedAt time.Time `gorm:"column:created_at;default:now()"`
	UpdatedAt time.Time `gorm:"column:updated_at;default:now()"`
	User      User      `gorm:"foreignKey:UserID"`
	Role      Role      `gorm:"foreignKey:RoleID"`
}

func (UserRole) TableName() string {
	return "user_roles"
}
