package impl

import (
	"context"
	"errors"
	"evasbr/mclamg/entity"
	"evasbr/mclamg/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type classroomRepositoryImpl struct {
	DB *gorm.DB
}

func NewClassroomRepositoryImpl(DB *gorm.DB) repository.ClassroomRepository {
	return &classroomRepositoryImpl{DB: DB}
}

func (r *classroomRepositoryImpl) FindAll(ctx context.Context) ([]entity.Classroom, error) {
	var classrooms []entity.Classroom
	err := r.DB.WithContext(ctx).Find(&classrooms).Error
	if err != nil {
		return nil, err
	}
	return classrooms, nil
}

func (r *classroomRepositoryImpl) FindAllByUserID(ctx context.Context, userID uuid.UUID) ([]entity.Classroom, error) {
	var classrooms []entity.Classroom
	err := r.DB.WithContext(ctx).
		Joins("JOIN classroom_roles ON classroom_roles.classroom_id = classrooms.id").
		Where("classroom_roles.user_id = ? AND classroom_roles.deleted_at IS NULL", userID).
		Find(&classrooms).Error
	if err != nil {
		return nil, err
	}
	return classrooms, nil
}

func (r *classroomRepositoryImpl) FindByID(ctx context.Context, id uuid.UUID) (entity.Classroom, error) {
	var classroom entity.Classroom
	err := r.DB.WithContext(ctx).Where("id = ?", id).First(&classroom).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return entity.Classroom{}, errors.New("classroom not found")
		}
		return entity.Classroom{}, err
	}
	return classroom, nil
}

func (r *classroomRepositoryImpl) FindByCode(ctx context.Context, code string) (entity.Classroom, error) {
	var classroom entity.Classroom
	err := r.DB.WithContext(ctx).Where("codes = ?", code).First(&classroom).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return entity.Classroom{}, errors.New("classroom not found with that code")
		}
		return entity.Classroom{}, err
	}
	return classroom, nil
}

func (r *classroomRepositoryImpl) Create(ctx context.Context, classroom entity.Classroom, creatorID uuid.UUID) (entity.Classroom, error) {
	err := r.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&classroom).Error; err != nil {
			return err
		}

		ownerRole := entity.ClassroomRole{
			ID:          uuid.New(),
			UserID:      creatorID,
			ClassroomID: classroom.ID,
			Role:        entity.RoleOwner,
		}

		if err := tx.Create(&ownerRole).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return entity.Classroom{}, err
	}
	return classroom, nil
}

func (r *classroomRepositoryImpl) Update(ctx context.Context, classroom entity.Classroom) (entity.Classroom, error) {
	err := r.DB.WithContext(ctx).Save(&classroom).Error
	if err != nil {
		return entity.Classroom{}, err
	}
	return classroom, nil
}

func (r *classroomRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.DB.WithContext(ctx).Where("id = ?", id).Delete(&entity.Classroom{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("classroom not found or already deleted")
	}
	return nil
}

func (r *classroomRepositoryImpl) AddMember(ctx context.Context, role entity.ClassroomRole) (entity.ClassroomRole, error) {
	err := r.DB.WithContext(ctx).Create(&role).Error
	if err != nil {
		return entity.ClassroomRole{}, err
	}
	return role, nil
}

func (r *classroomRepositoryImpl) FindMembers(ctx context.Context, classroomID uuid.UUID) ([]entity.ClassroomRole, error) {
	var roles []entity.ClassroomRole
	err := r.DB.WithContext(ctx).
		Preload("User").
		Where("classroom_id = ?", classroomID).
		Find(&roles).Error
	if err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *classroomRepositoryImpl) FindUserRole(ctx context.Context, classroomID uuid.UUID, userID uuid.UUID) (entity.ClassroomRole, error) {
	var role entity.ClassroomRole
	err := r.DB.WithContext(ctx).
		Where("classroom_id = ? AND user_id = ?", classroomID, userID).
		First(&role).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return entity.ClassroomRole{}, errors.New("membership not found")
		}
		return entity.ClassroomRole{}, err
	}
	return role, nil
}

func (r *classroomRepositoryImpl) FindUserRoles(ctx context.Context, userID uuid.UUID) ([]entity.ClassroomRole, error) {
	var roles []entity.ClassroomRole
	err := r.DB.WithContext(ctx).
		Where("user_id = ? AND deleted_at IS NULL", userID).
		Find(&roles).Error
	if err != nil {
		return nil, err
	}
	return roles, nil
}
