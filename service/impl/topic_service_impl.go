package impl

import (
	"context"
	"fmt"
	"mime/multipart"
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
	Storage         common.FileStorage
	log             *logrus.Entry
}

func NewTopicServiceImpl(topicRepository *repository.TopicRepository, storage common.FileStorage) service.TopicService {
	return &topicServiceImpl{
		TopicRepository: *topicRepository,
		Storage:         storage,
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

func (s *topicServiceImpl) Create(ctx context.Context, request model.CreateTopicRequest, femaleNormalHeader, maleNormalHeader, femaleDatingHeader, maleDatingHeader *multipart.FileHeader) (response model.TopicResponse, err error) {
	common.Validate(request)

	classroomUUID, err := uuid.Parse(request.ClassroomID)
	if err != nil {
		return model.TopicResponse{}, exception.ValidationError{
			Message: "invalid classroom ID format",
		}
	}

	if femaleNormalHeader == nil || maleNormalHeader == nil || femaleDatingHeader == nil || maleDatingHeader == nil {
		return model.TopicResponse{}, exception.ValidationError{
			Message: "all 4 character images (female_normal_img, male_normal_img, female_dating_img, male_dating_img) are required on creation",
		}
	}

	topicID := uuid.New()
	var uploaded []string
	defer func() {
		if err != nil {
			for _, u := range uploaded {
				_ = s.Storage.DeleteFile(ctx, u)
			}
		}
	}()

	uploadFn := func(header *multipart.FileHeader, suffix string) (string, error) {
		file, e := header.Open()
		if e != nil {
			return "", exception.ValidationError{
				Message: fmt.Sprintf("unable to open %s image: %v", suffix, e),
			}
		}
		defer file.Close()

		filename := fmt.Sprintf("%s_%s_%d", topicID.String(), suffix, time.Now().Unix())
		url, e := s.Storage.UploadFile(ctx, file, filename, common.FolderTopics)
		if e != nil {
			return "", e
		}
		uploaded = append(uploaded, url)
		return url, nil
	}

	femaleNormalImg, err := uploadFn(femaleNormalHeader, "female_normal")
	if err != nil {
		return model.TopicResponse{}, err
	}

	maleNormalImg, err := uploadFn(maleNormalHeader, "male_normal")
	if err != nil {
		return model.TopicResponse{}, err
	}

	femaleDatingImg, err := uploadFn(femaleDatingHeader, "female_dating")
	if err != nil {
		return model.TopicResponse{}, err
	}

	maleDatingImg, err := uploadFn(maleDatingHeader, "male_dating")
	if err != nil {
		return model.TopicResponse{}, err
	}

	levelSettings := make([]entity.LevelSetting, len(request.LevelSettings))
	for i, ls := range request.LevelSettings {
		levelSettings[i] = entity.LevelSetting{
			Level:      ls.Level,
			TrueScore:  ls.TrueScore,
			FalseScore: ls.FalseScore,
		}
	}

	topicEntity := entity.Topic{
		ID:                 topicID,
		ClassroomID:        classroomUUID,
		Name:               request.Name,
		Description:        request.Description,
		FemaleNormalImg:    femaleNormalImg,
		MaleNormalImg:      maleNormalImg,
		FemaleDatingImg:    femaleDatingImg,
		MaleDatingImg:      maleDatingImg,
		FemaleNormalDialog: request.FemaleNormalDialog,
		MaleNormalDialog:   request.MaleNormalDialog,
		FemaleDatingDialog: request.FemaleDatingDialog,
		MaleDatingDialog:   request.MaleDatingDialog,
		Status:             request.Status,
		LevelSettings:      levelSettings,
		MaxAttempts:        request.MaxAttempts,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	createdTopic, err := s.TopicRepository.Create(ctx, topicEntity)
	if err != nil {
		return model.TopicResponse{}, err
	}

	return s.toTopicResponse(createdTopic), nil
}

func (s *topicServiceImpl) Update(ctx context.Context, request model.UpdateTopicRequest, femaleNormalHeader, maleNormalHeader, femaleDatingHeader, maleDatingHeader *multipart.FileHeader, id string) (response model.TopicResponse, err error) {
	parsedUUID, err := uuid.Parse(id)
	if err != nil {
		return model.TopicResponse{}, exception.ValidationError{
			Message: "invalid topic ID format",
		}
	}

	common.Validate(request)

	existing, err := s.TopicRepository.FindByID(ctx, parsedUUID)
	if err != nil {
		return model.TopicResponse{}, exception.NotFoundError{
			Message: err.Error(),
		}
	}

	if request.ClassroomID != nil {
		classroomUUID, err := uuid.Parse(*request.ClassroomID)
		if err != nil {
			return model.TopicResponse{}, exception.ValidationError{
				Message: "invalid classroom ID format",
			}
		}
		existing.ClassroomID = classroomUUID
	}

	var newUploaded []string
	var oldToDelete []string

	defer func() {
		if err != nil {
			for _, u := range newUploaded {
				_ = s.Storage.DeleteFile(ctx, u)
			}
		} else {
			for _, u := range oldToDelete {
				_ = s.Storage.DeleteFile(ctx, u)
			}
		}
	}()

	uploadFn := func(header *multipart.FileHeader, suffix string) (string, error) {
		file, e := header.Open()
		if e != nil {
			return "", exception.ValidationError{
				Message: fmt.Sprintf("unable to open %s image: %v", suffix, e),
			}
		}
		defer file.Close()

		filename := fmt.Sprintf("%s_%s_%d", existing.ID.String(), suffix, time.Now().Unix())
		url, e := s.Storage.UploadFile(ctx, file, filename, common.FolderTopics)
		if e != nil {
			return "", e
		}
		newUploaded = append(newUploaded, url)
		return url, nil
	}

	femaleNormalImg := existing.FemaleNormalImg
	if femaleNormalHeader != nil {
		femaleNormalImg, err = uploadFn(femaleNormalHeader, "female_normal")
		if err != nil {
			return model.TopicResponse{}, err
		}
		if existing.FemaleNormalImg != "" {
			oldToDelete = append(oldToDelete, existing.FemaleNormalImg)
		}
	}

	maleNormalImg := existing.MaleNormalImg
	if maleNormalHeader != nil {
		maleNormalImg, err = uploadFn(maleNormalHeader, "male_normal")
		if err != nil {
			return model.TopicResponse{}, err
		}
		if existing.MaleNormalImg != "" {
			oldToDelete = append(oldToDelete, existing.MaleNormalImg)
		}
	}

	femaleDatingImg := existing.FemaleDatingImg
	if femaleDatingHeader != nil {
		femaleDatingImg, err = uploadFn(femaleDatingHeader, "female_dating")
		if err != nil {
			return model.TopicResponse{}, err
		}
		if existing.FemaleDatingImg != "" {
			oldToDelete = append(oldToDelete, existing.FemaleDatingImg)
		}
	}

	maleDatingImg := existing.MaleDatingImg
	if maleDatingHeader != nil {
		maleDatingImg, err = uploadFn(maleDatingHeader, "male_dating")
		if err != nil {
			return model.TopicResponse{}, err
		}
		if existing.MaleDatingImg != "" {
			oldToDelete = append(oldToDelete, existing.MaleDatingImg)
		}
	}

	if request.Name != nil {
		existing.Name = *request.Name
	}
	if request.Description != nil {
		existing.Description = request.Description
	}
	if request.FemaleNormalDialog != nil {
		existing.FemaleNormalDialog = *request.FemaleNormalDialog
	}
	if request.MaleNormalDialog != nil {
		existing.MaleNormalDialog = *request.MaleNormalDialog
	}
	if request.FemaleDatingDialog != nil {
		existing.FemaleDatingDialog = *request.FemaleDatingDialog
	}
	if request.MaleDatingDialog != nil {
		existing.MaleDatingDialog = *request.MaleDatingDialog
	}
	if request.Status != nil {
		existing.Status = *request.Status
	}
	if request.MaxAttempts != nil {
		existing.MaxAttempts = *request.MaxAttempts
	}

	if request.LevelSettings != nil {
		levelSettings := make([]entity.LevelSetting, len(request.LevelSettings))
		for i, ls := range request.LevelSettings {
			levelSettings[i] = entity.LevelSetting{
				Level:      ls.Level,
				TrueScore:  ls.TrueScore,
				FalseScore: ls.FalseScore,
			}
		}
		existing.LevelSettings = levelSettings
	}

	existing.FemaleNormalImg = femaleNormalImg
	existing.MaleNormalImg = maleNormalImg
	existing.FemaleDatingImg = femaleDatingImg
	existing.MaleDatingImg = maleDatingImg
	existing.UpdatedAt = time.Now()

	updatedTopic, err := s.TopicRepository.Update(ctx, existing)
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

	existing, err := s.TopicRepository.FindByID(ctx, parsedUUID)
	if err != nil {
		return exception.NotFoundError{
			Message: err.Error(),
		}
	}

	err = s.TopicRepository.Delete(ctx, parsedUUID)
	if err != nil {
		return exception.NotFoundError{
			Message: err.Error(),
		}
	}

	// Clean up storage images on successful database delete
	if existing.FemaleNormalImg != "" {
		_ = s.Storage.DeleteFile(ctx, existing.FemaleNormalImg)
	}
	if existing.MaleNormalImg != "" {
		_ = s.Storage.DeleteFile(ctx, existing.MaleNormalImg)
	}
	if existing.FemaleDatingImg != "" {
		_ = s.Storage.DeleteFile(ctx, existing.FemaleDatingImg)
	}
	if existing.MaleDatingImg != "" {
		_ = s.Storage.DeleteFile(ctx, existing.MaleDatingImg)
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
		ID:                 topic.ID.String(),
		ClassroomID:        topic.ClassroomID.String(),
		Name:               topic.Name,
		Description:        topic.Description,
		FemaleNormalImg:    topic.FemaleNormalImg,
		MaleNormalImg:      topic.MaleNormalImg,
		FemaleDatingImg:    topic.FemaleDatingImg,
		MaleDatingImg:      topic.MaleDatingImg,
		FemaleNormalDialog: topic.FemaleNormalDialog,
		MaleNormalDialog:   topic.MaleNormalDialog,
		FemaleDatingDialog: topic.FemaleDatingDialog,
		MaleDatingDialog:   topic.MaleDatingDialog,
		Status:             topic.Status,
		LevelSettings:      levelSettings,
		MaxAttempts:        topic.MaxAttempts,
		CreatedAt:          topic.CreatedAt.Format(time.RFC3339),
		UpdatedAt:          topic.UpdatedAt.Format(time.RFC3339),
	}
}
