package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ClassroomRoleType string

const (
	RoleOwner   ClassroomRoleType = "owner"
	RoleTeacher ClassroomRoleType = "teacher"
	RoleStudent ClassroomRoleType = "student"
)

type ClassroomRole struct {
	ID          uuid.UUID         `gorm:"primaryKey;column:id;type:uuid"`
	UserID      uuid.UUID         `gorm:"column:user_id;type:uuid;not null"`
	ClassroomID uuid.UUID         `gorm:"column:classroom_id;type:uuid;not null"`
	Role        ClassroomRoleType `gorm:"column:role;type:varchar(50);not null"`
	CreatedAt   time.Time         `gorm:"column:created_at;default:now()"`
	UpdatedAt   time.Time         `gorm:"column:updated_at;default:now()"`
	DeletedAt   gorm.DeletedAt    `gorm:"column:deleted_at;index"`
	User        User              `gorm:"foreignKey:UserID"`
	Classroom   Classroom         `gorm:"foreignKey:ClassroomID"`
}

func (ClassroomRole) TableName() string {
	return "classroom_roles"
}
