package backup

import (
	"context"
	"fmt"
	"io"
)

// NextcloudBackupSource implements BackupSource for Nextcloud
// TODO: Implement Nextcloud WebDAV integration
type NextcloudBackupSource struct {
	baseURL  string
	username string
	password string
}

// NewNextcloudBackupSource creates a new Nextcloud backup source
func NewNextcloudBackupSource(baseURL, username, password string) *NextcloudBackupSource {
	return &NextcloudBackupSource{
		baseURL:  baseURL,
		username: username,
		password: password,
	}
}

// Upload uploads data to Nextcloud
func (n *NextcloudBackupSource) Upload(ctx context.Context, reader io.Reader, filename string) error {
	return fmt.Errorf("Nextcloud backup not yet implemented")
}

// Download downloads data from Nextcloud
func (n *NextcloudBackupSource) Download(ctx context.Context, filename string) (io.Reader, error) {
	return nil, fmt.Errorf("Nextcloud backup not yet implemented")
}

// GetName returns the name of the backup source
func (n *NextcloudBackupSource) GetName() string {
	return "nextcloud"
}