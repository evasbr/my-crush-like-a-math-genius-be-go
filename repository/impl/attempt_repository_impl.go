package impl

import (
	"context"
	"errors"
	"time"

	"evasbr/mclamg/common"
	"evasbr/mclamg/entity"
	"evasbr/mclamg/model"
	"evasbr/mclamg/repository"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type attemptRepositoryImpl struct {
	DB  *gorm.DB
	log *logrus.Entry
}

func NewAttemptRepositoryImpl(db *gorm.DB) repository.AttemptRepository {
	return &attemptRepositoryImpl{
		DB:  db,
		log: common.Log.WithField("scope", "AttemptRepository"),
	}
}

func (r *attemptRepositoryImpl) GetTopicLevelCounts(ctx context.Context, topicId uuid.UUID, userId uuid.UUID) (map[string]int, map[string]int, error) {
	totalCounts := make(map[string]int)
	attemptedCounts := make(map[string]int)

	// Query total counts per level
	type CountResult struct {
		Level string
		Count int
	}
	var totals []CountResult
	err := r.DB.WithContext(ctx).
		Model(&entity.Question{}).
		Select("level, COUNT(*) as count").
		Where("topic_id = ? AND deleted_at IS NULL", topicId).
		Group("level").
		Scan(&totals).Error
	if err != nil {
		return nil, nil, err
	}
	for _, t := range totals {
		totalCounts[t.Level] = t.Count
	}

	// Query attempted counts per level
	var attempted []CountResult
	err = r.DB.WithContext(ctx).
		Table("attempt_details ad").
		Select("q.level, COUNT(DISTINCT q.id) as count").
		Joins("JOIN attempt_sessions s ON ad.attempt_session_id = s.id").
		Joins("JOIN questions q ON ad.question_id = q.id").
		Where("s.topic_id = ? AND s.user_id = ? AND q.deleted_at IS NULL", topicId, userId).
		Group("q.level").
		Scan(&attempted).Error
	if err != nil {
		return nil, nil, err
	}
	for _, a := range attempted {
		attemptedCounts[a.Level] = a.Count
	}

	return totalCounts, attemptedCounts, nil
}

func (r *attemptRepositoryImpl) GetUserAttemptsCount(ctx context.Context, topicId uuid.UUID, userId uuid.UUID) (int, error) {
	var count int64
	err := r.DB.WithContext(ctx).
		Model(&entity.AttemptSession{}).
		Where("topic_id = ? AND user_id = ?", topicId, userId).
		Count(&count).Error
	return int(count), err
}

func (r *attemptRepositoryImpl) FindAll(ctx context.Context, filter model.AttemptFilter, userId uuid.UUID) ([]entity.AttemptSession, int64, error) {
	var sessions []entity.AttemptSession
	var total int64

	dbQuery := r.DB.WithContext(ctx).Model(&entity.AttemptSession{}).Where("user_id = ?", userId)

	if filter.TopicID != "" {
		parsedTopicID, err := uuid.Parse(filter.TopicID)
		if err == nil {
			dbQuery = dbQuery.Where("topic_id = ?", parsedTopicID)
		}
	}

	err := dbQuery.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	page := filter.Page
	if page <= 0 {
		page = 1
	}
	limit := filter.Limit
	if limit <= 0 {
		limit = 10
	}
	offset := (page - 1) * limit

	err = dbQuery.
		Preload("Topic").
		Order("started_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&sessions).Error
	if err != nil {
		return nil, 0, err
	}

	return sessions, total, nil
}

func (r *attemptRepositoryImpl) FindByID(ctx context.Context, id uuid.UUID, userId uuid.UUID) (entity.AttemptSession, error) {
	var session entity.AttemptSession
	err := r.DB.WithContext(ctx).
		Preload("Topic").
		Preload("AttemptDetails", func(db *gorm.DB) *gorm.DB {
			return db.Order("id ASC")
		}).
		Preload("AttemptDetails.Question").
		Preload("AttemptDetails.Question.Answers").
		Preload("AttemptDetails.Answer").
		Where("id = ? AND user_id = ?", id, userId).
		First(&session).Error
	if err != nil {
		return entity.AttemptSession{}, err
	}
	return session, nil
}

func (r *attemptRepositoryImpl) Create(ctx context.Context, session entity.AttemptSession) (entity.AttemptSession, error) {
	err := r.DB.WithContext(ctx).Create(&session).Error
	if err != nil {
		return entity.AttemptSession{}, err
	}
	return session, nil
}

func (r *attemptRepositoryImpl) GetRandomUnattemptedQuestions(ctx context.Context, topicId uuid.UUID, userId uuid.UUID, level string, limit int) ([]entity.Question, error) {
	var questions []entity.Question

	err := r.DB.WithContext(ctx).
		Preload("Answers").
		Where("topic_id = ? AND level = ? AND deleted_at IS NULL AND id NOT IN (SELECT question_id FROM attempt_details ad JOIN attempt_sessions s ON ad.attempt_session_id = s.id WHERE s.user_id = ?)", topicId, level, userId).
		Order("RANDOM()").
		Limit(limit).
		Find(&questions).Error
	if err != nil {
		return nil, err
	}

	return questions, nil
}

func (r *attemptRepositoryImpl) GetNextUnansweredQuestion(ctx context.Context, sessionId uuid.UUID, userId uuid.UUID) (*entity.Question, error) {
	var detail entity.AttemptDetail
	err := r.DB.WithContext(ctx).
		Joins("JOIN attempt_sessions s ON attempt_details.attempt_session_id = s.id").
		Where("attempt_details.attempt_session_id = ? AND s.user_id = ? AND attempt_details.answer_id IS NULL", sessionId, userId).
		Order("attempt_details.id ASC").
		First(&detail).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	var question entity.Question
	err = r.DB.WithContext(ctx).
		Preload("Answers").
		Where("id = ?", detail.QuestionID).
		First(&question).Error
	if err != nil {
		return nil, err
	}

	return &question, nil
}

func (r *attemptRepositoryImpl) SubmitAnswer(ctx context.Context, sessionId uuid.UUID, questionId uuid.UUID, answerId uuid.UUID, userId uuid.UUID) (bool, uuid.UUID, bool, *int, error) {
	var isCorrect bool
	var correctAnswerId uuid.UUID
	var isFinished bool
	var finalScore *int

	err := r.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Verify session
		var session entity.AttemptSession
		if err := tx.Preload("Topic").Where("id = ? AND user_id = ?", sessionId, userId).First(&session).Error; err != nil {
			return err
		}
		if session.Status != "STARTED" {
			return errors.New("quiz session is not in STARTED status")
		}

		// 2. Verify and load details
		var detail entity.AttemptDetail
		if err := tx.Where("attempt_session_id = ? AND question_id = ?", sessionId, questionId).First(&detail).Error; err != nil {
			return err
		}
		if detail.AnswerID != nil {
			return errors.New("question has already been answered")
		}

		// 3. Verify answer option and grade it
		var answer entity.Answer
		if err := tx.Where("id = ? AND question_id = ?", answerId, questionId).First(&answer).Error; err != nil {
			return errors.New("invalid answer option for this question")
		}

		// Find correct answer id
		var correctAnswer entity.Answer
		if err := tx.Where("question_id = ? AND is_correct = true", questionId).First(&correctAnswer).Error; err != nil {
			return errors.New("correct answer option not found for this question")
		}
		correctAnswerId = correctAnswer.ID
		isCorrect = answer.IsCorrect

		// 4. Update detail
		nowTime := time.Now()
		detail.AnswerID = &answerId
		detail.IsCorrect = &isCorrect
		detail.AnsweredAt = &nowTime
		if err := tx.Save(&detail).Error; err != nil {
			return err
		}

		// 5. Check if all questions are answered
		var totalCount int64
		var answeredCount int64
		if err := tx.Model(&entity.AttemptDetail{}).Where("attempt_session_id = ?", sessionId).Count(&totalCount).Error; err != nil {
			return err
		}
		if err := tx.Model(&entity.AttemptDetail{}).Where("attempt_session_id = ? AND answer_id IS NOT NULL", sessionId).Count(&answeredCount).Error; err != nil {
			return err
		}

		if totalCount == answeredCount {
			isFinished = true
			session.Status = "FINISHED"
			session.FinishedAt = &nowTime

			// Calculate final score
			var allDetails []entity.AttemptDetail
			if err := tx.Where("attempt_session_id = ?", sessionId).Find(&allDetails).Error; err != nil {
				return err
			}

			// Find correct scores from topic settings
			trueScore := 0
			falseScore := 0
			for _, setting := range session.Topic.LevelSettings {
				if setting.Level == session.SelectedLevel {
					trueScore = setting.TrueScore
					falseScore = setting.FalseScore
					break
				}
			}

			scoreVal := 0
			for _, d := range allDetails {
				if d.IsCorrect != nil {
					if *d.IsCorrect {
						scoreVal += trueScore
					} else {
						scoreVal += falseScore
					}
				}
			}
			session.Score = &scoreVal
			finalScore = &scoreVal

			if err := tx.Save(&session).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return false, uuid.Nil, false, nil, err
	}

	return isCorrect, correctAnswerId, isFinished, finalScore, nil
}

func (r *attemptRepositoryImpl) ExpireSession(ctx context.Context, sessionId uuid.UUID) (entity.AttemptSession, error) {
	var session entity.AttemptSession
	err := r.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Preload("Topic").Where("id = ?", sessionId).First(&session).Error; err != nil {
			return err
		}
		if session.Status != "STARTED" {
			return nil
		}

		nowTime := time.Now()
		session.Status = "FINISHED"
		session.FinishedAt = &nowTime

		// Calculate score for already answered details
		var allDetails []entity.AttemptDetail
		if err := tx.Where("attempt_session_id = ?", sessionId).Find(&allDetails).Error; err != nil {
			return err
		}

		trueScore := 0
		falseScore := 0
		for _, setting := range session.Topic.LevelSettings {
			if setting.Level == session.SelectedLevel {
				trueScore = setting.TrueScore
				falseScore = setting.FalseScore
				break
			}
		}

		scoreVal := 0
		for _, d := range allDetails {
			if d.IsCorrect != nil {
				if *d.IsCorrect {
					scoreVal += trueScore
				} else {
					scoreVal += falseScore
				}
			} else {
				// unanswered/expired questions score 0
				scoreVal += 0
			}
		}
		session.Score = &scoreVal

		if err := tx.Save(&session).Error; err != nil {
			return err
		}
		return nil
	})

	return session, err
}
