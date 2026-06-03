package service

import (
	"context"
	"evasbr/mclamg/model"
)

type AuthService interface {
	Login(ctx context.Context, req model.LoginRequest) (model.LoginResponse, error)
	Register(ctx context.Context, req model.RegisterRequest) (model.RegisterResponse, error)
	RefreshToken(ctx context.Context, refreshTokenStr string) (model.RefreshTokenResponse, error)
	Logout(ctx context.Context, sidStr string) error
}
