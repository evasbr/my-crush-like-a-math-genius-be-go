package repository

import (
	"context"
	"evasbr/mclamg/entity"
	"evasbr/mclamg/model"

	"github.com/google/uuid"
)

type UserRepository interface {
	FindAll(ctx context.Context, filter model.UserFilter) ([]entity.User, error)
	FindByID(ctx context.Context, id uuid.UUID) (entity.User, error)
	Update(ctx context.Context, user entity.User) (entity.User, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
