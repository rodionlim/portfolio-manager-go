package blotter

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"portfolio-manager/internal/dal"
	"portfolio-manager/pkg/logging"
	"portfolio-manager/pkg/types"
	"regexp"
	"strings"
)

// ConfirmationMetadata represents metadata about a trade confirmation
type ConfirmationMetadata struct {
	TradeID      string `json:"tradeId"`
	FileName     string `json:"fileName"`
	ContentType  string `json:"contentType"`
	Size         int64  `json:"size"`
	UploadedDate string `json:"uploadedDate"`
}

// Confirmation represents a trade confirmation document
type Confirmation struct {
	Metadata ConfirmationMetadata `json:"metadata"`
	Data     []byte               `json:"data"`
}

// ConfirmationService handles trade confirmation operations
type ConfirmationService struct {
	db dal.Database
}

// NewConfirmationService creates a new confirmation service
func NewConfirmationService(db dal.Database) *ConfirmationService {
	return &ConfirmationService{
		db: db,
	}
}

// SaveConfirmation saves a trade confirmation to the database
func (cs *ConfirmationService) SaveConfirmation(tradeID string, fileName string, contentType string, data []byte, uploadedDate string) error {
	if tradeID == "" {
		return errors.New("trade ID cannot be empty")
	}

	if len(data) == 0 {
		return errors.New("confirmation data cannot be empty")
	}

	// Validate content type
	validContentTypes := map[string]bool{
		"application/pdf": true,
		"image/png":       true,
		"image/jpeg":      true,
	}
	if !validContentTypes[contentType] {
		return fmt.Errorf("invalid content type: %s. Only PDF, PNG, and JPEG files are allowed", contentType)
	}

	// Sanitize filename - remove path components and restrict to safe characters
	fileName = filepath.Base(fileName)
	fileName = sanitizeFilename(fileName)

	confirmation := Confirmation{
		Metadata: ConfirmationMetadata{
			TradeID:      tradeID,
			FileName:     fileName,
			ContentType:  contentType,
			Size:         int64(len(data)),
			UploadedDate: uploadedDate,
		},
		Data: data,
	}

	key := generateConfirmationKey(tradeID)
	err := cs.db.Put(key, confirmation)
	if err != nil {
		logging.GetLogger().Errorf("Failed to save confirmation for trade %s: %v", tradeID, err)
		return err
	}

	logging.GetLogger().Infof("Saved confirmation for trade %s", tradeID)
	return nil
}

// sanitizeFilename removes path traversal sequences and restricts to safe characters
func sanitizeFilename(filename string) string {
	// Remove any path components
	filename = filepath.Base(filename)

	// Replace unsafe characters with underscore
	reg := regexp.MustCompile(`[^a-zA-Z0-9._-]`)
	filename = reg.ReplaceAllString(filename, "_")

	// Ensure it's not empty after sanitization
	if filename == "" || filename == "." || filename == ".." {
		filename = "confirmation.bin"
	}

	return filename
}

// GetConfirmation retrieves a trade confirmation from the database
func (cs *ConfirmationService) GetConfirmation(tradeID string) (*Confirmation, error) {
	if tradeID == "" {
		return nil, errors.New("trade ID cannot be empty")
	}

	key := generateConfirmationKey(tradeID)
	var confirmation Confirmation
	err := cs.db.Get(key, &confirmation)
	if err != nil {
		return nil, fmt.Errorf("confirmation not found for trade %s", tradeID)
	}

	return &confirmation, nil
}

// GetConfirmationMetadata retrieves only the metadata of a trade confirmation
func (cs *ConfirmationService) GetConfirmationMetadata(tradeID string) (*ConfirmationMetadata, error) {
	confirmation, err := cs.GetConfirmation(tradeID)
	if err != nil {
		return nil, err
	}
	return &confirmation.Metadata, nil
}

// DeleteConfirmation deletes a trade confirmation from the database
func (cs *ConfirmationService) DeleteConfirmation(tradeID string) error {
	if tradeID == "" {
		return errors.New("trade ID cannot be empty")
	}

	key := generateConfirmationKey(tradeID)
	err := cs.db.Delete(key)
	if err != nil {
		logging.GetLogger().Errorf("Failed to delete confirmation for trade %s: %v", tradeID, err)
		return err
	}

	logging.GetLogger().Infof("Deleted confirmation for trade %s", tradeID)
	return nil
}

// HasConfirmation checks if a trade has a confirmation
func (cs *ConfirmationService) HasConfirmation(tradeID string) bool {
	_, err := cs.GetConfirmationMetadata(tradeID)
	return err == nil
}

// GetAllConfirmationMetadata retrieves metadata for all confirmations
func (cs *ConfirmationService) GetAllConfirmationMetadata() ([]ConfirmationMetadata, error) {
	keys, err := cs.db.GetAllKeysWithPrefix(string(types.ConfirmationKeyPrefix))
	if err != nil {
		return nil, err
	}

	metadata := make([]ConfirmationMetadata, 0, len(keys))
	for _, key := range keys {
		var confirmation Confirmation
		err := cs.db.Get(key, &confirmation)
		if err != nil {
			logging.GetLogger().Warnf("Failed to get confirmation for key %s: %v", key, err)
			continue
		}
		metadata = append(metadata, confirmation.Metadata)
	}

	return metadata, nil
}

// ExportConfirmationsAsZip exports confirmations for given trade IDs as a zip archive
func (cs *ConfirmationService) ExportConfirmationsAsZip(tradeIDs []string) ([]byte, error) {
	if len(tradeIDs) == 0 {
		return nil, errors.New("no trade IDs provided")
	}

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	exportedCount := 0
	for _, tradeID := range tradeIDs {
		confirmation, err := cs.GetConfirmation(tradeID)
		if err != nil {
			// Skip trades without confirmations
			logging.GetLogger().Debugf("No confirmation found for trade %s, skipping", tradeID)
			continue
		}

		// Extract file extension from original filename
		fileExt := ""
		if idx := strings.LastIndex(confirmation.Metadata.FileName, "."); idx != -1 {
			fileExt = confirmation.Metadata.FileName[idx:]
		}

		// Create zip file header for this file
		fileName := fmt.Sprintf("%s%s", tradeID, fileExt)
		header := &zip.FileHeader{
			Name:   fileName,
			Method: zip.Deflate,
		}
		header.SetMode(0644)

		writer, err := zw.CreateHeader(header)
		if err != nil {
			return nil, fmt.Errorf("failed to create zip entry for %s: %w", fileName, err)
		}

		if _, err := writer.Write(confirmation.Data); err != nil {
			return nil, fmt.Errorf("failed to write data for %s: %w", fileName, err)
		}

		exportedCount++
	}

	if exportedCount == 0 {
		return nil, errors.New("no confirmations found for the provided trade IDs")
	}

	if err := zw.Close(); err != nil {
		return nil, fmt.Errorf("failed to finalize zip: %w", err)
	}

	logging.GetLogger().Infof("Exported %d confirmations", exportedCount)
	return buf.Bytes(), nil
}

// ImportConfirmationsFromZip imports confirmations from a zip archive
func (cs *ConfirmationService) ImportConfirmationsFromZip(zipData []byte, uploadedDate string) (int, error) {
	buf := bytes.NewReader(zipData)
	zr, err := zip.NewReader(buf, int64(len(zipData)))
	if err != nil {
		return 0, fmt.Errorf("error reading zip: %w", err)
	}

	importedCount := 0
	for _, file := range zr.File {
		if file.FileInfo().IsDir() {
			continue
		}

		reader, err := file.Open()
		if err != nil {
			return importedCount, fmt.Errorf("error opening file %s: %w", file.Name, err)
		}
		data, readErr := io.ReadAll(reader)
		reader.Close()
		if readErr != nil {
			return importedCount, fmt.Errorf("error reading file %s: %w", file.Name, readErr)
		}

		fileName := filepath.Base(file.Name)
		tradeID := fileName
		if idx := strings.LastIndex(fileName, "."); idx != -1 {
			tradeID = fileName[:idx]
		}

		lowerName := strings.ToLower(fileName)
		contentType := "application/octet-stream"
		if strings.HasSuffix(lowerName, ".pdf") {
			contentType = "application/pdf"
		} else if strings.HasSuffix(lowerName, ".png") {
			contentType = "image/png"
		} else if strings.HasSuffix(lowerName, ".jpg") || strings.HasSuffix(lowerName, ".jpeg") {
			contentType = "image/jpeg"
		}

		err = cs.SaveConfirmation(tradeID, fileName, contentType, data, uploadedDate)
		if err != nil {
			logging.GetLogger().Warnf("Failed to import confirmation for %s: %v", tradeID, err)
			continue
		}

		importedCount++
	}

	logging.GetLogger().Infof("Imported %d confirmations from zip", importedCount)
	return importedCount, nil
}

// GetConfirmationsMap returns a map of trade IDs to boolean indicating if they have confirmations
func (cs *ConfirmationService) GetConfirmationsMap(tradeIDs []string) map[string]bool {
	result := make(map[string]bool, len(tradeIDs))
	for _, tradeID := range tradeIDs {
		result[tradeID] = cs.HasConfirmation(tradeID)
	}
	return result
}

// generateConfirmationKey generates a unique key for the confirmation
func generateConfirmationKey(tradeID string) string {
	return fmt.Sprintf("%s:%s", types.ConfirmationKeyPrefix, tradeID)
}

// MarshalJSON customizes JSON marshaling for Confirmation to avoid including large data in responses
func (c *Confirmation) MarshalJSON() ([]byte, error) {
	type Alias Confirmation
	return json.Marshal(&struct {
		Metadata ConfirmationMetadata `json:"metadata"`
		DataSize int                  `json:"dataSize"`
	}{
		Metadata: c.Metadata,
		DataSize: len(c.Data),
	})
}
