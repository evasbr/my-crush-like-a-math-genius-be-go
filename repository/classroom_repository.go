package repository

import (
	"context"
	"evasbr/mclamg/entity"

	"github.com/google/uuid"
)

type ClassroomRepository interface {
	FindAll(ctx context.Context) ([]entity.Classroom, error)
	FindAllByUserID(ctx context.Context, userID uuid.UUID) ([]entity.Classroom, error)
	FindByID(ctx context.Context, id uuid.UUID) (entity.Classroom, error)
	FindByCode(ctx context.Context, code string) (entity.Classroom, error)
	Create(ctx context.Context, classroom entity.Classroom, creatorID uuid.UUID) (entity.Classroom, error)
	Update(ctx context.Context, classroom entity.Classroom) (entity.Classroom, error)
	Delete(ctx context.Context, id uuid.UUID) error
	AddMember(ctx context.Context, role entity.ClassroomRole) (entity.ClassroomRole, error)
	FindMembers(ctx context.Context, classroomID uuid.UUID) ([]entity.ClassroomRole, error)
	FindUserRole(ctx context.Context, classroomID uuid.UUID, userID uuid.UUID) (entity.ClassroomRole, error)
	FindUserRoles(ctx context.Context, userID uuid.UUID) ([]entity.ClassroomRole, error)
}
