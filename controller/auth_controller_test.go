package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"evasbr/mclamg/entity"
	"evasbr/mclamg/middleware"
	"evasbr/mclamg/model"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestRegisterUser(t *testing.T) {
	// Clean up users first
	deleteAllUsers()

	lastName := "User"
	gender := "male"
	registerRequest := model.RegisterRequest{
		Username:  "newuser",
		Email:     "newuser@example.com",
		FirstName: "New",
		LastName:  &lastName,
		Gender:    &gender,
		Password:  "newpassword",
	}
	body, _ := json.Marshal(registerRequest)

	request := httptest.NewRequest("POST", "/authentication/register", bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")

	response, _ := appTest.Test(request)
	assert.Equal(t, 201, response.StatusCode)

	responseBody, _ := io.ReadAll(response.Body)
	var webResponse model.GeneralResponse
	_ = json.Unmarshal(responseBody, &webResponse)

	assert.Equal(t, 201, webResponse.Code)
	assert.Equal(t, "Success", webResponse.Message)

	dataMap := webResponse.Data.(map[string]interface{})
	assert.Equal(t, "newuser", dataMap["username"])

	// Try registering the same user again -> expect 400 Bad Request
	dupRequest := httptest.NewRequest("POST", "/authentication/register", bytes.NewBuffer(body))
	dupRequest.Header.Set("Content-Type", "application/json")
	dupRequest.Header.Set("Accept", "application/json")
	responseFail, _ := appTest.Test(dupRequest)
	assert.Equal(t, 400, responseFail.StatusCode)
}

func TestAuthenticationAndGetProfile(t *testing.T) {
	// 1. Authenticate / login
	tokenResponse := authenticationCreate()
	token := tokenResponse["access_token"].(string)
	assert.NotEmpty(t, token)

	// 2. Fetch profile /users/me
	request := httptest.NewRequest("GET", "/users/me", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Accept", "application/json")

	response, _ := appTest.Test(request)
	assert.Equal(t, 200, response.StatusCode)

	responseBody, _ := io.ReadAll(response.Body)
	var webResponse model.GeneralResponse
	_ = json.Unmarshal(responseBody, &webResponse)

	assert.Equal(t, 200, webResponse.Code)
	assert.Equal(t, "Success", webResponse.Message)

	profileData := webResponse.Data.(map[string]interface{})
	assert.Equal(t, "admin", profileData["username"])
	assert.Equal(t, "admin@example.com", profileData["email"])
	assert.Equal(t, "active", profileData["status"])
}

func TestGetProfileUnauthorized(t *testing.T) {
	request := httptest.NewRequest("GET", "/users/me", nil)
	request.Header.Set("Accept", "application/json")

	response, _ := appTest.Test(request)
	assert.Equal(t, 401, response.StatusCode)
}

func TestAdminDashboardRBAC(t *testing.T) {
	// Configure RBAC mode
	os.Setenv("AUTH_MODE", "RBAC")

	// 1. Admin login (has ROLE_ADMIN but dashboard route requires admin or superadmin)
	tokenResponse := authenticationCreate()
	token := tokenResponse["access_token"].(string)

	request := httptest.NewRequest("GET", "/admin/dashboard", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Accept", "application/json")

	response, _ := appTest.Test(request)
	assert.Equal(t, 403, response.StatusCode) // admin doesn't match allowedRoles []string{"admin", "superadmin"} - it matches ROLE_ADMIN but dashboard requires admin.
}

func TestAdminDashboardPBAC(t *testing.T) {
	// Configure PBAC mode
	os.Setenv("AUTH_MODE", "PBAC")
	defer os.Setenv("AUTH_MODE", "RBAC") // Reset back

	// 1. Admin login (ROLE_ADMIN has write:product, read:product, read:profile, but NOT read:dashboard)
	tokenResponse := authenticationCreate()
	token := tokenResponse["access_token"].(string)

	request := httptest.NewRequest("GET", "/admin/dashboard", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Accept", "application/json")

	response, _ := appTest.Test(request)
	assert.Equal(t, 403, response.StatusCode)
}

func TestRequireAuthNilValidation(t *testing.T) {
	assert.Panics(t, func() {
		// This should panic at initialization
		_ = middleware.RequireAuth(nil, config, redisClient)
	})
}

func TestAuthenticationCookiesAndLogout(t *testing.T) {
	// 1. Setup user in DB
	deleteAllUsers()
	password, _ := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)
	createTestUser("admin", "admin@example.com", "male", "admin", "admin", string(password), []string{"22b38d48-8605-4e1f-8630-2c2120fbd682"})

	// 2. Perform authentication request
	loginBody, _ := json.Marshal(model.LoginRequest{
		Identifier: "admin",
		Password:   "admin",
	})
	request := httptest.NewRequest("POST", "/authentication/login", bytes.NewBuffer(loginBody))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")

	response, _ := appTest.Test(request)
	assert.Equal(t, 200, response.StatusCode)

	// Verify httpOnly cookies are present in response
	cookies := response.Cookies()
	var accessTokenCookie, refreshTokenCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "access_token" {
			accessTokenCookie = c
		} else if c.Name == "refresh_token" {
			refreshTokenCookie = c
		}
	}
	assert.NotNil(t, accessTokenCookie)
	assert.NotNil(t, refreshTokenCookie)
	assert.True(t, accessTokenCookie.HttpOnly)
	assert.True(t, refreshTokenCookie.HttpOnly)

	// Verify whitelisted in Redis
	existsAccess, _ := redisClient.Exists(context.Background(), "whitelist:access_token:"+accessTokenCookie.Value).Result()
	existsRefresh, _ := redisClient.Exists(context.Background(), "whitelist:refresh_token:"+refreshTokenCookie.Value).Result()
	assert.Equal(t, int64(1), existsAccess)
	assert.Equal(t, int64(1), existsRefresh)

	// 3. Test accessing /users/me using the cookie
	meReq := httptest.NewRequest("GET", "/users/me", nil)
	meReq.AddCookie(accessTokenCookie)
	meReq.Header.Set("Accept", "application/json")
	meResp, _ := appTest.Test(meReq)
	assert.Equal(t, 200, meResp.StatusCode)

	// 4. Test Logout
	logoutReq := httptest.NewRequest("POST", "/authentication/logout", nil)
	logoutReq.AddCookie(accessTokenCookie)
	logoutReq.AddCookie(refreshTokenCookie)
	logoutResp, _ := appTest.Test(logoutReq)
	assert.Equal(t, 200, logoutResp.StatusCode)

	// Verify cookies are cleared (Expires set to past)
	logoutCookies := logoutResp.Cookies()
	var logoutAccessCleared, logoutRefreshCleared bool
	for _, c := range logoutCookies {
		if c.Name == "access_token" && c.Value == "" {
			logoutAccessCleared = true
		} else if c.Name == "refresh_token" && c.Value == "" {
			logoutRefreshCleared = true
		}
	}
	assert.True(t, logoutAccessCleared)
	assert.True(t, logoutRefreshCleared)

	// Verify deleted from Redis whitelist
	existsAccessPost, _ := redisClient.Exists(context.Background(), "whitelist:access_token:"+accessTokenCookie.Value).Result()
	existsRefreshPost, _ := redisClient.Exists(context.Background(), "whitelist:refresh_token:"+refreshTokenCookie.Value).Result()
	assert.Equal(t, int64(0), existsAccessPost)
	assert.Equal(t, int64(0), existsRefreshPost)

	// 5. Test accessing /users/me again with the cookie -> should fail
	failReq := httptest.NewRequest("GET", "/users/me", nil)
	failReq.AddCookie(accessTokenCookie)
	failReq.Header.Set("Accept", "application/json")
	failResp, _ := appTest.Test(failReq)
	assert.Equal(t, 401, failResp.StatusCode)
}

func TestAccessTokenExpirationAndRefreshWithRotation(t *testing.T) {
	// 1. Setup user in DB
	deleteAllUsers()
	password, _ := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)
	createTestUser("admin", "admin@example.com", "male", "admin", "admin", string(password), []string{"22b38d48-8605-4e1f-8630-2c2120fbd682"})

	// Create user roles, find by ID, retrieve fields for manually building expired token
	authResult, _ := authRepository.FindAuthentication(context.Background(), "admin", []string{string(entity.MethodLocalUsername)})
	userResult := authResult.User

	// 2. Manually generate expired access token and valid refresh token
	jwtSecret := config.Get("JWT_SECRET_KEY")
	expiredAccessClaims := jwt.MapClaims{
		"token_type":  "access",
		"user_id":     userResult.ID.String(),
		"username":    userResult.Username,
		"roles":       []string{"ROLE_ADMIN"},
		"permissions": map[string]interface{}{"PROFILE": []string{"read:profile"}},
		"sid":         "test-session-id-rotation",
		"exp":         time.Now().Add(-time.Hour).Unix(), // Expired
	}
	expiredToken := jwt.NewWithClaims(jwt.SigningMethodHS256, expiredAccessClaims)
	expiredAccessTokenStr, _ := expiredToken.SignedString([]byte(jwtSecret))

	// Refresh token expires in 24 hours (less than 3 days threshold -> should rotate!)
	validRefreshClaims := jwt.MapClaims{
		"token_type": "refresh",
		"user_id":    userResult.ID.String(),
		"username":   userResult.Username,
		"sid":        "test-session-id-rotation",
		"exp":        time.Now().Add(time.Hour * 24).Unix(), // 24 hours
	}
	validRefreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, validRefreshClaims)
	validRefreshTokenStr, _ := validRefreshToken.SignedString([]byte(jwtSecret))

	// Whitelist both in Redis
	redisClient.Set(context.Background(), "whitelist:access_token:"+expiredAccessTokenStr, "valid", time.Hour)
	redisClient.Set(context.Background(), "whitelist:refresh_token:"+validRefreshTokenStr, "valid", time.Hour)
	redisClient.SAdd(context.Background(), "session_tokens:test-session-id-rotation", "whitelist:access_token:"+expiredAccessTokenStr, "whitelist:refresh_token:"+validRefreshTokenStr)

	// 3. Request /users/me using expired token cookie -> Expect 401 with EXPIRED_ACCESS_TOKEN
	request := httptest.NewRequest("GET", "/users/me", nil)
	request.AddCookie(&http.Cookie{Name: "access_token", Value: expiredAccessTokenStr})
	request.Header.Set("Accept", "application/json")
	response, _ := appTest.Test(request)
	assert.Equal(t, 401, response.StatusCode)

	responseBody, _ := io.ReadAll(response.Body)
	var webResponse model.GeneralResponse
	_ = json.Unmarshal(responseBody, &webResponse)
	assert.Equal(t, "Access token expired", webResponse.Message)
	assert.Equal(t, "EXPIRED_ACCESS_TOKEN", webResponse.Data)

	// 4. Request /authentication/refresh to get new access token
	refreshReq := httptest.NewRequest("POST", "/authentication/refresh", nil)
	refreshReq.AddCookie(&http.Cookie{Name: "refresh_token", Value: validRefreshTokenStr})
	refreshReq.Header.Set("Accept", "application/json")
	refreshResp, _ := appTest.Test(refreshReq)
	assert.Equal(t, 200, refreshResp.StatusCode)

	// Verify both new access token and refresh token cookies are returned (rotation)
	cookies := refreshResp.Cookies()
	var newAccessTokenCookie, newRefreshTokenCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "access_token" {
			newAccessTokenCookie = c
		} else if c.Name == "refresh_token" {
			newRefreshTokenCookie = c
		}
	}
	assert.NotNil(t, newAccessTokenCookie)
	assert.NotEmpty(t, newAccessTokenCookie.Value)
	assert.NotNil(t, newRefreshTokenCookie)
	assert.NotEmpty(t, newRefreshTokenCookie.Value)
	assert.NotEqual(t, validRefreshTokenStr, newRefreshTokenCookie.Value)

	// Verify old refresh token is deleted from Redis
	existsOldRefresh, _ := redisClient.Exists(context.Background(), "whitelist:refresh_token:"+validRefreshTokenStr).Result()
	assert.Equal(t, int64(0), existsOldRefresh)

	// Verify new tokens are whitelisted in Redis
	existsNewAccess, _ := redisClient.Exists(context.Background(), "whitelist:access_token:"+newAccessTokenCookie.Value).Result()
	assert.Equal(t, int64(1), existsNewAccess)
	existsNewRefresh, _ := redisClient.Exists(context.Background(), "whitelist:refresh_token:"+newRefreshTokenCookie.Value).Result()
	assert.Equal(t, int64(1), existsNewRefresh)

	// 5. Query /users/me with new access token cookie -> Expect 200 OK
	meReq := httptest.NewRequest("GET", "/users/me", nil)
	meReq.AddCookie(newAccessTokenCookie)
	meReq.Header.Set("Accept", "application/json")
	meResp, _ := appTest.Test(meReq)
	assert.Equal(t, 200, meResp.StatusCode)
}

func TestAccessTokenExpirationAndRefreshNoRotation(t *testing.T) {
	// 1. Setup user in DB
	deleteAllUsers()
	password, _ := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)
	createTestUser("admin", "admin@example.com", "male", "admin", "admin", string(password), []string{"22b38d48-8605-4e1f-8630-2c2120fbd682"})

	// Create user roles, find by ID, retrieve fields for manually building expired token
	authResult, _ := authRepository.FindAuthentication(context.Background(), "admin", []string{string(entity.MethodLocalUsername)})
	userResult := authResult.User

	// 2. Manually generate expired access token and valid refresh token
	jwtSecret := config.Get("JWT_SECRET_KEY")
	expiredAccessClaims := jwt.MapClaims{
		"token_type":  "access",
		"user_id":     userResult.ID.String(),
		"username":    userResult.Username,
		"roles":       []string{"ROLE_ADMIN"},
		"permissions": map[string]interface{}{"PROFILE": []string{"read:profile"}},
		"sid":         "test-session-id-no-rotation",
		"exp":         time.Now().Add(-time.Hour).Unix(), // Expired
	}
	expiredToken := jwt.NewWithClaims(jwt.SigningMethodHS256, expiredAccessClaims)
	expiredAccessTokenStr, _ := expiredToken.SignedString([]byte(jwtSecret))

	// Refresh token expires in 6 days (more than 3 days threshold -> should NOT rotate!)
	validRefreshClaims := jwt.MapClaims{
		"token_type": "refresh",
		"user_id":    userResult.ID.String(),
		"username":   userResult.Username,
		"sid":        "test-session-id-no-rotation",
		"exp":        time.Now().Add(time.Hour * 24 * 6).Unix(), // 6 days
	}
	validRefreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, validRefreshClaims)
	validRefreshTokenStr, _ := validRefreshToken.SignedString([]byte(jwtSecret))

	// Whitelist both in Redis
	redisClient.Set(context.Background(), "whitelist:access_token:"+expiredAccessTokenStr, "valid", time.Hour)
	redisClient.Set(context.Background(), "whitelist:refresh_token:"+validRefreshTokenStr, "valid", time.Hour)
	redisClient.SAdd(context.Background(), "session_tokens:test-session-id-no-rotation", "whitelist:access_token:"+expiredAccessTokenStr, "whitelist:refresh_token:"+validRefreshTokenStr)

	// 3. Request /authentication/refresh to get new access token
	refreshReq := httptest.NewRequest("POST", "/authentication/refresh", nil)
	refreshReq.AddCookie(&http.Cookie{Name: "refresh_token", Value: validRefreshTokenStr})
	refreshReq.Header.Set("Accept", "application/json")
	refreshResp, _ := appTest.Test(refreshReq)
	assert.Equal(t, 200, refreshResp.StatusCode)

	// Verify only new access token is returned, no new refresh token (no rotation)
	cookies := refreshResp.Cookies()
	var newAccessTokenCookie, newRefreshTokenCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "access_token" {
			newAccessTokenCookie = c
		} else if c.Name == "refresh_token" {
			newRefreshTokenCookie = c
		}
	}
	assert.NotNil(t, newAccessTokenCookie)
	assert.NotEmpty(t, newAccessTokenCookie.Value)
	assert.Nil(t, newRefreshTokenCookie) // Should not rotate!

	// Verify old refresh token is STILL whitelisted in Redis
	existsOldRefresh, _ := redisClient.Exists(context.Background(), "whitelist:refresh_token:"+validRefreshTokenStr).Result()
	assert.Equal(t, int64(1), existsOldRefresh)

	// Verify new access token is whitelisted in Redis
	existsNewAccess, _ := redisClient.Exists(context.Background(), "whitelist:access_token:"+newAccessTokenCookie.Value).Result()
	assert.Equal(t, int64(1), existsNewAccess)
}

func TestRegisterUserValidationError(t *testing.T) {
	deleteAllUsers()

	// Short password and invalid email
	lastName := "User"
	gender := "male"
	registerRequest := model.RegisterRequest{
		Username:  "nu", // too short (min=3)
		Email:     "invalid-email",
		FirstName: "New",
		LastName:  &lastName,
		Gender:    &gender,
		Password:  "123", // too short (min=6)
	}
	body, _ := json.Marshal(registerRequest)

	request := httptest.NewRequest("POST", "/authentication/register", bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")

	response, _ := appTest.Test(request)
	assert.Equal(t, 400, response.StatusCode)

	responseBody, _ := io.ReadAll(response.Body)
	var webResponse model.GeneralResponse
	_ = json.Unmarshal(responseBody, &webResponse)

	assert.Equal(t, 400, webResponse.Code)
	assert.Equal(t, "Bad Request", webResponse.Message)
}

func TestLoginValidationError(t *testing.T) {
	loginRequest := model.LoginRequest{
		Identifier: "", // empty
		Password:   "", // empty
	}
	body, _ := json.Marshal(loginRequest)

	request := httptest.NewRequest("POST", "/authentication/login", bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")

	response, _ := appTest.Test(request)
	assert.Equal(t, 400, response.StatusCode)

	responseBody, _ := io.ReadAll(response.Body)
	var webResponse model.GeneralResponse
	_ = json.Unmarshal(responseBody, &webResponse)

	assert.Equal(t, 400, webResponse.Code)
	assert.Equal(t, "Bad Request", webResponse.Message)
}

func TestRefreshTokenWithAccessToken(t *testing.T) {
	// 1. Setup user and login
	tokenResponse := authenticationCreate()
	accessToken := tokenResponse["access_token"].(string)

	// 2. Try to refresh using access token
	refreshReq := httptest.NewRequest("POST", "/authentication/refresh", nil)
	refreshReq.Header.Set("Authorization", "Bearer "+accessToken)
	refreshReq.Header.Set("Accept", "application/json")
	refreshResp, _ := appTest.Test(refreshReq)

	// Expect unauthorized or bad request because access token is not a refresh token
	assert.Equal(t, 401, refreshResp.StatusCode)

	responseBody, _ := io.ReadAll(refreshResp.Body)
	var webResponse model.GeneralResponse
	_ = json.Unmarshal(responseBody, &webResponse)
	assert.Equal(t, 401, webResponse.Code)
	assert.Equal(t, "Unauthorized", webResponse.Message)
}

func TestProtectedEndpointWithRefreshToken(t *testing.T) {
	// 1. Setup user and login
	tokenResponse := authenticationCreate()
	refreshToken := tokenResponse["refresh_token"].(string)

	// 2. Try to access logout using refresh token
	logoutReq := httptest.NewRequest("POST", "/authentication/logout", nil)
	logoutReq.Header.Set("Authorization", "Bearer "+refreshToken)
	logoutReq.Header.Set("Accept", "application/json")
	logoutResp, _ := appTest.Test(logoutReq)

	// Expect unauthorized because refresh token is not allowed in protected route
	assert.Equal(t, 401, logoutResp.StatusCode)

	responseBody, _ := io.ReadAll(logoutResp.Body)
	var webResponse model.GeneralResponse
	_ = json.Unmarshal(responseBody, &webResponse)
	assert.Equal(t, 401, webResponse.Code)
	assert.Equal(t, "Unauthorized", webResponse.Message)
}
