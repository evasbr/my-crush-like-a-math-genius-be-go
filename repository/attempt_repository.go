package repository

import (
	"context"
	"evasbr/mclamg/entity"
	"evasbr/mclamg/model"

	"github.com/google/uuid"
)

type AttemptRepository interface {
	GetTopicLevelCounts(ctx context.Context, topicId uuid.UUID, userId uuid.UUID) (map[string]int, map[string]int, error)
	GetUserAttemptsCount(ctx context.Context, topicId uuid.UUID, userId uuid.UUID) (int, error)
	FindAll(ctx context.Context, filter model.AttemptFilter, userId uuid.UUID) ([]entity.AttemptSession, int64, error)
	FindByID(ctx context.Context, id uuid.UUID, userId uuid.UUID) (entity.AttemptSession, error)
	Create(ctx context.Context, session entity.AttemptSession) (entity.AttemptSession, error)
	GetRandomUnattemptedQuestions(ctx context.Context, topicId uuid.UUID, userId uuid.UUID, level string, limit int) ([]entity.Question, error)
	SubmitAnswer(ctx context.Context, sessionId uuid.UUID, questionId uuid.UUID, answerId uuid.UUID, userId uuid.UUID) (bool, uuid.UUID, bool, *int, error)
	GetNextUnansweredQuestion(ctx context.Context, sessionId uuid.UUID, userId uuid.UUID) (*entity.Question, error)
	ExpireSession(ctx context.Context, sessionId uuid.UUID) (entity.AttemptSession, error)
}
