package backup

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocalBackupSource(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "backup_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	source := NewLocalBackupSource(tempDir)
	ctx := context.Background()

	t.Run("Upload and Download", func(t *testing.T) {
		testData := "test backup data"
		filename := "test-backup.txt"

		// Test upload
		err := source.Upload(ctx, bytes.NewBufferString(testData), filename)
		assert.NoError(t, err)

		// Verify file exists
		filePath := filepath.Join(tempDir, filename)
		assert.FileExists(t, filePath)

		// Test download
		reader, err := source.Download(ctx, filename)
		assert.NoError(t, err)

		// Read data back
		data, err := io.ReadAll(reader)
		assert.NoError(t, err)
		assert.Equal(t, testData, string(data))

		// Close the reader
		if file, ok := reader.(*os.File); ok {
			file.Close()
		}
	})

	t.Run("GetName", func(t *testing.T) {
		assert.Equal(t, "local", source.GetName())
	})

	t.Run("SetBasePath", func(t *testing.T) {
		newPath := "/new/path"
		source.SetBasePath(newPath)
		assert.Equal(t, newPath, source.basePath)
	})
}

func TestBackupService(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "backup_service_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	service := NewService()

	t.Run("GetBackupSize", func(t *testing.T) {
		// Create a test database directory with some files
		dbPath := filepath.Join(tempDir, "test.db")
		err := os.MkdirAll(dbPath, 0755)
		require.NoError(t, err)

		// Create test files
		testFile1 := filepath.Join(dbPath, "file1.log")
		testFile2 := filepath.Join(dbPath, "file2.dat")

		err = os.WriteFile(testFile1, []byte("test data 1"), 0644)
		require.NoError(t, err)

		err = os.WriteFile(testFile2, []byte("test data 2 longer"), 0644)
		require.NoError(t, err)

		// Test size calculation
		size, err := service.GetBackupSize(dbPath, false)
		assert.NoError(t, err)
		assert.Greater(t, size, int64(0))

		// Size should be the sum of both files
		expectedSize := int64(len("test data 1") + len("test data 2 longer"))
		assert.Equal(t, expectedSize, size)
	})

	t.Run("GetBackupSize with optional data", func(t *testing.T) {
		dbPath := filepath.Join(tempDir, "test_optional.db")
		err := os.MkdirAll(dbPath, 0755)
		require.NoError(t, err)

		// Create a fake ./data folder in the current directory for the duration of the test
		// Note: This relies on the test being run with the project root or test dir as CWD
		// Usually tests should use a mockable path, but we'll follow previous pattern here
		dataDir := "./data"
		dataDirExists := false
		if _, err := os.Stat(dataDir); err == nil {
			dataDirExists = true
		}

		if !dataDirExists {
			err = os.MkdirAll(dataDir, 0755)
			require.NoError(t, err)
			defer os.RemoveAll(dataDir)

			testFileData := "some data file"
			err = os.WriteFile(filepath.Join(dataDir, "test.xlsx"), []byte(testFileData), 0644)
			require.NoError(t, err)

			// Test with includeData=true
			sizeWithData, err := service.GetBackupSize(dbPath, true)
			assert.NoError(t, err)

			// Test with includeData=false
			sizeWithoutData, err := service.GetBackupSize(dbPath, false)
			assert.NoError(t, err)

			assert.Greater(t, sizeWithData, sizeWithoutData)
		} else {
			t.Log("Skipping optional data test because ./data already exists")
		}
	})

	t.Run("Backup and Restore", func(t *testing.T) {
		// Create a test database directory
		dbPath := filepath.Join(tempDir, "portfolio-manager.db")
		err := os.MkdirAll(dbPath, 0755)
		require.NoError(t, err)

		// Create test database files
		testFile := filepath.Join(dbPath, "test.log")
		testData := "test database content"
		err = os.WriteFile(testFile, []byte(testData), 0644)
		require.NoError(t, err)

		// Create backup configuration
		backupDir := filepath.Join(tempDir, "backups")
		config := BackupConfig{
			Source: "local",
			URI:    backupDir,
		}

		ctx := context.Background()

		// Test backup
		err = service.Backup(ctx, dbPath, config)
		assert.NoError(t, err)

		// Verify backup file was created
		files, err := os.ReadDir(backupDir)
		assert.NoError(t, err)
		assert.Len(t, files, 1)
		assert.Contains(t, files[0].Name(), "portfolio-manager-backup-")
		assert.Contains(t, files[0].Name(), ".tar.gz")

		// Test restore
		backupFilename := files[0].Name()
		restoreConfig := BackupConfig{
			Source: "local",
			URI:    filepath.Join(backupDir, backupFilename),
		}

		// Remove original database
		err = os.RemoveAll(dbPath)
		require.NoError(t, err)

		// Restore from backup
		err = service.Restore(ctx, dbPath, restoreConfig)
		assert.NoError(t, err)

		// Verify restored data
		restoredFile := filepath.Join(dbPath, "test.log")
		assert.FileExists(t, restoredFile)

		restoredData, err := os.ReadFile(restoredFile)
		assert.NoError(t, err)
		assert.Equal(t, testData, string(restoredData))
	})
}

func TestCreateBackupSource(t *testing.T) {
	service := NewService()

	t.Run("Local source", func(t *testing.T) {
		config := BackupConfig{
			Source: "local",
			URI:    "/test/path",
		}

		source, err := service.createBackupSource(config)
		assert.NoError(t, err)
		assert.NotNil(t, source)
		assert.Equal(t, "local", source.GetName())
	})

	t.Run("Local source with empty URI", func(t *testing.T) {
		config := BackupConfig{
			Source: "local",
		}

		source, err := service.createBackupSource(config)
		assert.NoError(t, err)
		assert.NotNil(t, source)
		assert.Equal(t, "local", source.GetName())
	})

	t.Run("Unsupported source", func(t *testing.T) {
		config := BackupConfig{
			Source: "unsupported",
		}

		source, err := service.createBackupSource(config)
		assert.Error(t, err)
		assert.Nil(t, source)
		assert.Contains(t, err.Error(), "unsupported backup source")
	})

	t.Run("gdrive source", func(t *testing.T) {
		config := BackupConfig{
			Source: "gdrive",
			User:   "test@gmail.com",
		}

		source, err := service.createBackupSource(config)
		assert.NoError(t, err)
		assert.NotNil(t, source)
		assert.Equal(t, "gdrive", source.GetName())
	})

	t.Run("nextcloud source", func(t *testing.T) {
		config := BackupConfig{
			Source:   "nextcloud",
			URI:      "https://nextcloud.example.com",
			User:     "testuser",
			Password: "testpass",
		}

		source, err := service.createBackupSource(config)
		assert.NoError(t, err)
		assert.NotNil(t, source)
		assert.Equal(t, "nextcloud", source.GetName())
	})
}
