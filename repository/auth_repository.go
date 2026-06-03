package repository

import (
	"context"
	"evasbr/mclamg/entity"
)

type RegisterUserPayload struct {
	Username          *string
	Email             string
	FirstName         string
	LastName          *string
	Gender            *string
	Password          string // hashed password
	RoleIDs           []string
}

type AuthRepository interface {
	Register(ctx context.Context, payload RegisterUserPayload) (entity.User, error)
	FindAuthentication(ctx context.Context, providerUserID string, methods []string) (entity.Authentication, error)
}
