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
	var session entity.AttemptSession
	err := r.DB.WithContext(ctx).
		Preload("AttemptDetails", func(db *gorm.DB) *gorm.DB {
			return db.Order("id ASC")
		}).
		Preload("AttemptDetails.Question").
		Preload("AttemptDetails.Question.Answers").
		Where("id = ? AND user_id = ?", sessionId, userId).
		First(&session).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	nowTime := time.Now()
	var activeQuestion *entity.Question

	deadlines := make([]time.Time, len(session.AttemptDetails))
	for i, d := range session.AttemptDetails {
		var startTime time.Time
		if i == 0 {
			startTime = session.StartedAt
		} else {
			prevDetail := session.AttemptDetails[i-1]
			if prevDetail.AnsweredAt != nil {
				startTime = *prevDetail.AnsweredAt
			} else {
				startTime = deadlines[i-1]
			}
		}
		deadlines[i] = startTime.Add(time.Duration(d.Question.TimeLimit) * time.Second)

		// Allow 2-second grace period/tolerance for network latency
		if d.AnswerID == nil && nowTime.Before(deadlines[i].Add(2*time.Second)) {
			activeQuestion = &d.Question
			break
		}
	}

	return activeQuestion, nil
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

		// 2. Verify and load all details to calculate cumulative limits and order
		var allDetails []entity.AttemptDetail
		if err := tx.Preload("Question").Where("attempt_session_id = ?", sessionId).Order("id ASC").Find(&allDetails).Error; err != nil {
			return err
		}

		nowTime := time.Now()
		var targetDetail *entity.AttemptDetail
		var targetIndex int = -1

		// Calculate deadlines dynamically for all details
		deadlines := make([]time.Time, len(allDetails))
		for i, d := range allDetails {
			var startTime time.Time
			if i == 0 {
				startTime = session.StartedAt
			} else {
				prevDetail := allDetails[i-1]
				if prevDetail.AnsweredAt != nil {
					startTime = *prevDetail.AnsweredAt
				} else {
					startTime = deadlines[i-1]
				}
			}
			deadlines[i] = startTime.Add(time.Duration(d.Question.TimeLimit) * time.Second)

			if d.QuestionID == questionId {
				targetDetail = &allDetails[i]
				targetIndex = i
			}
		}

		if targetDetail == nil {
			return errors.New("question not found in this attempt session")
		}

		if targetDetail.AnswerID != nil {
			return errors.New("question has already been answered")
		}

		// Check deadline for the target question (allowing 2-second grace period/tolerance)
		if nowTime.After(deadlines[targetIndex].Add(2 * time.Second)) {
			return errors.New("time limit exceeded for this question")
		}

		// Verify prior questions: they must either be answered or expired (allowing 2-second grace period/tolerance)
		for i := 0; i < targetIndex; i++ {
			if allDetails[i].AnswerID == nil && nowTime.Before(deadlines[i].Add(2*time.Second)) {
				return errors.New("cannot answer this question yet; please answer prior active questions first")
			}
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

		// 4. Update target detail
		targetDetail.AnswerID = &answerId
		targetDetail.IsCorrect = &isCorrect
		targetDetail.AnsweredAt = &nowTime
		if err := tx.Save(targetDetail).Error; err != nil {
			return err
		}

		// 5. Check if there are any remaining active questions left
		hasActiveQuestions := false
		tempDetails := make([]entity.AttemptDetail, len(allDetails))
		copy(tempDetails, allDetails)
		tempDetails[targetIndex].AnswerID = &answerId
		tempDetails[targetIndex].AnsweredAt = &nowTime

		tempDeadlines := make([]time.Time, len(tempDetails))
		for i, d := range tempDetails {
			var startTime time.Time
			if i == 0 {
				startTime = session.StartedAt
			} else {
				prevDetail := tempDetails[i-1]
				if prevDetail.AnsweredAt != nil {
					startTime = *prevDetail.AnsweredAt
				} else {
					startTime = tempDeadlines[i-1]
				}
			}
			tempDeadlines[i] = startTime.Add(time.Duration(d.Question.TimeLimit) * time.Second)

			// Allow 2-second grace period/tolerance for network latency
			if d.AnswerID == nil && nowTime.Before(tempDeadlines[i].Add(2*time.Second)) {
				hasActiveQuestions = true
				break
			}
		}

		if !hasActiveQuestions {
			isFinished = true
			session.Status = "FINISHED"
			session.FinishedAt = &nowTime

			// Calculate final score
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
				isCorrVal := d.IsCorrect
				if d.QuestionID == questionId {
					isCorrVal = &isCorrect
				}

				if isCorrVal != nil {
					if *isCorrVal {
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
