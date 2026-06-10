package model

type LevelSettingDto struct {
	Level      string `json:"level" validate:"required,oneof=easy medium hard"`
	TrueScore  int    `json:"true_score"`
	FalseScore int    `json:"false_score"`
}

type CreateTopicRequest struct {
	Name          string            `json:"name" validate:"required"`
	LevelSettings []LevelSettingDto `json:"level_settings" validate:"required,dive"`
	MaxAttempts   int               `json:"max_attempts" validate:"required,min=1"`
}

type UpdateTopicRequest struct {
	Name          string            `json:"name" validate:"required"`
	LevelSettings []LevelSettingDto `json:"level_settings" validate:"required,dive"`
	MaxAttempts   int               `json:"max_attempts" validate:"required,min=1"`
}

type TopicResponse struct {
	ID            string            `json:"id"`
	Name          string            `json:"name"`
	LevelSettings []LevelSettingDto `json:"level_settings"`
	MaxAttempts   int               `json:"max_attempts"`
	CreatedAt     string            `json:"created_at"`
	UpdatedAt     string            `json:"updated_at"`
}

type TopicFilter struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
}
