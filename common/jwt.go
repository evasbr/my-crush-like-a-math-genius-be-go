// Package common provides cross-cutting utility helper functions
// that can be utilized across all layers of the application.
package common

import (
	"evasbr/mclamg/configuration"
	"evasbr/mclamg/exception"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

// GenerateToken generates a new JSON Web Token (JWT) for user authentication.
// The token includes claims such as user_id, username, user roles,
// user permissions (map[string]interface{}), an expiration timestamp (exp)
// fetched from the environment configurations (.env), and a session ID (sid).
func GenerateToken(userID string, username string, roles []string, permissions map[string]interface{}, config configuration.Config, sid string) string {
	jwtSecret := config.Get("JWT_SECRET_KEY")
	
	jwtExpiredMinutes := 15
	if expStr := config.Get("JWT_EXPIRE_MINUTES_COUNT"); expStr != "" {
		if val, err := strconv.Atoi(expStr); err == nil {
			jwtExpiredMinutes = val
		}
	}

	claims := jwt.MapClaims{
		"token_type":  "access",
		"user_id":     userID,
		"username":    username,
		"roles":       roles,
		"permissions": permissions,
		"sid":         sid,
		"exp":         time.Now().Add(time.Minute * time.Duration(jwtExpiredMinutes)).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenSigned, err := token.SignedString([]byte(jwtSecret))
	exception.PanicLogging(err)

	return tokenSigned
}

// GenerateTokenPair generates both an access token and a refresh token.
// If the passed sid is empty, a new UUID session ID is generated.
func GenerateTokenPair(userID string, username string, roles []string, permissions map[string]interface{}, config configuration.Config, sid string) (string, string, string) {
	jwtSecret := config.Get("JWT_SECRET_KEY")

	// Access token expiration (default 15 minutes)
	jwtExpiredMinutes := 15
	if expStr := config.Get("JWT_EXPIRE_MINUTES_COUNT"); expStr != "" {
		if val, err := strconv.Atoi(expStr); err == nil {
			jwtExpiredMinutes = val
		}
	}

	// Refresh token expiration (default 7 days / 10080 minutes)
	refreshExpiredMinutes := 10080
	if refExpStr := config.Get("JWT_REFRESH_EXPIRE_MINUTES_COUNT"); refExpStr != "" {
		if val, err := strconv.Atoi(refExpStr); err == nil {
			refreshExpiredMinutes = val
		}
	}

	if sid == "" {
		sid = uuid.New().String()
	}

	// 1. Generate Access Token
	accessClaims := jwt.MapClaims{
		"token_type":  "access",
		"user_id":     userID,
		"username":    username,
		"roles":       roles,
		"permissions": permissions,
		"sid":         sid,
		"exp":         time.Now().Add(time.Minute * time.Duration(jwtExpiredMinutes)).Unix(),
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenSigned, err := accessToken.SignedString([]byte(jwtSecret))
	exception.PanicLogging(err)

	// 2. Generate Refresh Token
	refreshClaims := jwt.MapClaims{
		"token_type": "refresh",
		"user_id":    userID,
		"username":   username,
		"sid":        sid,
		"exp":        time.Now().Add(time.Minute * time.Duration(refreshExpiredMinutes)).Unix(),
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenSigned, err := refreshToken.SignedString([]byte(jwtSecret))
	exception.PanicLogging(err)

	return accessTokenSigned, refreshTokenSigned, sid
}
