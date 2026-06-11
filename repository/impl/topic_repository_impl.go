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

type topicRepositoryImpl struct {
	DB *gorm.DB
}

func NewTopicRepositoryImpl(DB *gorm.DB) repository.TopicRepository {
	return &topicRepositoryImpl{DB: DB}
}

func (r *topicRepositoryImpl) FindAll(ctx context.Context, filter model.TopicFilter) ([]entity.Topic, error) {
	var topics []entity.Topic
	offset := (filter.Page - 1) * filter.Limit

	query := r.DB.WithContext(ctx)
	if filter.ClassroomID != "" {
		query = query.Where("classroom_id = ?", filter.ClassroomID)
	}

	err := query.
		Limit(filter.Limit).
		Offset(offset).
		Find(&topics).Error

	if err != nil {
		return nil, err
	}
	return topics, nil
}

func (r *topicRepositoryImpl) FindByID(ctx context.Context, id uuid.UUID) (entity.Topic, error) {
	var topic entity.Topic
	err := r.DB.WithContext(ctx).Where("id = ?", id).First(&topic).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return entity.Topic{}, errors.New("topic not found")
		}
		return entity.Topic{}, err
	}
	return topic, nil
}

func (r *topicRepositoryImpl) Create(ctx context.Context, topic entity.Topic) (entity.Topic, error) {
	err := r.DB.WithContext(ctx).Create(&topic).Error
	if err != nil {
		return entity.Topic{}, err
	}
	return topic, nil
}

func (r *topicRepositoryImpl) Update(ctx context.Context, topic entity.Topic) (entity.Topic, error) {
	err := r.DB.WithContext(ctx).Save(&topic).Error
	if err != nil {
		return entity.Topic{}, err
	}
	return topic, nil
}

func (r *topicRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.DB.WithContext(ctx).Where("id = ?", id).Delete(&entity.Topic{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("topic not found or already deleted")
	}
	return nil
}
