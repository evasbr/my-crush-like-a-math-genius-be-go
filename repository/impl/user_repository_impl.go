package impl

import (
	"context"
	"errors"

	"evasbr/mclamg/entity"
	"evasbr/mclamg/model"
	"evasbr/mclamg/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type userRepositoryImpl struct {
	DB *gorm.DB
}

func NewUserRepositoryImpl(DB *gorm.DB) repository.UserRepository {
	return &userRepositoryImpl{DB: DB}
}

func (userRepository *userRepositoryImpl) FindByID(ctx context.Context, id uuid.UUID) (entity.User, error) {
	var user entity.User
	err := userRepository.DB.WithContext(ctx).
		Preload("UserRoles.Role").
		Preload("Authentications").
		Where("id = ? AND status = ?", id, "active").
		First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return entity.User{}, errors.New("user not found")
		}
		return entity.User{}, err
	}

	user.Username = user.Email
	for _, auth := range user.Authentications {
		if auth.Method == string(entity.MethodLocalUsername) {
			user.Username = auth.ProviderUserID
			break
		}
	}

	return user, nil
}

func (userRepository *userRepositoryImpl) FindAll(ctx context.Context, filter model.UserFilter) ([]entity.User, error) {
	var users []entity.User

	offset := (filter.Page - 1) * filter.Limit

	query := userRepository.DB.WithContext(ctx)

	// Jika IncludeDeleted = true, kita pakai .Unscoped() untuk menembus filter soft-delete bawaan GORM
	if filter.IncludeDeleted {
		query = query.Unscoped()
	} else {
		query = query.Where("status = ?", "active")
	}

	err := query.
		Preload("UserRoles.Role").
		Preload("Authentications").
		Limit(filter.Limit).
		Offset(offset).
		Find(&users).Error

	if err != nil {
		return nil, err
	}

	// Mapping username secara massal di memori
	for i := range users {
		users[i].Username = users[i].Email
		for _, auth := range users[i].Authentications {
			if auth.Method == string(entity.MethodLocalUsername) {
				users[i].Username = auth.ProviderUserID
				break
			}
		}
	}

	return users, nil
}


func (userRepository *userRepositoryImpl) Update(ctx context.Context, user entity.User) (entity.User, error) {
	err := userRepository.DB.WithContext(ctx).
		Model(&user).
		Where("id = ? AND status = ?", user.ID, "active").
		Updates(entity.User{
			FirstName:         user.FirstName,
			LastName:          user.LastName,
			Gender:            user.Gender,
			ProfilePictureURL: user.ProfilePictureURL,
			MetaData:          user.MetaData,
		}).Error

	if err != nil {
		return entity.User{}, err
	}

	return user, nil
}

func (userRepository *userRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	err := userRepository.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {

		result := tx.Model(&entity.User{}).Where("id = ? AND status = ?", id, "active").Update("status", "inactive")
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return errors.New("user not found or already deleted")
		}

		if err := tx.Where("id = ?", id).Delete(&entity.User{}).Error; err != nil {
			return err
		}

		if err := tx.Where("user_id = ?", id).Delete(&entity.Authentication{}).Error; err != nil {
			return err
		}

		return nil
	})

	return err
}
