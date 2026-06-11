package controller

import (
	"bytes"
	"encoding/json"
	"evasbr/mclamg/model"
	"io"
	"mime/multipart"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpdateProfilePicture(t *testing.T) {
	// 1. Authenticate / login
	tokenResponse := authenticationCreate()
	token := tokenResponse["access_token"].(string)
	assert.NotEmpty(t, token)

	t.Run("Missing File Field", func(t *testing.T) {
		request := httptest.NewRequest("PATCH", "/users/me/profile-picture", nil)
		request.Header.Set("Authorization", "Bearer "+token)
		request.Header.Set("Accept", "application/json")

		response, _ := appTest.Test(request)
		assert.Equal(t, 400, response.StatusCode)

		responseBody, _ := io.ReadAll(response.Body)
		var webResponse model.GeneralResponse
		_ = json.Unmarshal(responseBody, &webResponse)
		assert.Equal(t, 400, webResponse.Code)
		assert.Contains(t, webResponse.Data, "profile_picture field is required")
	})

	t.Run("Invalid File Extension", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		part, err := writer.CreateFormFile("profile_picture", "malicious.exe")
		assert.NoError(t, err)
		_, err = part.Write([]byte("malicious content"))
		assert.NoError(t, err)
		writer.Close()

		request := httptest.NewRequest("PATCH", "/users/me/profile-picture", body)
		request.Header.Set("Authorization", "Bearer "+token)
		request.Header.Set("Content-Type", writer.FormDataContentType())
		request.Header.Set("Accept", "application/json")

		response, _ := appTest.Test(request)
		assert.Equal(t, 400, response.StatusCode)

		responseBody, _ := io.ReadAll(response.Body)
		var webResponse model.GeneralResponse
		_ = json.Unmarshal(responseBody, &webResponse)
		assert.Equal(t, 400, webResponse.Code)
		assert.Contains(t, webResponse.Data.(string), "invalid file extension")
	})
}
