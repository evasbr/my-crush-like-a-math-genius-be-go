package impl

import (
	"context"
	"errors"
	"strconv"
	"time"

	"evasbr/mclamg/common"
	"evasbr/mclamg/configuration"
	"evasbr/mclamg/entity"
	"evasbr/mclamg/model"
	"evasbr/mclamg/repository"
	"evasbr/mclamg/service"

	"github.com/go-redis/redis/v9"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

type authServiceImpl struct {
	UserRepository repository.UserRepository
	AuthRepository repository.AuthRepository
	Redis          *redis.Client
	Config         configuration.Config
	log            *logrus.Entry
}

func NewAuthServiceImpl(
	userRepository *repository.UserRepository,
	authRepository *repository.AuthRepository,
	redisClient *redis.Client,
	config configuration.Config,
) service.AuthService {
	return &authServiceImpl{
		UserRepository: *userRepository,
		AuthRepository: *authRepository,
		Redis:          redisClient,
		Config:         config,
		log:            common.Log.WithField("scope", "AuthService"),
	}
}

func (s *authServiceImpl) Register(ctx context.Context, req model.RegisterRequest) (model.RegisterResponse, error) {
	common.Validate(req)

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return model.RegisterResponse{}, err
	}

	payload := repository.RegisterUserPayload{
		Username:  &req.Username,
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Gender:    req.Gender,
		Password:  string(hashedPassword),
		RoleIDs:   []string{"d4df0794-22e8-4a30-9039-bbcd76447b56"},
	}

	user, err := s.AuthRepository.Register(ctx, payload)
	if err != nil {
		return model.RegisterResponse{}, err
	}

	var username string
	if req.Username != "" {
		username = req.Username
	} else {
		username = user.Email
	}

	return model.RegisterResponse{
		ID:                user.ID.String(),
		Username:          username,
		Email:             user.Email,
		FirstName:         user.FirstName,
		LastName:          user.LastName,
		Gender:            user.Gender,
		ProfilePictureURL: user.ProfilePictureURL,
	}, nil
}

func (s *authServiceImpl) Login(ctx context.Context, req model.LoginRequest) (model.LoginResponse, error) {
	common.Validate(req)

	auth, err := s.AuthRepository.FindAuthentication(ctx, req.Identifier, []string{
		string(entity.MethodLocalEmail),
		string(entity.MethodLocalUsername),
	})
	if err != nil {
		return model.LoginResponse{}, err
	}

	if auth.Password == nil {
		return model.LoginResponse{}, errors.New("invalid credentials")
	}

	err = bcrypt.CompareHashAndPassword([]byte(*auth.Password), []byte(req.Password))
	if err != nil {
		return model.LoginResponse{}, errors.New("invalid password")
	}

	user := auth.User

	// Determine username
	user.Username = user.Email
	for _, a := range user.Authentications {
		if a.Method == string(entity.MethodLocalUsername) {
			user.Username = a.ProviderUserID
			break
		}
	}

	var roles []string
	for _, userRole := range user.UserRoles {
		roles = append(roles, userRole.Role.Name)
	}

	var permissionsList []map[string]interface{}
	for _, userRole := range user.UserRoles {
		permissionsList = append(permissionsList, userRole.Role.Permissions)
	}
	permissions := common.MergePermissions(permissionsList)

	accessToken, refreshToken := common.GenerateTokenPair(user.ID.String(), user.Username, roles, permissions, s.Config)

	// Fetch expiration settings
	jwtExpiredMinutes := 15
	if expStr := s.Config.Get("JWT_EXPIRE_MINUTES_COUNT"); expStr != "" {
		if val, err := strconv.Atoi(expStr); err == nil {
			jwtExpiredMinutes = val
		}
	}

	refreshExpiredMinutes := 10080
	if refExpStr := s.Config.Get("JWT_REFRESH_EXPIRE_MINUTES_COUNT"); refExpStr != "" {
		if val, err := strconv.Atoi(refExpStr); err == nil {
			refreshExpiredMinutes = val
		}
	}

	// Store both tokens in Redis whitelist
	err = s.Redis.Set(ctx, "whitelist:token:"+accessToken, "valid", time.Minute*time.Duration(jwtExpiredMinutes)).Err()
	if err != nil {
		return model.LoginResponse{}, err
	}

	err = s.Redis.Set(ctx, "whitelist:token:"+refreshToken, "valid", time.Minute*time.Duration(refreshExpiredMinutes)).Err()
	if err != nil {
		return model.LoginResponse{}, err
	}

	var permissionsResponse map[string]interface{}
	if s.Config.Get("AUTH_MODE") != "RBAC" {
		permissionsResponse = permissions
	}

	var lastName string
	if user.LastName != nil {
		lastName = *user.LastName
	}
	var gender string
	if user.Gender != nil {
		gender = *user.Gender
	}
	var profilePictureURL string
	if user.ProfilePictureURL != nil {
		profilePictureURL = *user.ProfilePictureURL
	}

	userDto := model.UserDto{
		Id:                user.ID.String(),
		Email:             user.Email,
		FirstName:         user.FirstName,
		LastName:          lastName,
		Gender:            gender,
		ProfilePictureURL: profilePictureURL,
		Status:            user.Status,
	}

	return model.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Username:     user.Username,
		User:         userDto,
		Roles:        roles,
		Permissions:  permissionsResponse,
	}, nil
}

func (s *authServiceImpl) RefreshToken(ctx context.Context, refreshTokenStr string) (model.RefreshTokenResponse, error) {
	// Check if Refresh Token exists in Redis whitelist
	exists, rErr := s.Redis.Exists(ctx, "whitelist:token:"+refreshTokenStr).Result()
	if rErr != nil || exists == 0 {
		return model.RefreshTokenResponse{}, errors.New("refresh token is revoked or not whitelisted")
	}

	jwtSecret := s.Config.Get("JWT_SECRET_KEY")
	token, err := jwt.Parse(refreshTokenStr, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})
	if err != nil {
		return model.RefreshTokenResponse{}, err
	}

	if !token.Valid {
		return model.RefreshTokenResponse{}, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return model.RefreshTokenResponse{}, errors.New("invalid claims")
	}

	tokenType, ok := claims["token_type"].(string)
	if !ok || tokenType != "refresh" {
		return model.RefreshTokenResponse{}, errors.New("invalid token type")
	}

	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return model.RefreshTokenResponse{}, errors.New("missing user_id claim")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return model.RefreshTokenResponse{}, errors.New("invalid user_id claim format")
	}

	user, err := s.UserRepository.FindByID(ctx, userID)
	if err != nil {
		return model.RefreshTokenResponse{}, err
	}

	var roles []string
	for _, userRole := range user.UserRoles {
		roles = append(roles, userRole.Role.Name)
	}

	var permissionsList []map[string]interface{}
	for _, userRole := range user.UserRoles {
		permissionsList = append(permissionsList, userRole.Role.Permissions)
	}
	permissions := common.MergePermissions(permissionsList)

	var expTime time.Time
	if expVal, existsVal := claims["exp"]; existsVal {
		switch v := expVal.(type) {
		case float64:
			expTime = time.Unix(int64(v), 0)
		case int64:
			expTime = time.Unix(v, 0)
		}
	}

	remainingDuration := time.Until(expTime)
	threshold := 3 * 24 * time.Hour // 3 days threshold

	jwtExpiredMinutes := 15
	if expStr := s.Config.Get("JWT_EXPIRE_MINUTES_COUNT"); expStr != "" {
		if val, err := strconv.Atoi(expStr); err == nil {
			jwtExpiredMinutes = val
		}
	}

	if remainingDuration < threshold {
		// Rotate both Access Token and Refresh Token
		newAccessToken, newRefreshToken := common.GenerateTokenPair(user.ID.String(), user.Username, roles, permissions, s.Config)

		refreshExpiredMinutes := 10080
		if refExpStr := s.Config.Get("JWT_REFRESH_EXPIRE_MINUTES_COUNT"); refExpStr != "" {
			if val, err := strconv.Atoi(refExpStr); err == nil {
				refreshExpiredMinutes = val
			}
		}

		err = s.Redis.Set(ctx, "whitelist:token:"+newAccessToken, "valid", time.Minute*time.Duration(jwtExpiredMinutes)).Err()
		if err != nil {
			return model.RefreshTokenResponse{}, err
		}

		err = s.Redis.Set(ctx, "whitelist:token:"+newRefreshToken, "valid", time.Minute*time.Duration(refreshExpiredMinutes)).Err()
		if err != nil {
			return model.RefreshTokenResponse{}, err
		}

		// Revoke old refresh token
		s.Redis.Del(ctx, "whitelist:token:"+refreshTokenStr)

		return model.RefreshTokenResponse{
			AccessToken:  newAccessToken,
			RefreshToken: newRefreshToken,
			Rotated:      true,
		}, nil
	}

	// Generate only a new access token
	newAccessToken := common.GenerateToken(user.ID.String(), user.Username, roles, permissions, s.Config)
	err = s.Redis.Set(ctx, "whitelist:token:"+newAccessToken, "valid", time.Minute*time.Duration(jwtExpiredMinutes)).Err()
	if err != nil {
		return model.RefreshTokenResponse{}, err
	}

	return model.RefreshTokenResponse{
		AccessToken: newAccessToken,
		Rotated:     false,
	}, nil
}

func (s *authServiceImpl) Logout(ctx context.Context, accessTokenStr, refreshTokenStr string) error {
	if accessTokenStr != "" {
		s.Redis.Del(ctx, "whitelist:token:"+accessTokenStr)
	}
	if refreshTokenStr != "" {
		s.Redis.Del(ctx, "whitelist:token:"+refreshTokenStr)
	}
	return nil
}


