package service

import (
	"context"
	"evasbr/mclamg/model"
	"mime/multipart"
)

type TopicService interface {
	FindAll(ctx context.Context, filter model.TopicFilter) ([]model.TopicResponse, error)
	FindByID(ctx context.Context, id string) (model.TopicResponse, error)
	Create(ctx context.Context, request model.CreateTopicRequest, femaleNormalHeader, maleNormalHeader, femaleDatingHeader, maleDatingHeader *multipart.FileHeader) (model.TopicResponse, error)
	Update(ctx context.Context, request model.UpdateTopicRequest, femaleNormalHeader, maleNormalHeader, femaleDatingHeader, maleDatingHeader *multipart.FileHeader, id string) (model.TopicResponse, error)
	Delete(ctx context.Context, id string) error
}
