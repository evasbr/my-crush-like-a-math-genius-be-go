package impl

import (
	"context"
	"time"

	"evasbr/mclamg/common"
	"evasbr/mclamg/entity"
	"evasbr/mclamg/exception"
	"evasbr/mclamg/model"
	"evasbr/mclamg/repository"
	"evasbr/mclamg/service"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type topicServiceImpl struct {
	TopicRepository repository.TopicRepository
	log             *logrus.Entry
}

func NewTopicServiceImpl(topicRepository *repository.TopicRepository) service.TopicService {
	return &topicServiceImpl{
		TopicRepository: *topicRepository,
		log:             common.Log.WithField("scope", "TopicService"),
	}
}

func (s *topicServiceImpl) FindAll(ctx context.Context, filter model.TopicFilter) ([]model.TopicResponse, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 {
		filter.Limit = 10
	}

	topics, err := s.TopicRepository.FindAll(ctx, filter)
	if err != nil {
		return nil, err
	}

	var response []model.TopicResponse
	for _, topic := range topics {
		response = append(response, s.toTopicResponse(topic))
	}
	return response, nil
}

func (s *topicServiceImpl) FindByID(ctx context.Context, id string) (model.TopicResponse, error) {
	parsedUUID, err := uuid.Parse(id)
	if err != nil {
		return model.TopicResponse{}, exception.ValidationError{
			Message: "invalid topic ID format",
		}
	}

	topic, err := s.TopicRepository.FindByID(ctx, parsedUUID)
	if err != nil {
		return model.TopicResponse{}, exception.NotFoundError{
			Message: err.Error(),
		}
	}

	return s.toTopicResponse(topic), nil
}

func (s *topicServiceImpl) Create(ctx context.Context, request model.CreateTopicRequest) (model.TopicResponse, error) {
	common.Validate(request)

	levelSettings := make([]entity.LevelSetting, len(request.LevelSettings))
	for i, ls := range request.LevelSettings {
		levelSettings[i] = entity.LevelSetting{
			Level:      ls.Level,
			TrueScore:  ls.TrueScore,
			FalseScore: ls.FalseScore,
		}
	}

	topicEntity := entity.Topic{
		ID:            uuid.New(),
		Name:          request.Name,
		LevelSettings: levelSettings,
		MaxAttempts:   request.MaxAttempts,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	createdTopic, err := s.TopicRepository.Create(ctx, topicEntity)
	if err != nil {
		return model.TopicResponse{}, err
	}

	return s.toTopicResponse(createdTopic), nil
}

func (s *topicServiceImpl) Update(ctx context.Context, request model.UpdateTopicRequest, id string) (model.TopicResponse, error) {
	parsedUUID, err := uuid.Parse(id)
	if err != nil {
		return model.TopicResponse{}, exception.ValidationError{
			Message: "invalid topic ID format",
		}
	}

	common.Validate(request)

	levelSettings := make([]entity.LevelSetting, len(request.LevelSettings))
	for i, ls := range request.LevelSettings {
		levelSettings[i] = entity.LevelSetting{
			Level:      ls.Level,
			TrueScore:  ls.TrueScore,
			FalseScore: ls.FalseScore,
		}
	}

	topicEntity := entity.Topic{
		ID:            parsedUUID,
		Name:          request.Name,
		LevelSettings: levelSettings,
		MaxAttempts:   request.MaxAttempts,
		UpdatedAt:     time.Now(),
	}

	updatedTopic, err := s.TopicRepository.Update(ctx, topicEntity)
	if err != nil {
		return model.TopicResponse{}, exception.NotFoundError{
			Message: err.Error(),
		}
	}

	return s.toTopicResponse(updatedTopic), nil
}

func (s *topicServiceImpl) Delete(ctx context.Context, id string) error {
	parsedUUID, err := uuid.Parse(id)
	if err != nil {
		return exception.ValidationError{
			Message: "invalid topic ID format",
		}
	}

	err = s.TopicRepository.Delete(ctx, parsedUUID)
	if err != nil {
		return exception.NotFoundError{
			Message: err.Error(),
		}
	}

	return nil
}

func (s *topicServiceImpl) toTopicResponse(topic entity.Topic) model.TopicResponse {
	levelSettings := make([]model.LevelSettingDto, len(topic.LevelSettings))
	for i, ls := range topic.LevelSettings {
		levelSettings[i] = model.LevelSettingDto{
			Level:      ls.Level,
			TrueScore:  ls.TrueScore,
			FalseScore: ls.FalseScore,
		}
	}

	return model.TopicResponse{
		ID:            topic.ID.String(),
		Name:          topic.Name,
		LevelSettings: levelSettings,
		MaxAttempts:   topic.MaxAttempts,
		CreatedAt:     topic.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     topic.UpdatedAt.Format(time.RFC3339),
	}
}
