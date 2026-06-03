package model

type UserPublicDto struct {
	Id                string `json:"id"`
	FirstName         string `json:"first_name"`
	LastName          string `json:"last_name"`
	ProfilePictureURL string `json:"profile_picture_url"`
}

type UserDto struct {
	Id                string `json:"id"`
	Email             string `json:"email"`
	FirstName         string `json:"first_name"`
	LastName          string `json:"last_name"`
	Gender            string `json:"gender"`
	ProfilePictureURL string `json:"profile_picture_url"`
	Status            string `json:"status"`
}

type UserFilter struct {
	Page           int
	Limit          int
	IncludeDeleted bool
}

type UpdateUser struct {
	Gender            *string `json:"gender" validate:"omitempty,oneof=male female"`
	FirstName         string  `json:"first_name" validate:"required,min=3,max=50"`
	LastName          *string `json:"last_name" validate:"omitempty,max=50"`
	ProfilePictureURL *string `json:"profile_picture_url" validate:"omitempty,url"`
}

type UserProfileResponse struct {
	Id                string                 `json:"id"`
	Email             string                 `json:"email"`
	Username          string                 `json:"username"`
	FirstName         string                 `json:"first_name"`
	LastName          string                 `json:"last_name"`
	Gender            string                 `json:"gender"`
	ProfilePictureURL string                 `json:"profile_picture_url"`
	Status            string                 `json:"status"`
	Roles             []string               `json:"roles"`
	Permissions       map[string]interface{} `json:"permissions,omitempty"`
}
