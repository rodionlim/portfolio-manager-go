package backup

import (
	"context"
	"fmt"
	"io"
)

// GDriveBackupSource implements BackupSource for Google Drive
// TODO: Implement Google Drive API integration
type GDriveBackupSource struct {
	credentials string
}

// NewGDriveBackupSource creates a new Google Drive backup source
func NewGDriveBackupSource(credentials string) *GDriveBackupSource {
	return &GDriveBackupSource{
		credentials: credentials,
	}
}

// Upload uploads data to Google Drive
func (g *GDriveBackupSource) Upload(ctx context.Context, reader io.Reader, filename string) error {
	return fmt.Errorf("Google Drive backup not yet implemented")
}

// Download downloads data from Google Drive
func (g *GDriveBackupSource) Download(ctx context.Context, filename string) (io.Reader, error) {
	return nil, fmt.Errorf("Google Drive backup not yet implemented")
}

// GetName returns the name of the backup source
func (g *GDriveBackupSource) GetName() string {
	return "gdrive"
}