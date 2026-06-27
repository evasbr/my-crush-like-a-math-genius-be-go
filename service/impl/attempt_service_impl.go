package impl

import (
	"context"
	"fmt"
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

type attemptServiceImpl struct {
	AttemptRepository repository.AttemptRepository
	TopicRepository   repository.TopicRepository
	log               *logrus.Entry
}

func NewAttemptServiceImpl(attemptRepository *repository.AttemptRepository, topicRepository *repository.TopicRepository) service.AttemptService {
	return &attemptServiceImpl{
		AttemptRepository: *attemptRepository,
		TopicRepository:   *topicRepository,
		log:               common.Log.WithField("scope", "AttemptService"),
	}
}

func (s *attemptServiceImpl) GetTopicAttemptInfo(ctx context.Context, topicId string, userId string) (model.TopicAttemptInfoResponse, error) {
	parsedTopicID, err := uuid.Parse(topicId)
	if err != nil {
		return model.TopicAttemptInfoResponse{}, exception.ValidationError{Message: "invalid topic ID format"}
	}
	parsedUserID, err := uuid.Parse(userId)
	if err != nil {
		return model.TopicAttemptInfoResponse{}, exception.ValidationError{Message: "invalid user ID format"}
	}

	topic, err := s.TopicRepository.FindByID(ctx, parsedTopicID)
	if err != nil {
		return model.TopicAttemptInfoResponse{}, exception.NotFoundError{Message: "topic not found"}
	}

	currentAttempts, err := s.AttemptRepository.GetUserAttemptsCount(ctx, parsedTopicID, parsedUserID)
	if err != nil {
		return model.TopicAttemptInfoResponse{}, err
	}

	totalCounts, attemptedCounts, err := s.AttemptRepository.GetTopicLevelCounts(ctx, parsedTopicID, parsedUserID)
	if err != nil {
		return model.TopicAttemptInfoResponse{}, err
	}

	levelSettings := make([]model.TopicAttemptLevelSetting, len(topic.LevelSettings))
	for i, ls := range topic.LevelSettings {
		total := totalCounts[ls.Level]
		attempted := attemptedCounts[ls.Level]
		remaining := total - attempted
		if remaining < 0 {
			remaining = 0
		}

		levelSettings[i] = model.TopicAttemptLevelSetting{
			Level:              ls.Level,
			TrueScore:          ls.TrueScore,
			FalseScore:         ls.FalseScore,
			TotalQuestions:     total,
			RemainingQuestions: remaining,
		}
	}

	remainingAttempts := topic.MaxAttempts - currentAttempts
	if remainingAttempts < 0 {
		remainingAttempts = 0
	}

	return model.TopicAttemptInfoResponse{
		MaxAttempts:       topic.MaxAttempts,
		CurrentAttempts:   currentAttempts,
		RemainingAttempts: remainingAttempts,
		LevelSettings:     levelSettings,
	}, nil
}

func (s *attemptServiceImpl) StartAttempt(ctx context.Context, request model.StartAttemptRequest, userId string) (model.AttemptSessionResponse, error) {
	common.Validate(request)

	parsedTopicID, err := uuid.Parse(request.TopicID)
	if err != nil {
		return model.AttemptSessionResponse{}, exception.ValidationError{Message: "invalid topic ID format"}
	}
	parsedUserID, err := uuid.Parse(userId)
	if err != nil {
		return model.AttemptSessionResponse{}, exception.ValidationError{Message: "invalid user ID format"}
	}

	topic, err := s.TopicRepository.FindByID(ctx, parsedTopicID)
	if err != nil {
		return model.AttemptSessionResponse{}, exception.NotFoundError{Message: "topic not found"}
	}

	// 1. Check max attempts
	currentAttempts, err := s.AttemptRepository.GetUserAttemptsCount(ctx, parsedTopicID, parsedUserID)
	if err != nil {
		return model.AttemptSessionResponse{}, err
	}
	if currentAttempts >= topic.MaxAttempts {
		return model.AttemptSessionResponse{}, exception.ValidationError{Message: "max attempts reached for this topic"}
	}

	// 2. Select questions
	questions, err := s.AttemptRepository.GetRandomUnattemptedQuestions(ctx, parsedTopicID, parsedUserID, request.Level, request.RequestedQuestions)
	if err != nil {
		return model.AttemptSessionResponse{}, err
	}
	if len(questions) < request.RequestedQuestions {
		return model.AttemptSessionResponse{}, exception.ValidationError{
			Message: fmt.Sprintf("not enough remaining questions available at level '%s'. Required: %d, Available: %d", request.Level, request.RequestedQuestions, len(questions)),
		}
	}

	// 3. Calculate time limit and expires_at
	totalTimeLimit := 0
	for _, q := range questions {
		totalTimeLimit += q.TimeLimit
	}

	nowTime := time.Now()
	expiresTime := nowTime.Add(time.Duration(totalTimeLimit) * time.Second)

	metaData := map[string]interface{}{
		"expires_at": expiresTime.Format(time.RFC3339),
	}

	attemptSessionID := uuid.New()
	var details []entity.AttemptDetail
	for _, q := range questions {
		details = append(details, entity.AttemptDetail{
			ID:               uuid.New(),
			AttemptSessionID: attemptSessionID,
			QuestionID:       q.ID,
		})
	}

	sessionEntity := entity.AttemptSession{
		ID:                 attemptSessionID,
		UserID:             parsedUserID,
		TopicID:            parsedTopicID,
		SelectedLevel:      request.Level,
		RequestedQuestions: request.RequestedQuestions,
		Status:             "STARTED",
		StartedAt:          nowTime,
		MetaData:           metaData,
		AttemptDetails:     details,
	}

	createdSession, err := s.AttemptRepository.Create(ctx, sessionEntity)
	if err != nil {
		return model.AttemptSessionResponse{}, err
	}

	// Fetch fully loaded session
	fullSession, err := s.AttemptRepository.FindByID(ctx, createdSession.ID, parsedUserID)
	if err != nil {
		return model.AttemptSessionResponse{}, err
	}

	return s.toAttemptSessionResponse(fullSession, false, false), nil
}

func (s *attemptServiceImpl) checkExpiration(ctx context.Context, session entity.AttemptSession) (entity.AttemptSession, error) {
	if session.Status != "STARTED" {
		return session, nil
	}

	expiresAtStr, ok := session.MetaData["expires_at"].(string)
	if !ok {
		return session, nil
	}

	expiresAt, err := time.Parse(time.RFC3339, expiresAtStr)
	if err != nil {
		return session, nil
	}

	if time.Now().After(expiresAt) {
		updatedSession, err := s.AttemptRepository.ExpireSession(ctx, session.ID)
		if err != nil {
			return session, err
		}
		return updatedSession, nil
	}

	return session, nil
}

func (s *attemptServiceImpl) GetNextQuestion(ctx context.Context, sessionId string, userId string) (model.ActiveQuestionResponse, error) {
	parsedSessionID, err := uuid.Parse(sessionId)
	if err != nil {
		return model.ActiveQuestionResponse{}, exception.ValidationError{Message: "invalid session ID format"}
	}
	parsedUserID, err := uuid.Parse(userId)
	if err != nil {
		return model.ActiveQuestionResponse{}, exception.ValidationError{Message: "invalid user ID format"}
	}

	session, err := s.AttemptRepository.FindByID(ctx, parsedSessionID, parsedUserID)
	if err != nil {
		return model.ActiveQuestionResponse{}, exception.NotFoundError{Message: "attempt session not found"}
	}

	// Check deadline
	session, err = s.checkExpiration(ctx, session)
	if err != nil {
		return model.ActiveQuestionResponse{}, err
	}

	if session.Status == "FINISHED" {
		return model.ActiveQuestionResponse{
			IsFinished: true,
		}, exception.SessionExpiredError{Message: "attempt session has expired or finished"}
	}

	question, err := s.AttemptRepository.GetNextUnansweredQuestion(ctx, parsedSessionID, parsedUserID)
	if err != nil {
		return model.ActiveQuestionResponse{}, err
	}

	if question == nil {
		return model.ActiveQuestionResponse{
			IsFinished: true,
		}, nil
	}

	// Map question response (redacting isCorrect)
	options := make([]model.AnswerOptionDto, len(question.Answers))
	for i, o := range question.Answers {
		ansIDStr := o.ID.String()
		options[i] = model.AnswerOptionDto{
			ID:      &ansIDStr,
			Content: o.Content,
		}
	}

	qResponse := model.QuestionResponse{
		ID:        question.ID.String(),
		TopicID:   question.TopicID.String(),
		Content:   question.Content,
		Level:     question.Level,
		TimeLimit: question.TimeLimit,
		Options:   options,
		CreatedAt: question.CreatedAt.Format(time.RFC3339),
		UpdatedAt: question.UpdatedAt.Format(time.RFC3339),
	}

	return model.ActiveQuestionResponse{
		IsFinished: false,
		Question:   &qResponse,
	}, nil
}

func (s *attemptServiceImpl) SubmitAnswer(ctx context.Context, request model.SubmitAnswerRequest, userId string) (model.SubmitAnswerResponse, error) {
	common.Validate(request)

	parsedSessionID, err := uuid.Parse(request.AttemptSessionID)
	if err != nil {
		return model.SubmitAnswerResponse{}, exception.ValidationError{Message: "invalid session ID format"}
	}
	parsedQuestionID, err := uuid.Parse(request.QuestionID)
	if err != nil {
		return model.SubmitAnswerResponse{}, exception.ValidationError{Message: "invalid question ID format"}
	}
	parsedAnswerID, err := uuid.Parse(request.AnswerID)
	if err != nil {
		return model.SubmitAnswerResponse{}, exception.ValidationError{Message: "invalid answer ID format"}
	}
	parsedUserID, err := uuid.Parse(userId)
	if err != nil {
		return model.SubmitAnswerResponse{}, exception.ValidationError{Message: "invalid user ID format"}
	}

	session, err := s.AttemptRepository.FindByID(ctx, parsedSessionID, parsedUserID)
	if err != nil {
		return model.SubmitAnswerResponse{}, exception.NotFoundError{Message: "attempt session not found"}
	}

	// Check deadline
	session, err = s.checkExpiration(ctx, session)
	if err != nil {
		return model.SubmitAnswerResponse{}, err
	}

	if session.Status == "FINISHED" {
		return model.SubmitAnswerResponse{}, exception.SessionExpiredError{Message: "attempt session has expired or finished"}
	}

	isCorrect, correctAnswerId, isFinished, finalScore, err := s.AttemptRepository.SubmitAnswer(ctx, parsedSessionID, parsedQuestionID, parsedAnswerID, parsedUserID)
	if err != nil {
		return model.SubmitAnswerResponse{}, exception.ValidationError{Message: err.Error()}
	}

	return model.SubmitAnswerResponse{
		IsCorrect:       isCorrect,
		CorrectAnswerID: correctAnswerId.String(),
		IsFinished:      isFinished,
		Score:           finalScore,
	}, nil
}

func (s *attemptServiceImpl) FindByID(ctx context.Context, id string, userId string) (model.AttemptSessionResponse, error) {
	parsedSessionID, err := uuid.Parse(id)
	if err != nil {
		return model.AttemptSessionResponse{}, exception.ValidationError{Message: "invalid session ID format"}
	}
	parsedUserID, err := uuid.Parse(userId)
	if err != nil {
		return model.AttemptSessionResponse{}, exception.ValidationError{Message: "invalid user ID format"}
	}

	session, err := s.AttemptRepository.FindByID(ctx, parsedSessionID, parsedUserID)
	if err != nil {
		return model.AttemptSessionResponse{}, exception.NotFoundError{Message: "attempt session not found"}
	}

	// Check deadline
	session, err = s.checkExpiration(ctx, session)
	if err != nil {
		return model.AttemptSessionResponse{}, err
	}

	showIsCorrect := session.Status == "FINISHED"
	return s.toAttemptSessionResponse(session, showIsCorrect, session.Status == "FINISHED"), nil
}

func (s *attemptServiceImpl) GetAttemptDetails(ctx context.Context, attemptId string, userId string) ([]model.AttemptDetailDto, error) {
	parsedSessionID, err := uuid.Parse(attemptId)
	if err != nil {
		return nil, exception.ValidationError{Message: "invalid attempt ID format"}
	}
	parsedUserID, err := uuid.Parse(userId)
	if err != nil {
		return nil, exception.ValidationError{Message: "invalid user ID format"}
	}

	session, err := s.AttemptRepository.FindByID(ctx, parsedSessionID, parsedUserID)
	if err != nil {
		return nil, exception.NotFoundError{Message: "attempt session not found"}
	}

	// Check deadline
	session, err = s.checkExpiration(ctx, session)
	if err != nil {
		return nil, err
	}

	showIsCorrect := session.Status == "FINISHED"
	response := s.toAttemptSessionResponse(session, showIsCorrect, true)
	return response.Details, nil
}

func (s *attemptServiceImpl) FindAll(ctx context.Context, filter model.AttemptFilter, userId string) ([]model.AttemptSessionResponse, int64, error) {
	parsedUserID, err := uuid.Parse(userId)
	if err != nil {
		return nil, 0, exception.ValidationError{Message: "invalid user ID format"}
	}

	sessions, total, err := s.AttemptRepository.FindAll(ctx, filter, parsedUserID)
	if err != nil {
		return nil, 0, err
	}

	var response []model.AttemptSessionResponse
	for _, sEntity := range sessions {
		// Just in case, check expiration for started sessions on history list
		if sEntity.Status == "STARTED" {
			sEntity, _ = s.checkExpiration(ctx, sEntity)
		}
		response = append(response, s.toAttemptSessionResponse(sEntity, sEntity.Status == "FINISHED", false))
	}

	return response, total, nil
}

func (s *attemptServiceImpl) toAttemptSessionResponse(session entity.AttemptSession, showIsCorrect bool, includeDetails bool) model.AttemptSessionResponse {
	expiresAtStr, _ := session.MetaData["expires_at"].(string)

	var details []model.AttemptDetailDto
	if includeDetails {
		for _, d := range session.AttemptDetails {
			var selectedAnsStr *string
			if d.AnswerID != nil {
				sAns := d.AnswerID.String()
				selectedAnsStr = &sAns
			}

			options := make([]model.AnswerOptionDto, len(d.Question.Answers))
			for j, o := range d.Question.Answers {
				ansIDStr := o.ID.String()
				opt := model.AnswerOptionDto{
					ID:      &ansIDStr,
					Content: o.Content,
				}
				// Only include correctness if showIsCorrect is true
				if showIsCorrect {
					isCor := o.IsCorrect
					opt.IsCorrect = &isCor
				}
				options[j] = opt
			}

			var answeredAtStr *string
			if d.AnsweredAt != nil {
				aAt := d.AnsweredAt.Format(time.RFC3339)
				answeredAtStr = &aAt
			}

			var detailIsCorrect *bool
			if showIsCorrect {
				detailIsCorrect = d.IsCorrect
			}

			details = append(details, model.AttemptDetailDto{
				QuestionID:       d.QuestionID.String(),
				QuestionContent:  d.Question.Content,
				Options:          options,
				SelectedAnswerID: selectedAnsStr,
				IsCorrect:        detailIsCorrect,
				AnsweredAt:       answeredAtStr,
			})
		}
	}

	var finishedAtStr *string
	if session.FinishedAt != nil {
		fAt := session.FinishedAt.Format(time.RFC3339)
		finishedAtStr = &fAt
	}

	return model.AttemptSessionResponse{
		ID:                 session.ID.String(),
		UserID:             session.UserID.String(),
		TopicID:            session.TopicID.String(),
		TopicName:          session.Topic.Name,
		SelectedLevel:      session.SelectedLevel,
		RequestedQuestions: session.RequestedQuestions,
		Score:              session.Score,
		Status:             session.Status,
		StartedAt:          session.StartedAt.Format(time.RFC3339),
		FinishedAt:         finishedAtStr,
		ExpiresAt:          expiresAtStr,
		Details:            details,
	}
}
