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

type questionServiceImpl struct {
	QuestionRepository repository.QuestionRepository
	log                *logrus.Entry
}

func NewQuestionServiceImpl(questionRepository *repository.QuestionRepository) service.QuestionService {
	return &questionServiceImpl{
		QuestionRepository: *questionRepository,
		log:                common.Log.WithField("scope", "QuestionService"),
	}
}

func (s *questionServiceImpl) FindAll(ctx context.Context, filter model.QuestionFilter, includeIsCorrect bool) ([]model.QuestionResponse, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 {
		filter.Limit = 10
	}

	questions, err := s.QuestionRepository.FindAll(ctx, filter)
	if err != nil {
		return nil, err
	}

	var response []model.QuestionResponse
	for _, question := range questions {
		response = append(response, s.toQuestionResponse(question, includeIsCorrect))
	}
	return response, nil
}

func (s *questionServiceImpl) FindByID(ctx context.Context, id string, includeIsCorrect bool) (model.QuestionResponse, error) {
	parsedUUID, err := uuid.Parse(id)
	if err != nil {
		return model.QuestionResponse{}, exception.ValidationError{
			Message: "invalid question ID format",
		}
	}

	question, err := s.QuestionRepository.FindByID(ctx, parsedUUID)
	if err != nil {
		return model.QuestionResponse{}, exception.NotFoundError{
			Message: err.Error(),
		}
	}

	return s.toQuestionResponse(question, includeIsCorrect), nil
}

func (s *questionServiceImpl) CreateBatch(ctx context.Context, request model.CreateQuestionBatchRequest) ([]model.QuestionResponse, error) {
	common.Validate(request)

	parsedTopicID, err := uuid.Parse(request.TopicID)
	if err != nil {
		return nil, exception.ValidationError{
			Message: "invalid topic ID format",
		}
	}

	var questions []entity.Question
	for _, q := range request.Questions {
		correctCount := 0
		for _, o := range q.Options {
			if o.IsCorrect != nil && *o.IsCorrect {
				correctCount++
			}
		}
		if correctCount > 1 {
			return nil, exception.ValidationError{
				Message: "each question can only have at most one correct answer option",
			}
		}

		questionID := uuid.New()
		
		var answers []entity.Answer
		for _, o := range q.Options {
			isCorrectVal := false
			if o.IsCorrect != nil {
				isCorrectVal = *o.IsCorrect
			}

			answers = append(answers, entity.Answer{
				ID:         uuid.New(),
				QuestionID: questionID,
				Content:    o.Content,
				IsCorrect:  isCorrectVal,
			})
		}

		questions = append(questions, entity.Question{
			ID:        questionID,
			TopicID:   parsedTopicID,
			Content:   q.Content,
			Level:     request.Level,
			TimeLimit: q.TimeLimit,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Answers:   answers,
		})
	}

	createdQuestions, err := s.QuestionRepository.CreateBatch(ctx, questions)
	if err != nil {
		return nil, err
	}

	var response []model.QuestionResponse
	for _, q := range createdQuestions {
		response = append(response, s.toQuestionResponse(q, true)) // Return with isCorrect on creation
	}
	return response, nil
}

func (s *questionServiceImpl) Update(ctx context.Context, request model.UpdateQuestionRequest, id string) (model.QuestionResponse, error) {
	parsedQuestionID, err := uuid.Parse(id)
	if err != nil {
		return model.QuestionResponse{}, exception.ValidationError{
			Message: "invalid question ID format",
		}
	}

	common.Validate(request)

	correctCount := 0
	for _, o := range request.Options {
		if o.IsCorrect != nil && *o.IsCorrect {
			correctCount++
		}
	}
	if correctCount > 1 {
		return model.QuestionResponse{}, exception.ValidationError{
			Message: "each question can only have at most one correct answer option",
		}
	}

	questionEntity := entity.Question{
		ID:        parsedQuestionID,
		Content:   request.Content,
		TimeLimit: request.TimeLimit,
		UpdatedAt: time.Now(),
	}

	var options []entity.Answer
	for _, o := range request.Options {
		var answerID uuid.UUID
		if o.ID != nil && *o.ID != "" {
			parsedAnsID, err := uuid.Parse(*o.ID)
			if err != nil {
				return model.QuestionResponse{}, exception.ValidationError{
					Message: "invalid answer ID format: " + *o.ID,
				}
			}
			answerID = parsedAnsID
		} else {
			answerID = uuid.Nil
		}

		isCorrectVal := false
		if o.IsCorrect != nil {
			isCorrectVal = *o.IsCorrect
		}

		options = append(options, entity.Answer{
			ID:         answerID,
			QuestionID: parsedQuestionID,
			Content:    o.Content,
			IsCorrect:  isCorrectVal,
		})
	}

	updatedQuestion, err := s.QuestionRepository.Update(ctx, questionEntity, options)
	if err != nil {
		return model.QuestionResponse{}, exception.NotFoundError{
			Message: err.Error(),
		}
	}

	return s.toQuestionResponse(updatedQuestion, true), nil
}

func (s *questionServiceImpl) Delete(ctx context.Context, id string) error {
	parsedUUID, err := uuid.Parse(id)
	if err != nil {
		return exception.ValidationError{
			Message: "invalid question ID format",
		}
	}

	err = s.QuestionRepository.Delete(ctx, parsedUUID)
	if err != nil {
		return exception.NotFoundError{
			Message: err.Error(),
		}
	}

	return nil
}

func (s *questionServiceImpl) toQuestionResponse(q entity.Question, includeIsCorrect bool) model.QuestionResponse {
	options := make([]model.AnswerOptionDto, len(q.Answers))
	for i, o := range q.Answers {
		ansIDStr := o.ID.String()
		options[i] = model.AnswerOptionDto{
			ID:      &ansIDStr,
			Content: o.Content,
		}
		if includeIsCorrect {
			isCorrectVal := o.IsCorrect
			options[i].IsCorrect = &isCorrectVal
		}
	}

	return model.QuestionResponse{
		ID:        q.ID.String(),
		TopicID:   q.TopicID.String(),
		Content:   q.Content,
		Level:     q.Level,
		TimeLimit: q.TimeLimit,
		Options:   options,
		CreatedAt: q.CreatedAt.Format(time.RFC3339),
		UpdatedAt: q.UpdatedAt.Format(time.RFC3339),
	}
}
