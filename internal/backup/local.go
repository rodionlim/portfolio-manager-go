package backup

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// LocalBackupSource implements BackupSource for local file system
type LocalBackupSource struct {
	basePath string
}

// NewLocalBackupSource creates a new local backup source
func NewLocalBackupSource(basePath string) *LocalBackupSource {
	return &LocalBackupSource{
		basePath: basePath,
	}
}

// SetBasePath updates the base path for the local backup source
func (l *LocalBackupSource) SetBasePath(basePath string) {
	l.basePath = basePath
}

// Upload saves the data to a local file
func (l *LocalBackupSource) Upload(ctx context.Context, reader io.Reader, filename string) error {
	// Ensure the base path exists
	if err := os.MkdirAll(l.basePath, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}
	
	filePath := filepath.Join(l.basePath, filename)
	
	// Create the backup file
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create backup file: %w", err)
	}
	defer file.Close()
	
	// Copy data from reader to file
	_, err = io.Copy(file, reader)
	if err != nil {
		return fmt.Errorf("failed to write backup data: %w", err)
	}
	
	return nil
}

// Download reads data from a local file
func (l *LocalBackupSource) Download(ctx context.Context, filename string) (io.Reader, error) {
	filePath := filepath.Join(l.basePath, filename)
	
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open backup file: %w", err)
	}
	
	return file, nil
}

// GetName returns the name of the backup source
func (l *LocalBackupSource) GetName() string {
	return "local"
}