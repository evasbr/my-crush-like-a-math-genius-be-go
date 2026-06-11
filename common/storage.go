package common

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

// StorageFolder defines a strongly-typed enum for Cloudinary upload folders to prevent typos.
type StorageFolder string

const (
	FolderDefault         StorageFolder = "mclamg/upload"
	FolderProfilePictures StorageFolder = "mclamg/profile_pictures"
)

// FileStorage defines a provider-agnostic interface for uploading and deleting files.
type FileStorage interface {
	UploadFile(ctx context.Context, file io.Reader, filename string, folder StorageFolder) (string, error)
	DeleteFile(ctx context.Context, fileURL string) error
}

type cloudinaryStorage struct {
	cld *cloudinary.Cloudinary
}

// NewCloudinaryStorage initializes a Cloudinary storage implementation.
func NewCloudinaryStorage(cloudinaryURL string) (FileStorage, error) {
	if cloudinaryURL == "" {
		return nil, errors.New("cloudinary URL is required")
	}
	cld, err := cloudinary.NewFromURL(cloudinaryURL)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize cloudinary: %w", err)
	}
	return &cloudinaryStorage{cld: cld}, nil
}

func (c *cloudinaryStorage) UploadFile(ctx context.Context, file io.Reader, filename string, folder StorageFolder) (string, error) {
	// Strip extension to get the raw name
	cleanName := strings.TrimSuffix(filename, filepath.Ext(filename))

	// Append current Unix timestamp to make the filename unique
	publicID := fmt.Sprintf("%s_%d", cleanName, time.Now().Unix())

	folderName := string(folder)
	if folderName == "" {
		folderName = string(FolderDefault)
	}

	resp, err := c.cld.Upload.Upload(ctx, file, uploader.UploadParams{
		Folder:         folderName,
		PublicID:       publicID,
		UniqueFilename: api.Bool(false),
		Overwrite:      api.Bool(true),
	})
	if err != nil {
		return "", fmt.Errorf("cloudinary upload failed: %w", err)
	}

	return resp.SecureURL, nil
}

func (c *cloudinaryStorage) DeleteFile(ctx context.Context, fileURL string) error {
	if fileURL == "" {
		return nil
	}

	publicID := ExtractCloudinaryPublicID(fileURL)
	if publicID == "" {
		return nil // URL does not match Cloudinary format or public ID could not be resolved
	}

	resp, err := c.cld.Upload.Destroy(ctx, uploader.DestroyParams{
		PublicID: publicID,
	})
	if err != nil {
		return fmt.Errorf("cloudinary destroy failed: %w", err)
	}

	if resp.Result != "ok" && resp.Result != "not found" {
		return fmt.Errorf("cloudinary delete returned status: %s", resp.Result)
	}

	return nil
}

// disabledStorage is a Null Object implementation of FileStorage used when no credentials are provided.
type disabledStorage struct{}

func NewDisabledStorage() FileStorage {
	return &disabledStorage{}
}

func (d *disabledStorage) UploadFile(ctx context.Context, file io.Reader, filename string, folder StorageFolder) (string, error) {
	return "", errors.New("file storage is not configured (missing CLOUDINARY_URL)")
}

func (d *disabledStorage) DeleteFile(ctx context.Context, fileURL string) error {
	return errors.New("file storage is not configured (missing CLOUDINARY_URL)")
}

// NewFileStorage is a helper factory to initialize the correct storage engine.
func NewFileStorage(cloudinaryURL string) (FileStorage, error) {
	if cloudinaryURL == "" {
		return NewDisabledStorage(), nil
	}
	return NewCloudinaryStorage(cloudinaryURL)
}

// ExtractCloudinaryPublicID parses the public ID of an asset from its full Cloudinary secure URL.
func ExtractCloudinaryPublicID(url string) string {
	// Example URL: https://res.cloudinary.com/cloud_name/image/upload/v1234567/folder/subfolder/filename.jpg
	parts := strings.Split(url, "/image/upload/")
	if len(parts) < 2 {
		return ""
	}
	pathPart := parts[1] // e.g. "v1234567/folder/subfolder/filename.jpg"

	subParts := strings.Split(pathPart, "/")
	if len(subParts) == 0 {
		return ""
	}

	// Skip the version number if it exists (e.g. starts with 'v' followed by digits)
	startIndex := 0
	if len(subParts) > 1 && strings.HasPrefix(subParts[0], "v") {
		versionStr := subParts[0][1:]
		isNumeric := true
		for _, char := range versionStr {
			if char < '0' || char > '9' {
				isNumeric = false
				break
			}
		}
		if isNumeric {
			startIndex = 1
		}
	}

	remaining := strings.Join(subParts[startIndex:], "/")

	// Strip the file extension (e.g. .jpg)
	ext := filepath.Ext(remaining)
	if ext != "" {
		remaining = strings.TrimSuffix(remaining, ext)
	}

	return remaining
}
