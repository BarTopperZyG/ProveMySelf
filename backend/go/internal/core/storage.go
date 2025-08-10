package core

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"path/filepath"
	"strings"
	"time"
)

var (
	ErrFileNotFound     = errors.New("file not found")
	ErrFileTooBig       = errors.New("file too big")
	ErrInvalidFileType  = errors.New("invalid file type")
	ErrStorageUnavailable = errors.New("storage service unavailable")
)

// StorageMetadata contains metadata about a stored file
type StorageMetadata struct {
	Key          string    `json:"key"`
	OriginalName string    `json:"original_name"`
	ContentType  string    `json:"content_type"`
	Size         int64     `json:"size"`
	UploadedAt   time.Time `json:"uploaded_at"`
	URL          string    `json:"url"`
	ETag         string    `json:"etag,omitempty"`
}

// UploadOptions contains options for file upload
type UploadOptions struct {
	MaxSize          int64
	AllowedTypes     []string
	GenerateUniqueName bool
	Prefix           string
}

// Storage defines the interface for file storage operations
type Storage interface {
	// Upload stores a file and returns metadata
	Upload(ctx context.Context, key string, reader io.Reader, opts UploadOptions) (*StorageMetadata, error)
	
	// Download retrieves a file by key
	Download(ctx context.Context, key string) (io.ReadCloser, *StorageMetadata, error)
	
	// Delete removes a file by key
	Delete(ctx context.Context, key string) error
	
	// Exists checks if a file exists
	Exists(ctx context.Context, key string) (bool, error)
	
	// GetURL returns a public URL for the file (if supported)
	GetURL(ctx context.Context, key string) (string, error)
	
	// GetSignedURL returns a signed URL for temporary access
	GetSignedURL(ctx context.Context, key string, expiration time.Duration) (string, error)
	
	// List lists files with optional prefix
	List(ctx context.Context, prefix string, limit int) ([]*StorageMetadata, error)
	
	// HealthCheck checks if the storage service is available
	HealthCheck(ctx context.Context) error
}

// FileUpload represents an uploaded file
type FileUpload struct {
	OriginalName string
	ContentType  string
	Size         int64
	Reader       io.Reader
}

// StorageService handles file storage operations
type StorageService struct {
	storage Storage
	config  StorageConfig
}

// StorageConfig contains storage service configuration
type StorageConfig struct {
	MaxFileSize      int64
	AllowedFileTypes []string
	BaseURL          string
}

// NewStorageService creates a new storage service
func NewStorageService(storage Storage, config StorageConfig) *StorageService {
	return &StorageService{
		storage: storage,
		config:  config,
	}
}

// UploadFile uploads a file with validation
func (s *StorageService) UploadFile(ctx context.Context, projectID string, file FileUpload) (*StorageMetadata, error) {
	// Validate file size
	if file.Size > s.config.MaxFileSize {
		return nil, fmt.Errorf("%w: file size %d exceeds maximum %d", ErrFileTooBig, file.Size, s.config.MaxFileSize)
	}

	// Validate file type
	if !s.isAllowedFileType(file.ContentType) {
		return nil, fmt.Errorf("%w: %s not in allowed types %v", ErrInvalidFileType, file.ContentType, s.config.AllowedFileTypes)
	}

	// Generate storage key
	key := s.generateFileKey(projectID, file.OriginalName)

	// Upload options
	opts := UploadOptions{
		MaxSize:            s.config.MaxFileSize,
		AllowedTypes:       s.config.AllowedFileTypes,
		GenerateUniqueName: true,
		Prefix:             fmt.Sprintf("projects/%s/assets/", projectID),
	}

	// Upload file
	metadata, err := s.storage.Upload(ctx, key, file.Reader, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	return metadata, nil
}

// GetFile retrieves a file by key
func (s *StorageService) GetFile(ctx context.Context, key string) (io.ReadCloser, *StorageMetadata, error) {
	return s.storage.Download(ctx, key)
}

// DeleteFile removes a file by key
func (s *StorageService) DeleteFile(ctx context.Context, key string) error {
	return s.storage.Delete(ctx, key)
}

// GetFileURL returns a public URL for a file
func (s *StorageService) GetFileURL(ctx context.Context, key string) (string, error) {
	return s.storage.GetURL(ctx, key)
}

// GetSignedURL returns a signed URL for temporary access
func (s *StorageService) GetSignedURL(ctx context.Context, key string, expiration time.Duration) (string, error) {
	return s.storage.GetSignedURL(ctx, key, expiration)
}

// ListProjectFiles lists all files for a project
func (s *StorageService) ListProjectFiles(ctx context.Context, projectID string, limit int) ([]*StorageMetadata, error) {
	prefix := fmt.Sprintf("projects/%s/assets/", projectID)
	return s.storage.List(ctx, prefix, limit)
}

// CleanupProjectFiles removes all files for a project
func (s *StorageService) CleanupProjectFiles(ctx context.Context, projectID string) error {
	files, err := s.ListProjectFiles(ctx, projectID, 1000) // Get up to 1000 files
	if err != nil {
		return fmt.Errorf("failed to list project files: %w", err)
	}

	for _, file := range files {
		if err := s.DeleteFile(ctx, file.Key); err != nil {
			// Log error but continue with other files
			// In a real implementation, you'd use structured logging here
			continue
		}
	}

	return nil
}

// HealthCheck checks storage service availability
func (s *StorageService) HealthCheck(ctx context.Context) error {
	return s.storage.HealthCheck(ctx)
}

// generateFileKey creates a unique storage key for a file
func (s *StorageService) generateFileKey(projectID, originalName string) string {
	ext := filepath.Ext(originalName)
	baseName := strings.TrimSuffix(originalName, ext)
	timestamp := time.Now().Unix()
	
	// Create a unique key with timestamp
	return fmt.Sprintf("projects/%s/assets/%s_%d%s", projectID, baseName, timestamp, ext)
}

// isAllowedFileType checks if the content type is allowed
func (s *StorageService) isAllowedFileType(contentType string) bool {
	if len(s.config.AllowedFileTypes) == 0 {
		return true // Allow all types if none specified
	}

	for _, allowed := range s.config.AllowedFileTypes {
		if contentType == allowed {
			return true
		}
	}

	return false
}

// ValidateFileUpload performs basic validation on file upload data
func ValidateFileUpload(file FileUpload, maxSize int64, allowedTypes []string) error {
	if file.Size <= 0 {
		return errors.New("file size must be greater than 0")
	}

	if file.Size > maxSize {
		return fmt.Errorf("%w: file size %d exceeds maximum %d", ErrFileTooBig, file.Size, maxSize)
	}

	if file.OriginalName == "" {
		return errors.New("original filename is required")
	}

	if file.ContentType == "" {
		// Try to detect content type from filename
		file.ContentType = mime.TypeByExtension(filepath.Ext(file.OriginalName))
		if file.ContentType == "" {
			return errors.New("content type could not be determined")
		}
	}

	// Validate file type
	if len(allowedTypes) > 0 {
		allowed := false
		for _, allowedType := range allowedTypes {
			if file.ContentType == allowedType {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("%w: %s not in allowed types %v", ErrInvalidFileType, file.ContentType, allowedTypes)
		}
	}

	return nil
}

// GetContentTypeFromFilename determines content type from filename
func GetContentTypeFromFilename(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	contentType := mime.TypeByExtension(ext)
	
	if contentType == "" {
		// Fallback for common types not detected by mime package
		switch ext {
		case ".webp":
			return "image/webp"
		case ".svg":
			return "image/svg+xml"
		case ".mp4":
			return "video/mp4"
		case ".webm":
			return "video/webm"
		default:
			return "application/octet-stream"
		}
	}
	
	return contentType
}