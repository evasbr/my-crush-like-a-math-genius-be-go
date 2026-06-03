package impl

import (
	"context"

	"evasbr/mclamg/common"
	"evasbr/mclamg/entity"
	"evasbr/mclamg/exception"
	"evasbr/mclamg/model"
	"evasbr/mclamg/repository"
	"evasbr/mclamg/service"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type userServiceImpl struct {
	UserRepository repository.UserRepository
	log            *logrus.Entry
}

func NewUserServiceImpl(userRepository *repository.UserRepository) service.UserService {
	return &userServiceImpl{
		UserRepository: *userRepository,
		log:            common.Log.WithField("scope", "UserService"),
	}
}


func (userService *userServiceImpl) FindByID(ctx context.Context, id string) (entity.User, error) {
	parsedUUID, err := uuid.Parse(id)
	if err != nil {
		return entity.User{}, exception.ValidationError{
			Message: "invalid user ID format",
		}
	}

	userResult, err := userService.UserRepository.FindByID(ctx, parsedUUID)
	if err != nil {
		return entity.User{}, exception.NotFoundError{
			Message: err.Error(),
		}
	}

	return userResult, nil
}

func (userService *userServiceImpl) FindAll(ctx context.Context, filter model.UserFilter) ([]entity.User, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 {
		filter.Limit = 10
	}

	users, err := userService.UserRepository.FindAll(ctx, filter)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (userService *userServiceImpl) Update(ctx context.Context, input model.UpdateUser, id string) (entity.User, error) {
	parsedUUID, err := uuid.Parse(id)
	if err != nil {
		return entity.User{}, exception.ValidationError{
			Message: "invalid user ID format",
		}
	}

	// Buat payload entity untuk dikirim ke Repository
	// Catatan: Pastikan model.UserModel kamu mendukung mapping data opsional ke format pointer
	userEntity := entity.User{
		ID:        parsedUUID,
		FirstName: input.FirstName,
		// Ambil pointer string dari model input jika ada
		LastName: nil,
		Gender:   nil,
	}

	updatedUser, err := userService.UserRepository.Update(ctx, userEntity)
	if err != nil {
		return entity.User{}, exception.NotFoundError{
			Message: err.Error(),
		}
	}

	return updatedUser, nil
}

func (userService *userServiceImpl) Delete(ctx context.Context, id string) error {
	parsedUUID, err := uuid.Parse(id)
	if err != nil {
		return exception.ValidationError{
			Message: "invalid user ID format",
		}
	}

	err = userService.UserRepository.Delete(ctx, parsedUUID)
	if err != nil {
		return exception.NotFoundError{
			Message: err.Error(),
		}
	}

	return nil
}
