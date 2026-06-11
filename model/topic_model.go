package model

type LevelSettingDto struct {
	Level      string `json:"level" validate:"required,oneof=easy medium hard"`
	TrueScore  int    `json:"true_score"`
	FalseScore int    `json:"false_score"`
}

type CreateTopicRequest struct {
	ClassroomID        string            `json:"classroom_id" form:"classroom_id" validate:"required,uuid"`
	Name               string            `json:"name" form:"name" validate:"required"`
	Description        *string           `json:"description" form:"description"`
	FemaleNormalImg    string            `json:"female_normal_img" form:"female_normal_img"`
	MaleNormalImg      string            `json:"male_normal_img" form:"male_normal_img"`
	FemaleDatingImg    string            `json:"female_dating_img" form:"female_dating_img"`
	MaleDatingImg      string            `json:"male_dating_img" form:"male_dating_img"`
	FemaleNormalDialog string            `json:"female_normal_dialog" form:"female_normal_dialog" validate:"required"`
	MaleNormalDialog   string            `json:"male_normal_dialog" form:"male_normal_dialog" validate:"required"`
	FemaleDatingDialog string            `json:"female_dating_dialog" form:"female_dating_dialog" validate:"required"`
	MaleDatingDialog   string            `json:"male_dating_dialog" form:"male_dating_dialog" validate:"required"`
	Status             string            `json:"status" form:"status" validate:"required"`
	LevelSettings      []LevelSettingDto `json:"level_settings" form:"level_settings" validate:"required,dive"`
	MaxAttempts        int               `json:"max_attempts" form:"max_attempts" validate:"required,min=1"`
}

type UpdateTopicRequest struct {
	ClassroomID        *string           `json:"classroom_id" form:"classroom_id"`
	Name               *string           `json:"name" form:"name"`
	Description        *string           `json:"description" form:"description"`
	FemaleNormalImg    *string           `json:"female_normal_img" form:"female_normal_img"`
	MaleNormalImg      *string           `json:"male_normal_img" form:"male_normal_img"`
	FemaleDatingImg    *string           `json:"female_dating_img" form:"female_dating_img"`
	MaleDatingImg      *string           `json:"male_dating_img" form:"male_dating_img"`
	FemaleNormalDialog *string           `json:"female_normal_dialog" form:"female_normal_dialog"`
	MaleNormalDialog   *string           `json:"male_normal_dialog" form:"male_normal_dialog"`
	FemaleDatingDialog *string           `json:"female_dating_dialog" form:"female_dating_dialog"`
	MaleDatingDialog   *string           `json:"male_dating_dialog" form:"male_dating_dialog"`
	Status             *string           `json:"status" form:"status"`
	LevelSettings      []LevelSettingDto `json:"level_settings" form:"level_settings"`
	MaxAttempts        *int              `json:"max_attempts" form:"max_attempts"`
}

type TopicResponse struct {
	ID                 string            `json:"id"`
	ClassroomID        string            `json:"classroom_id"`
	Name               string            `json:"name"`
	Description        *string           `json:"description"`
	FemaleNormalImg    string            `json:"female_normal_img"`
	MaleNormalImg      string            `json:"male_normal_img"`
	FemaleDatingImg    string            `json:"female_dating_img"`
	MaleDatingImg      string            `json:"male_dating_img"`
	FemaleNormalDialog string            `json:"female_normal_dialog"`
	MaleNormalDialog   string            `json:"male_normal_dialog"`
	FemaleDatingDialog string            `json:"female_dating_dialog"`
	MaleDatingDialog   string            `json:"male_dating_dialog"`
	Status             string            `json:"status"`
	LevelSettings      []LevelSettingDto `json:"level_settings"`
	MaxAttempts        int               `json:"max_attempts"`
	CreatedAt          string            `json:"created_at"`
	UpdatedAt          string            `json:"updated_at"`
}

type TopicFilter struct {
	Page        int    `query:"page"`
	Limit       int    `query:"limit"`
	ClassroomID string `query:"classroomId"`
}
