package service

import (
	"context"
	"evasbr/mclamg/model"
)

type TopicService interface {
	FindAll(ctx context.Context, filter model.TopicFilter) ([]model.TopicResponse, error)
	FindByID(ctx context.Context, id string) (model.TopicResponse, error)
	Create(ctx context.Context, request model.CreateTopicRequest) (model.TopicResponse, error)
	Update(ctx context.Context, request model.UpdateTopicRequest, id string) (model.TopicResponse, error)
	Delete(ctx context.Context, id string) error
}
