package store

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/provemyself/backend/internal/core"
)

// LocalStorage implements the Storage interface using local filesystem
type LocalStorage struct {
	basePath string
	baseURL  string
}

// NewLocalStorage creates a new local storage instance
func NewLocalStorage(basePath, baseURL string) *LocalStorage {
	return &LocalStorage{
		basePath: basePath,
		baseURL:  baseURL,
	}
}

// Upload stores a file on the local filesystem
func (ls *LocalStorage) Upload(ctx context.Context, key string, reader io.Reader, opts core.UploadOptions) (*core.StorageMetadata, error) {
	// Ensure the directory exists
	fullPath := filepath.Join(ls.basePath, key)
	dir := filepath.Dir(fullPath)
	
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	// Create the file
	file, err := os.Create(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Create a hash to track file integrity
	hash := md5.New()
	
	// Copy data while calculating hash and size
	multiWriter := io.MultiWriter(file, hash)
	size, err := io.Copy(multiWriter, reader)
	if err != nil {
		// Clean up the file if copy failed
		os.Remove(fullPath)
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	// Check size limit
	if opts.MaxSize > 0 && size > opts.MaxSize {
		os.Remove(fullPath)
		return nil, fmt.Errorf("%w: file size %d exceeds limit %d", core.ErrFileTooBig, size, opts.MaxSize)
	}

	// Calculate ETag
	etag := hex.EncodeToString(hash.Sum(nil))

	// Create metadata
	metadata := &core.StorageMetadata{
		Key:          key,
		OriginalName: filepath.Base(key),
		ContentType:  core.GetContentTypeFromFilename(key),
		Size:         size,
		UploadedAt:   time.Now(),
		URL:          ls.getPublicURL(key),
		ETag:         etag,
	}

	return metadata, nil
}

// Download retrieves a file from the local filesystem
func (ls *LocalStorage) Download(ctx context.Context, key string) (io.ReadCloser, *core.StorageMetadata, error) {
	fullPath := filepath.Join(ls.basePath, key)
	
	// Check if file exists
	fileInfo, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, core.ErrFileNotFound
		}
		return nil, nil, fmt.Errorf("failed to stat file: %w", err)
	}

	// Open file
	file, err := os.Open(fullPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open file: %w", err)
	}

	// Create metadata
	metadata := &core.StorageMetadata{
		Key:          key,
		OriginalName: filepath.Base(key),
		ContentType:  core.GetContentTypeFromFilename(key),
		Size:         fileInfo.Size(),
		UploadedAt:   fileInfo.ModTime(),
		URL:          ls.getPublicURL(key),
	}

	return file, metadata, nil
}

// Delete removes a file from the local filesystem
func (ls *LocalStorage) Delete(ctx context.Context, key string) error {
	fullPath := filepath.Join(ls.basePath, key)
	
	if err := os.Remove(fullPath); err != nil {
		if os.IsNotExist(err) {
			return core.ErrFileNotFound
		}
		return fmt.Errorf("failed to delete file: %w", err)
	}

	// Try to remove empty directories
	dir := filepath.Dir(fullPath)
	ls.removeEmptyDirs(dir)

	return nil
}

// Exists checks if a file exists
func (ls *LocalStorage) Exists(ctx context.Context, key string) (bool, error) {
	fullPath := filepath.Join(ls.basePath, key)
	
	_, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check file existence: %w", err)
	}

	return true, nil
}

// GetURL returns a public URL for the file
func (ls *LocalStorage) GetURL(ctx context.Context, key string) (string, error) {
	// Check if file exists
	exists, err := ls.Exists(ctx, key)
	if err != nil {
		return "", err
	}
	if !exists {
		return "", core.ErrFileNotFound
	}

	return ls.getPublicURL(key), nil
}

// GetSignedURL returns a signed URL (not implemented for local storage)
func (ls *LocalStorage) GetSignedURL(ctx context.Context, key string, expiration time.Duration) (string, error) {
	// Local storage doesn't support signed URLs, return regular URL
	return ls.GetURL(ctx, key)
}

// List lists files with optional prefix
func (ls *LocalStorage) List(ctx context.Context, prefix string, limit int) ([]*core.StorageMetadata, error) {
	var files []*core.StorageMetadata
	searchPath := filepath.Join(ls.basePath, prefix)

	err := filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// Skip directories that don't exist or can't be accessed
			return nil
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Check limit
		if limit > 0 && len(files) >= limit {
			return filepath.SkipDir
		}

		// Get relative path from base
		relPath, err := filepath.Rel(ls.basePath, path)
		if err != nil {
			return nil // Skip files we can't get relative path for
		}

		// Convert to forward slashes for consistency
		key := filepath.ToSlash(relPath)

		metadata := &core.StorageMetadata{
			Key:          key,
			OriginalName: info.Name(),
			ContentType:  core.GetContentTypeFromFilename(key),
			Size:         info.Size(),
			UploadedAt:   info.ModTime(),
			URL:          ls.getPublicURL(key),
		}

		files = append(files, metadata)
		return nil
	})

	if err != nil {
		// If the search path doesn't exist, return empty list
		if os.IsNotExist(err) {
			return []*core.StorageMetadata{}, nil
		}
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	return files, nil
}

// HealthCheck checks if the storage is accessible
func (ls *LocalStorage) HealthCheck(ctx context.Context) error {
	// Check if base directory exists and is writable
	if err := os.MkdirAll(ls.basePath, 0755); err != nil {
		return fmt.Errorf("storage directory not accessible: %w", err)
	}

	// Try to create and delete a test file
	testPath := filepath.Join(ls.basePath, ".health_check")
	
	file, err := os.Create(testPath)
	if err != nil {
		return fmt.Errorf("storage not writable: %w", err)
	}
	file.Close()

	if err := os.Remove(testPath); err != nil {
		return fmt.Errorf("storage cleanup failed: %w", err)
	}

	return nil
}

// getPublicURL constructs a public URL for a file
func (ls *LocalStorage) getPublicURL(key string) string {
	if ls.baseURL == "" {
		return "/" + key // Default to relative path
	}
	
	return strings.TrimSuffix(ls.baseURL, "/") + "/" + key
}

// removeEmptyDirs removes empty directories up the path
func (ls *LocalStorage) removeEmptyDirs(dir string) {
	// Don't remove the base path
	if dir == ls.basePath || !strings.HasPrefix(dir, ls.basePath) {
		return
	}

	// Try to remove directory if empty
	if err := os.Remove(dir); err == nil {
		// If successful, try parent directory
		parent := filepath.Dir(dir)
		ls.removeEmptyDirs(parent)
	}
}