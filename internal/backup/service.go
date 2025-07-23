package backup

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"portfolio-manager/pkg/logging"
)

// Service implements BackupService
type Service struct{}

// NewService creates a new backup service
func NewService() *Service {
	return &Service{}
}

// Backup creates a backup of the LevelDB database
func (s *Service) Backup(ctx context.Context, dbPath string, config BackupConfig) error {
	logger := logging.GetLogger()
	
	// Check if database exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return fmt.Errorf("database path does not exist: %s", dbPath)
	}
	
	// Create backup source
	source, err := s.createBackupSource(config)
	if err != nil {
		return fmt.Errorf("failed to create backup source: %w", err)
	}
	
	// Generate backup filename with timestamp
	timestamp := time.Now().Format("20060102-150405")
	backupFilename := fmt.Sprintf("portfolio-manager-backup-%s.tar.gz", timestamp)
	
	logger.Infof("Creating backup: %s", backupFilename)
	
	// Create a pipe to stream the compressed backup
	reader, writer := io.Pipe()
	
	// Start compression in a goroutine
	go func() {
		defer writer.Close()
		if err := s.compressDatabase(dbPath, writer); err != nil {
			logger.Errorf("Failed to compress database: %v", err)
			writer.CloseWithError(err)
		}
	}()
	
	// Upload the backup
	if err := source.Upload(ctx, reader, backupFilename); err != nil {
		return fmt.Errorf("failed to upload backup: %w", err)
	}
	
	logger.Infof("Backup completed successfully: %s", backupFilename)
	return nil
}

// Restore restores the database from a backup
func (s *Service) Restore(ctx context.Context, dbPath string, config BackupConfig) error {
	logger := logging.GetLogger()
	
	// Create backup source
	source, err := s.createBackupSource(config)
	if err != nil {
		return fmt.Errorf("failed to create backup source: %w", err)
	}
	
	// Determine backup filename based on source and URI
	var backupFilename string
	if config.Source == "local" && config.URI != "" {
		// For local source, URI contains the full path to the backup file
		backupFilename = filepath.Base(config.URI)
		// Update the local source to use the directory containing the backup
		if localSource, ok := source.(*LocalBackupSource); ok {
			localSource.SetBasePath(filepath.Dir(config.URI))
		}
	} else {
		// For other sources, use a default filename
		backupFilename = "portfolio-manager-backup.tar.gz"
	}
	
	logger.Infof("Restoring from backup: %s", backupFilename)
	
	// Download the backup
	reader, err := source.Download(ctx, backupFilename)
	if err != nil {
		return fmt.Errorf("failed to download backup: %w", err)
	}
	
	// Ensure we close the reader if it's a file
	if file, ok := reader.(*os.File); ok {
		defer file.Close()
	}
	
	// Remove existing database if it exists
	if _, err := os.Stat(dbPath); err == nil {
		if err := os.RemoveAll(dbPath); err != nil {
			return fmt.Errorf("failed to remove existing database: %w", err)
		}
	}
	
	// Extract the backup
	if err := s.extractDatabase(reader, filepath.Dir(dbPath)); err != nil {
		return fmt.Errorf("failed to extract backup: %w", err)
	}
	
	logger.Infof("Restore completed successfully")
	return nil
}

// GetBackupSize returns the size of the database directory in bytes
func (s *Service) GetBackupSize(dbPath string) (int64, error) {
	var size int64
	err := filepath.Walk(dbPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size, err
}

// IsApplicationRunning checks if the portfolio manager is currently running
func (s *Service) IsApplicationRunning() (bool, error) {
	// Simple check: try to connect to the default port
	// In a real implementation, you might want to check for a PID file or use a more robust method
	return false, nil // For now, assume it's not running
}

// compressDatabase compresses the database directory into a tar.gz stream
func (s *Service) compressDatabase(dbPath string, writer io.Writer) error {
	gzipWriter := gzip.NewWriter(writer)
	defer gzipWriter.Close()
	
	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()
	
	return filepath.Walk(dbPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// Create tar header
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		
		// Update header name to be relative to database path
		relPath, err := filepath.Rel(filepath.Dir(dbPath), path)
		if err != nil {
			return err
		}
		header.Name = relPath
		
		// Write header
		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}
		
		// Write file data if it's a regular file
		if info.Mode().IsRegular() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			
			_, err = io.Copy(tarWriter, file)
			if err != nil {
				return err
			}
		}
		
		return nil
	})
}

// extractDatabase extracts a tar.gz backup to the specified directory
func (s *Service) extractDatabase(reader io.Reader, targetDir string) error {
	gzipReader, err := gzip.NewReader(reader)
	if err != nil {
		return err
	}
	defer gzipReader.Close()
	
	tarReader := tar.NewReader(gzipReader)
	
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		
		targetPath := filepath.Join(targetDir, header.Name)
		
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(targetPath, os.FileMode(header.Mode)); err != nil {
				return err
			}
		case tar.TypeReg:
			// Ensure parent directory exists
			if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
				return err
			}
			
			file, err := os.Create(targetPath)
			if err != nil {
				return err
			}
			
			if _, err := io.Copy(file, tarReader); err != nil {
				file.Close()
				return err
			}
			file.Close()
			
			// Set file permissions
			if err := os.Chmod(targetPath, os.FileMode(header.Mode)); err != nil {
				return err
			}
		}
	}
	
	return nil
}

// createBackupSource creates a backup source based on the configuration
func (s *Service) createBackupSource(config BackupConfig) (BackupSource, error) {
	switch config.Source {
	case "local":
		basePath := config.URI
		if basePath == "" {
			basePath = "./backups"
		}
		return NewLocalBackupSource(basePath), nil
	case "gdrive":
		return NewGDriveBackupSource(config.User), nil
	case "nextcloud":
		return NewNextcloudBackupSource(config.URI, config.User, config.Password), nil
	default:
		return nil, fmt.Errorf("unsupported backup source: %s", config.Source)
	}
}