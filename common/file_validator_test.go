package common

import (
	"bytes"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"testing"

	"github.com/stretchr/testify/assert"
)

func createMultipartFileHeader(filename string, contentType string, content []byte) (*multipart.FileHeader, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="file"; filename="%s"`, filename))
	h.Set("Content-Type", contentType)

	part, err := writer.CreatePart(h)
	if err != nil {
		return nil, err
	}
	_, err = part.Write(content)
	if err != nil {
		return nil, err
	}
	err = writer.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", "/", body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	err = req.ParseMultipartForm(10 * 1024 * 1024)
	if err != nil {
		return nil, err
	}

	files := req.MultipartForm.File["file"]
	if len(files) == 0 {
		return nil, errors.New("no files parsed")
	}
	return files[0], nil
}

func TestValidateImageFile(t *testing.T) {
	// Magic bytes for JPEG/PNG
	jpegBytes := []byte{0xff, 0xd8, 0xff, 0xe0, 0x00, 0x10, 0x4a, 0x46, 0x49, 0x46, 0x00, 0x01}
	pngBytes := []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a}
	plainTextBytes := []byte("plain text content here that has no image signature")

	tests := []struct {
		name        string
		filename    string
		contentType string
		content     []byte
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Valid JPEG",
			filename:    "profile.jpg",
			contentType: "image/jpeg",
			content:     jpegBytes,
			expectError: false,
		},
		{
			name:        "Valid PNG",
			filename:    "avatar.png",
			contentType: "image/png",
			content:     pngBytes,
			expectError: false,
		},
		{
			name:        "Invalid Extension",
			filename:    "script.exe",
			contentType: "image/jpeg",
			content:     jpegBytes,
			expectError: true,
			errorMsg:    "invalid file extension",
		},
		{
			name:        "Spoofed MIME Type (Plain text as JPG)",
			filename:    "malicious.jpg",
			contentType: "image/jpeg",
			content:     plainTextBytes,
			expectError: true,
			errorMsg:    "invalid file type detected",
		},
		{
			name:        "Exceeds Max Size",
			filename:    "large.jpg",
			contentType: "image/jpeg",
			content:     bytes.Repeat(jpegBytes, 100000), // ~1.2MB
			expectError: true,
			errorMsg:    "file size exceeds limit",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fh, err := createMultipartFileHeader(tc.filename, tc.contentType, tc.content)
			assert.NoError(t, err)

			err = ValidateImageFile(fh)
			if tc.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}

	t.Run("Nil File Header", func(t *testing.T) {
		err := ValidateImageFile(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no file uploaded")
	})
}
