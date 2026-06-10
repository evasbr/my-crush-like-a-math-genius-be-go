package model

type CreateAnswerOptionDto struct {
	Content   string `json:"content" validate:"required"`
	IsCorrect *bool  `json:"isCorrect" validate:"required"`
}

type AnswerOptionDto struct {
	ID        *string `json:"id,omitempty"`
	Content   string  `json:"content" validate:"required"`
	IsCorrect *bool   `json:"isCorrect,omitempty"`
}

type CreateQuestionBatchItem struct {
	Content   string                  `json:"content" validate:"required"`
	TimeLimit int                     `json:"timeLimit" validate:"required,min=1"`
	Options   []CreateAnswerOptionDto `json:"options" validate:"required,min=1,dive"`
}

type CreateQuestionBatchRequest struct {
	TopicID   string                    `json:"topicId" validate:"required,uuid"`
	Level     string                    `json:"level" validate:"required,oneof=easy medium hard"`
	Questions []CreateQuestionBatchItem `json:"questions" validate:"required,min=1,dive"`
}

type UpdateQuestionRequest struct {
	Content   string            `json:"content" validate:"required"`
	TimeLimit int               `json:"timeLimit" validate:"required,min=1"`
	Options   []AnswerOptionDto `json:"options" validate:"required,min=1,dive"`
}

type QuestionResponse struct {
	ID        string            `json:"id"`
	TopicID   string            `json:"topicId"`
	Content   string            `json:"content"`
	Level     string            `json:"level"`
	TimeLimit int               `json:"timeLimit"`
	Options   []AnswerOptionDto `json:"options"`
	CreatedAt string            `json:"createdAt"`
	UpdatedAt string            `json:"updatedAt"`
}

type QuestionFilter struct {
	TopicID string `query:"topicId"`
	Page    int    `query:"page"`
	Limit   int    `query:"limit"`
}
