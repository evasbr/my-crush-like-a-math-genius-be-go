package service

import (
	"context"
	"evasbr/mclamg/entity"
	"evasbr/mclamg/model"
	"mime/multipart"
)

type UserService interface {
	FindAll(ctx context.Context, filter model.UserFilter) ([]entity.User, error)
	FindByID(ctx context.Context, id string) (entity.User, error)
	Update(ctx context.Context, model model.UpdateUser, id string) (entity.User, error)
	Delete(ctx context.Context, id string) error
	UpdateProfilePicture(ctx context.Context, userID string, fileHeader *multipart.FileHeader) (entity.User, error)
}
