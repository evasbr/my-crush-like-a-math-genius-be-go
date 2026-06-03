package entity

import "github.com/google/uuid"

type Role struct {
	ID          uuid.UUID              `gorm:"primaryKey;column:id;type:uuid"`
	Name        string                 `gorm:"column:name;type:varchar(255)"`
	RoleType    *string                `gorm:"column:role_type;type:varchar(100)"`
	Permissions map[string]interface{} `gorm:"serializer:json;column:permissions;type:jsonb"`
}

func (Role) TableName() string {
	return "roles"
}
