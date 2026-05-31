package service

import (
	"context"
	"evasbr/mclamg/entity"
	"evasbr/mclamg/model"
)

type UserService interface {
	Authentication(ctx context.Context, model model.UserModel) entity.User
}
