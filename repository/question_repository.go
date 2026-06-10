package repository

import (
	"context"
	"evasbr/mclamg/entity"
	"evasbr/mclamg/model"

	"github.com/google/uuid"
)

type QuestionRepository interface {
	FindAll(ctx context.Context, filter model.QuestionFilter) ([]entity.Question, error)
	FindByID(ctx context.Context, id uuid.UUID) (entity.Question, error)
	CreateBatch(ctx context.Context, questions []entity.Question) ([]entity.Question, error)
	Update(ctx context.Context, question entity.Question, options []entity.Answer) (entity.Question, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
