package common

import (
	"evasbr/mclamg/exception"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
)

// MaxProfilePicSize is set to 1MB
const MaxProfilePicSize = 1 * 1024 * 1024

var AllowedMIMETypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
}

var AllowedExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".webp": true,
}

// ValidateImageFile validates the size, extension, and actual content signature of the file.
func ValidateImageFile(fileHeader *multipart.FileHeader) error {
	if fileHeader == nil {
		return exception.ValidationError{Message: "no file uploaded"}
	}

	// 1. Check file size
	if fileHeader.Size > MaxProfilePicSize {
		return exception.ValidationError{
			Message: fmt.Sprintf("file size exceeds limit of %d bytes (1MB)", MaxProfilePicSize),
		}
	}

	// 2. Check extension
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if !AllowedExtensions[ext] {
		return exception.ValidationError{
			Message: "invalid file extension, only .jpg, .jpeg, .png, and .webp are allowed",
		}
	}

	// 3. Open file to inspect content type
	file, err := fileHeader.Open()
	if err != nil {
		return exception.ValidationError{
			Message: fmt.Sprintf("unable to open file: %v", err),
		}
	}
	defer file.Close()

	// Read first 512 bytes for content-type detection
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return exception.ValidationError{
			Message: fmt.Sprintf("unable to read file headers: %v", err),
		}
	}

	detectedType := http.DetectContentType(buffer[:n])
	if !AllowedMIMETypes[detectedType] {
		return exception.ValidationError{
			Message: fmt.Sprintf("invalid file type detected: %s. Only JPEG, PNG, and WebP are allowed", detectedType),
		}
	}

	return nil
}
