package service

import (
	"context"
	"evasbr/mclamg/model"
)

type QuestionService interface {
	FindAll(ctx context.Context, filter model.QuestionFilter, includeIsCorrect bool) ([]model.QuestionResponse, error)
	FindByID(ctx context.Context, id string, includeIsCorrect bool) (model.QuestionResponse, error)
	CreateBatch(ctx context.Context, request model.CreateQuestionBatchRequest) ([]model.QuestionResponse, error)
	Update(ctx context.Context, request model.UpdateQuestionRequest, id string) (model.QuestionResponse, error)
	Delete(ctx context.Context, id string) error
}
