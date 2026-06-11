package impl

import (
	"context"
	"fmt"
	"mime/multipart"

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
	Storage        common.FileStorage
	log            *logrus.Entry
}

func NewUserServiceImpl(userRepository *repository.UserRepository, storage common.FileStorage) service.UserService {
	return &userServiceImpl{
		UserRepository: *userRepository,
		Storage:        storage,
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
	userEntity := entity.User{
		ID:                parsedUUID,
		FirstName:         input.FirstName,
		LastName:          input.LastName,
		Gender:            input.Gender,
		ProfilePictureURL: input.ProfilePictureURL,
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

func (userService *userServiceImpl) UpdateProfilePicture(ctx context.Context, userID string, fileHeader *multipart.FileHeader) (entity.User, error) {
	// 1. Validate the file constraints
	err := common.ValidateImageFile(fileHeader)
	if err != nil {
		return entity.User{}, err
	}

	// 2. Parse User ID and fetch the current user profile
	parsedUUID, err := uuid.Parse(userID)
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

	// 3. Open the file stream for upload
	file, err := fileHeader.Open()
	if err != nil {
		return entity.User{}, exception.ValidationError{
			Message: fmt.Sprintf("unable to open file stream: %v", err),
		}
	}
	defer file.Close()

	// 4. Delete the old profile picture from storage if it exists
	if userResult.ProfilePictureURL != nil && *userResult.ProfilePictureURL != "" {
		err = userService.Storage.DeleteFile(ctx, *userResult.ProfilePictureURL)
		if err != nil {
			userService.log.Warnf("failed to delete old profile picture from storage: %v", err)
		}
	}

	// 5. Upload the new image to storage
	filename := fmt.Sprintf("%s_%s", userID, fileHeader.Filename)
	newURL, err := userService.Storage.UploadFile(ctx, file, filename, common.FolderProfilePictures)
	if err != nil {
		return entity.User{}, err
	}

	// 6. Update user's ProfilePictureURL in database
	userResult.ProfilePictureURL = &newURL
	_, err = userService.UserRepository.Update(ctx, userResult)
	if err != nil {
		return entity.User{}, err
	}

	return userResult, nil
}
