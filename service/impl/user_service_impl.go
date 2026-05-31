package impl

import (
	"context"
	"evasbr/mclamg/common"
	"evasbr/mclamg/entity"
	"evasbr/mclamg/exception"
	"evasbr/mclamg/model"
	"evasbr/mclamg/repository"
	"evasbr/mclamg/service"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

func NewUserServiceImpl(userRepository *repository.UserRepository) service.UserService {
	return &userServiceImpl{
		UserRepository: *userRepository,
		log:            common.Log.WithField("scope", "UserService"),
	}
}

type userServiceImpl struct {
	repository.UserRepository
	log *logrus.Entry
}

func (userService *userServiceImpl) Authentication(ctx context.Context, model model.UserModel) entity.User {
	userResult, err := userService.UserRepository.Authentication(ctx, model.Username)
	if err != nil {
		panic(exception.UnauthorizedError{
			Message: err.Error(),
		})
	}
	err = bcrypt.CompareHashAndPassword([]byte(userResult.Password), []byte(model.Password))
	if err != nil {
		panic(exception.UnauthorizedError{
			Message: "incorrect username and password",
		})
	}
	return userResult
}
