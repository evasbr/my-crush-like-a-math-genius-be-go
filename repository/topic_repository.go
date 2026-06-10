package repository

import (
	"context"
	"evasbr/mclamg/entity"
	"evasbr/mclamg/model"

	"github.com/google/uuid"
)

type TopicRepository interface {
	FindAll(ctx context.Context, filter model.TopicFilter) ([]entity.Topic, error)
	FindByID(ctx context.Context, id uuid.UUID) (entity.Topic, error)
	Create(ctx context.Context, topic entity.Topic) (entity.Topic, error)
	Update(ctx context.Context, topic entity.Topic) (entity.Topic, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
