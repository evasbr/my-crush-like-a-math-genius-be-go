package common

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractCloudinaryPublicID(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{
			name:     "Standard Cloudinary URL with version",
			url:      "https://res.cloudinary.com/demo/image/upload/v1570975253/sample.jpg",
			expected: "sample",
		},
		{
			name:     "Cloudinary URL with subfolder and version",
			url:      "http://res.cloudinary.com/demo/image/upload/v99999999/profile_pics/user_123.png",
			expected: "profile_pics/user_123",
		},
		{
			name:     "Cloudinary URL without version",
			url:      "https://res.cloudinary.com/demo/image/upload/profile_pics/user_456.webp",
			expected: "profile_pics/user_456",
		},
		{
			name:     "Cloudinary URL with nested folders",
			url:      "https://res.cloudinary.com/demo/image/upload/v1/a/b/c/d.jpeg",
			expected: "a/b/c/d",
		},
		{
			name:     "Non-cloudinary URL",
			url:      "https://google.com/image.png",
			expected: "",
		},
		{
			name:     "Empty URL",
			url:      "",
			expected: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := ExtractCloudinaryPublicID(tc.url)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestDisabledStorage(t *testing.T) {
	storage := NewDisabledStorage()
	ctx := context.Background()

	// UploadFile should return error
	url, err := storage.UploadFile(ctx, strings.NewReader(""), "test.jpg", StorageFolder("test_folder"))
	assert.Error(t, err)
	assert.Equal(t, "", url)
	assert.Contains(t, err.Error(), "missing CLOUDINARY_URL")

	// DeleteFile should return error
	err = storage.DeleteFile(ctx, "https://res.cloudinary.com/demo/image/upload/v1/sample.jpg")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing CLOUDINARY_URL")
}

func TestNewFileStorage(t *testing.T) {
	t.Run("Empty Config returns DisabledStorage", func(t *testing.T) {
		storage, err := NewFileStorage("")
		assert.NoError(t, err)
		_, ok := storage.(*disabledStorage)
		assert.True(t, ok)
	})

	t.Run("Malformed Cloudinary URL returns initialization error", func(t *testing.T) {
		storage, err := NewFileStorage("cloudinary://invalid%%")
		assert.Error(t, err)
		assert.Nil(t, storage)
	})
}
