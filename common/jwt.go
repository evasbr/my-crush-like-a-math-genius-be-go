// Package common provides cross-cutting utility helper functions
// that can be utilized across all layers of the application.
package common

import (
	"evasbr/mclamg/configuration"
	"evasbr/mclamg/exception"
	"github.com/golang-jwt/jwt/v4"
	"strconv"
	"time"
)

// GenerateToken generates a new JSON Web Token (JWT) for user authentication.
// The token includes claims such as username, user roles, and an expiration timestamp (exp)
// fetched from the environment configurations (.env).
//
// Service / Controller Layer Usage Example:
//
//	func (s *authService) Login(ctx context.Context, req model.LoginRequest) string {
//	    // ... verify credentials ...
//	    roles := []map[string]interface{}{{"role": "ROLE_USER"}}
//	    token := common.GenerateToken(req.Username, roles, s.Config)
//	    return token
//	}
func GenerateToken(username string, roles []map[string]interface{}, config configuration.Config) string {
	jwtSecret := config.Get("JWT_SECRET_KEY")
	jwtExpired, err := strconv.Atoi(config.Get("JWT_EXPIRE_MINUTES_COUNT"))
	exception.PanicLogging(err)

	claims := jwt.MapClaims{
		"username": username,
		"roles":    roles,
		"exp":      time.Now().Add(time.Minute * time.Duration(jwtExpired)).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenSigned, err := token.SignedString([]byte(jwtSecret))
	exception.PanicLogging(err)

	return tokenSigned
}
