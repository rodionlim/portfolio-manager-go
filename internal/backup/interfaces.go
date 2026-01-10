package backup

import (
	"context"
	"io"
)

// BackupSource defines the interface for different backup sources
type BackupSource interface {
	// Upload uploads data to the backup source
	Upload(ctx context.Context, reader io.Reader, filename string) error

	// Download downloads data from the backup source
	Download(ctx context.Context, filename string) (io.Reader, error)

	// GetName returns the name of the backup source
	GetName() string
}

// BackupConfig holds configuration for backup operations
type BackupConfig struct {
	Source      string `json:"source"`       // local, gdrive, nextcloud
	URI         string `json:"uri"`          // file location or URL
	User        string `json:"user"`         // username for remote sources
	Password    string `json:"password"`     // password for remote sources
	IncludeData bool   `json:"include_data"` // whether to include the data folder
}

// BackupService handles backup and restore operations
type BackupService interface {
	// Backup creates a backup of the database
	Backup(ctx context.Context, dbPath string, config BackupConfig) error

	// Restore restores the database from a backup
	Restore(ctx context.Context, dbPath string, config BackupConfig) error

	// GetBackupSize returns the size of the backup in bytes
	GetBackupSize(dbPath string, includeData bool) (int64, error)

	// IsApplicationRunning checks if the portfolio manager is currently running
	IsApplicationRunning() (bool, error)
}
