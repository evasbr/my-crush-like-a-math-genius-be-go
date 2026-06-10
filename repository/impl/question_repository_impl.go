package impl

import (
	"context"
	"errors"
	"evasbr/mclamg/entity"
	"evasbr/mclamg/model"
	"evasbr/mclamg/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type questionRepositoryImpl struct {
	DB *gorm.DB
}

func NewQuestionRepositoryImpl(DB *gorm.DB) repository.QuestionRepository {
	return &questionRepositoryImpl{DB: DB}
}

func (r *questionRepositoryImpl) FindAll(ctx context.Context, filter model.QuestionFilter) ([]entity.Question, error) {
	var questions []entity.Question
	offset := (filter.Page - 1) * filter.Limit

	query := r.DB.WithContext(ctx).Preload("Answers")

	if filter.TopicID != "" {
		parsedTopicID, err := uuid.Parse(filter.TopicID)
		if err == nil {
			query = query.Where("topic_id = ?", parsedTopicID)
		}
	}

	err := query.
		Limit(filter.Limit).
		Offset(offset).
		Find(&questions).Error

	if err != nil {
		return nil, err
	}
	return questions, nil
}

func (r *questionRepositoryImpl) FindByID(ctx context.Context, id uuid.UUID) (entity.Question, error) {
	var question entity.Question
	err := r.DB.WithContext(ctx).
		Preload("Answers").
		Where("id = ?", id).
		First(&question).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return entity.Question{}, errors.New("question not found")
		}
		return entity.Question{}, err
	}
	return question, nil
}

func (r *questionRepositoryImpl) CreateBatch(ctx context.Context, questions []entity.Question) ([]entity.Question, error) {
	err := r.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i := range questions {
			err := tx.Create(&questions[i]).Error
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return questions, nil
}

func (r *questionRepositoryImpl) Update(ctx context.Context, question entity.Question, options []entity.Answer) (entity.Question, error) {
	err := r.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Update question fields
		err := tx.Model(&question).Where("id = ?", question.ID).Updates(entity.Question{
			Content:   question.Content,
			TimeLimit: question.TimeLimit,
		}).Error
		if err != nil {
			return err
		}

		// 2. Fetch existing options
		var existingOptions []entity.Answer
		err = tx.Where("question_id = ?", question.ID).Find(&existingOptions).Error
		if err != nil {
			return err
		}

		existingMap := make(map[uuid.UUID]entity.Answer)
		for _, o := range existingOptions {
			existingMap[o.ID] = o
		}

		// 3. Save new/updated options
		var updatedIDs []uuid.UUID
		for _, o := range options {
			if o.ID != uuid.Nil {
				err = tx.Model(&entity.Answer{}).Where("id = ?", o.ID).Updates(entity.Answer{
					Content:   o.Content,
					IsCorrect: o.IsCorrect,
				}).Error
				if err != nil {
					return err
				}
				updatedIDs = append(updatedIDs, o.ID)
			} else {
				o.ID = uuid.New()
				o.QuestionID = question.ID
				err = tx.Create(&o).Error
				if err != nil {
					return err
				}
				updatedIDs = append(updatedIDs, o.ID)
			}
		}

		// 4. Delete removed options
		for id := range existingMap {
			found := false
			for _, uid := range updatedIDs {
				if uid == id {
					found = true
					break
				}
			}
			if !found {
				err = tx.Where("id = ?", id).Delete(&entity.Answer{}).Error
				if err != nil {
					return err
				}
			}
		}

		return nil
	})

	if err != nil {
		return entity.Question{}, err
	}

	return r.FindByID(ctx, question.ID)
}

func (r *questionRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.DB.WithContext(ctx).Where("id = ?", id).Delete(&entity.Question{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("question not found or already deleted")
	}
	return nil
}
