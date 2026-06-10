package service

import (
	"context"
	"evasbr/mclamg/model"
)

type AttemptService interface {
	GetTopicAttemptInfo(ctx context.Context, topicId string, userId string) (model.TopicAttemptInfoResponse, error)
	StartAttempt(ctx context.Context, request model.StartAttemptRequest, userId string) (model.AttemptSessionResponse, error)
	GetNextQuestion(ctx context.Context, sessionId string, userId string) (model.ActiveQuestionResponse, error)
	SubmitAnswer(ctx context.Context, request model.SubmitAnswerRequest, userId string) (model.SubmitAnswerResponse, error)
	FindByID(ctx context.Context, id string, userId string) (model.AttemptSessionResponse, error)
	FindAll(ctx context.Context, filter model.AttemptFilter, userId string) ([]model.AttemptSessionResponse, int64, error)
}
