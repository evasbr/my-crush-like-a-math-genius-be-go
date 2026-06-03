package impl

import (
	"context"
	"errors"
	"evasbr/mclamg/entity"
	"evasbr/mclamg/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type authRepositoryImpl struct {
	DB *gorm.DB
}

func NewAuthRepositoryImpl(DB *gorm.DB) repository.AuthRepository {
	return &authRepositoryImpl{DB: DB}
}

func (r *authRepositoryImpl) Register(ctx context.Context, payload repository.RegisterUserPayload) (entity.User, error) {
	var userEntity entity.User
	err := r.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		userID := uuid.New()

		userEntity = entity.User{
			ID:        userID,
			FirstName: payload.FirstName,
			LastName:  payload.LastName,
			Gender:    payload.Gender,
			Email:     payload.Email,
			Status:    "active",
		}

		if err := tx.Create(&userEntity).Error; err != nil {
			return err
		}

		for _, roleID := range payload.RoleIDs {
			var role entity.Role
			if err := tx.Where("id = ?", roleID).First(&role).Error; err != nil {
				return err
			}

			userRole := entity.UserRole{
				ID:     uuid.New(),
				UserID: userID,
				RoleID: role.ID,
			}
			if err := tx.Create(&userRole).Error; err != nil {
				return err
			}
		}

		pwdValues := payload.Password

		authEmail := entity.Authentication{
			ID:             uuid.New(),
			UserID:         userID,
			Method:         string(entity.MethodLocalEmail),
			ProviderUserID: payload.Email,
			Password:       &pwdValues,
			IsVerified:     true,
		}
		if err := tx.Create(&authEmail).Error; err != nil {
			return err
		}

		if payload.Username != nil && *payload.Username != "" {
			authUsername := entity.Authentication{
				ID:             uuid.New(),
				UserID:         userID,
				Method:         string(entity.MethodLocalUsername),
				ProviderUserID: *payload.Username,
				Password:       &pwdValues,
				IsVerified:     true,
			}
			if err := tx.Create(&authUsername).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return entity.User{}, err
	}

	return userEntity, nil
}

func (r *authRepositoryImpl) FindAuthentication(ctx context.Context, providerUserID string, methods []string) (entity.Authentication, error) {
	var auth entity.Authentication
	err := r.DB.WithContext(ctx).
		Preload("User").
		Preload("User.UserRoles.Role").
		Preload("User.Authentications").
		Where("provider_user_id = ? AND method IN ?", providerUserID, methods).
		First(&auth).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return entity.Authentication{}, errors.New("authentication not found")
		}
		return entity.Authentication{}, err
	}

	return auth, nil
}
