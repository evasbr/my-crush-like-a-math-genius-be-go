package model

type RegisterRequest struct {
	Username          string  `json:"username" validate:"required,min=3,max=30"`
	Email             string  `json:"email" validate:"required,email"`
	FirstName         string  `json:"first_name" validate:"required,min=3,max=50"`
	LastName          *string `json:"last_name" validate:"omitempty,max=50"`
	Password          string  `json:"password" validate:"required,min=6,max=50"`
	Gender            *string `json:"gender" validate:"omitempty,oneof=male female"`
	ProfilePictureURL *string `json:"profile_picture_url" validate:"omitempty,url"`
}

type RegisterResponse struct {
	ID                string  `json:"id"`
	Username          string  `json:"username"`
	Email             string  `json:"email"`
	FirstName         string  `json:"first_name"`
	LastName          *string `json:"last_name"`
	Gender            *string `json:"gender"`
	ProfilePictureURL *string `json:"profile_picture_url"`
}

type LoginRequest struct {
	Identifier string `json:"identifier" validate:"required"`
	Password   string `json:"password" validate:"required"`
}

type LoginResponse struct {
	AccessToken  string                 `json:"access_token"`
	RefreshToken string                 `json:"refresh_token"`
	Username     string                 `json:"username"`
	User         UserDto                `json:"user"`
	Roles        []string               `json:"roles"`
	Permissions  map[string]interface{} `json:"permissions,omitempty"`
}

type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Rotated      bool   `json:"rotated"`
}
