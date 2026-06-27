package model

type TopicAttemptLevelSetting struct {
	Level              string `json:"level"`
	TrueScore          int    `json:"true_score"`
	FalseScore         int    `json:"false_score"`
	TotalQuestions     int    `json:"total_questions"`
	RemainingQuestions int    `json:"remaining_questions"`
}

type TopicAttemptInfoResponse struct {
	MaxAttempts       int                        `json:"max_attempts"`
	CurrentAttempts   int                        `json:"current_attempts"`
	RemainingAttempts int                        `json:"remaining_attempts"`
	LevelSettings     []TopicAttemptLevelSetting `json:"level_settings"`
}

type StartAttemptRequest struct {
	TopicID            string `json:"topicId" validate:"required,uuid"`
	Level              string `json:"level" validate:"required,oneof=easy medium hard"`
	RequestedQuestions int    `json:"requestedQuestions" validate:"required,min=1"`
}

type SubmitAnswerRequest struct {
	AttemptSessionID string `json:"attemptSessionId" validate:"required,uuid"`
	QuestionID       string `json:"questionId" validate:"required,uuid"`
	AnswerID         string `json:"answerId" validate:"omitempty,uuid"`
}

type SubmitAnswerResponse struct {
	IsCorrect       bool   `json:"isCorrect"`
	CorrectAnswerID string `json:"correctAnswerId"`
	IsFinished      bool   `json:"isFinished"`
	Score           *int   `json:"score,omitempty"`
}

type ActiveQuestionResponse struct {
	IsFinished bool              `json:"isFinished"`
	Question   *QuestionResponse `json:"question,omitempty"`
}

type AttemptDetailDto struct {
	QuestionID       string            `json:"questionId"`
	QuestionContent  string            `json:"questionContent"`
	Options          []AnswerOptionDto `json:"options"`
	SelectedAnswerID *string           `json:"selectedAnswerId,omitempty"`
	IsCorrect        *bool             `json:"isCorrect,omitempty"`
	AnsweredAt       *string           `json:"answeredAt,omitempty"`
}

type AttemptSessionResponse struct {
	ID                 string             `json:"id"`
	UserID             string             `json:"userId"`
	TopicID            string             `json:"topicId"`
	TopicName          string             `json:"topicName"`
	SelectedLevel      string             `json:"selectedLevel"`
	RequestedQuestions int                `json:"requestedQuestions"`
	Score              *int               `json:"score,omitempty"`
	Status             string             `json:"status"`
	StartedAt          string             `json:"startedAt"`
	FinishedAt         *string            `json:"finishedAt,omitempty"`
	ExpiresAt          string             `json:"expiresAt"`
	Details            []AttemptDetailDto `json:"details,omitempty"`
}

type AttemptFilter struct {
	TopicID string `query:"topicId"`
	Page    int    `query:"page"`
	Limit   int    `query:"limit"`
}
