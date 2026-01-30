package blotter

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"portfolio-manager/internal/mocks"
)

func TestConfirmationService_SaveConfirmation(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	cs := NewConfirmationService(mockDB)

	tradeID := "test-trade-id"
	fileName := "confirmation.pdf"
	contentType := "application/pdf"
	data := []byte("test data")
	uploadedDate := time.Now().Format(time.RFC3339)

	mockDB.On("Put", mock.Anything, mock.Anything).Return(nil)

	err := cs.SaveConfirmation(tradeID, fileName, contentType, data, uploadedDate)

	assert.NoError(t, err)
	mockDB.AssertExpectations(t)
}

func TestConfirmationService_SaveConfirmation_EmptyTradeID(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	cs := NewConfirmationService(mockDB)

	err := cs.SaveConfirmation("", "file.pdf", "application/pdf", []byte("data"), time.Now().Format(time.RFC3339))

	assert.Error(t, err)
	assert.Equal(t, "trade ID cannot be empty", err.Error())
}

func TestConfirmationService_SaveConfirmation_EmptyData(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	cs := NewConfirmationService(mockDB)

	err := cs.SaveConfirmation("trade-id", "file.pdf", "application/pdf", []byte{}, time.Now().Format(time.RFC3339))

	assert.Error(t, err)
	assert.Equal(t, "confirmation data cannot be empty", err.Error())
}

func TestConfirmationService_GetConfirmation(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	cs := NewConfirmationService(mockDB)

	tradeID := "test-trade-id"
	expectedConfirmation := Confirmation{
		Metadata: ConfirmationMetadata{
			TradeID:      tradeID,
			FileName:     "test.pdf",
			ContentType:  "application/pdf",
			Size:         100,
			UploadedDate: time.Now().Format(time.RFC3339),
		},
		Data: []byte("test data"),
	}

	mockDB.On("Get", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		arg := args.Get(1).(*Confirmation)
		*arg = expectedConfirmation
	}).Return(nil)

	confirmation, err := cs.GetConfirmation(tradeID)

	assert.NoError(t, err)
	assert.NotNil(t, confirmation)
	assert.Equal(t, expectedConfirmation.Metadata.TradeID, confirmation.Metadata.TradeID)
	assert.Equal(t, expectedConfirmation.Metadata.FileName, confirmation.Metadata.FileName)
	mockDB.AssertExpectations(t)
}

func TestConfirmationService_GetConfirmation_NotFound(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	cs := NewConfirmationService(mockDB)

	mockDB.On("Get", mock.Anything, mock.Anything).Return(assert.AnError)

	confirmation, err := cs.GetConfirmation("non-existent-id")

	assert.Error(t, err)
	assert.Nil(t, confirmation)
	mockDB.AssertExpectations(t)
}

func TestConfirmationService_DeleteConfirmation(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	cs := NewConfirmationService(mockDB)

	tradeID := "test-trade-id"
	mockDB.On("Delete", mock.Anything).Return(nil)

	err := cs.DeleteConfirmation(tradeID)

	assert.NoError(t, err)
	mockDB.AssertExpectations(t)
}

func TestConfirmationService_DeleteConfirmation_EmptyTradeID(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	cs := NewConfirmationService(mockDB)

	err := cs.DeleteConfirmation("")

	assert.Error(t, err)
	assert.Equal(t, "trade ID cannot be empty", err.Error())
}

func TestConfirmationService_HasConfirmation(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	cs := NewConfirmationService(mockDB)

	tradeID := "test-trade-id"
	expectedConfirmation := Confirmation{
		Metadata: ConfirmationMetadata{
			TradeID: tradeID,
		},
	}

	mockDB.On("Get", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		arg := args.Get(1).(*Confirmation)
		*arg = expectedConfirmation
	}).Return(nil)

	has := cs.HasConfirmation(tradeID)

	assert.True(t, has)
	mockDB.AssertExpectations(t)
}

func TestConfirmationService_HasConfirmation_NotFound(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	cs := NewConfirmationService(mockDB)

	mockDB.On("Get", mock.Anything, mock.Anything).Return(assert.AnError)

	has := cs.HasConfirmation("non-existent-id")

	assert.False(t, has)
	mockDB.AssertExpectations(t)
}

func TestConfirmationService_ExportConfirmationsAsTar(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	cs := NewConfirmationService(mockDB)

	tradeIDs := []string{"trade-1", "trade-2"}
	
	// Mock confirmations for both trades
	confirmation1 := Confirmation{
		Metadata: ConfirmationMetadata{
			TradeID:      "trade-1",
			FileName:     "test1.pdf",
			ContentType:  "application/pdf",
			Size:         10,
			UploadedDate: time.Now().Format(time.RFC3339),
		},
		Data: []byte("test data 1"),
	}
	
	confirmation2 := Confirmation{
		Metadata: ConfirmationMetadata{
			TradeID:      "trade-2",
			FileName:     "test2.pdf",
			ContentType:  "application/pdf",
			Size:         10,
			UploadedDate: time.Now().Format(time.RFC3339),
		},
		Data: []byte("test data 2"),
	}

	mockDB.On("Get", "CONFIRMATION:trade-1", mock.Anything).Run(func(args mock.Arguments) {
		arg := args.Get(1).(*Confirmation)
		*arg = confirmation1
	}).Return(nil)

	mockDB.On("Get", "CONFIRMATION:trade-2", mock.Anything).Run(func(args mock.Arguments) {
		arg := args.Get(1).(*Confirmation)
		*arg = confirmation2
	}).Return(nil)

	tarData, err := cs.ExportConfirmationsAsTar(tradeIDs)

	assert.NoError(t, err)
	assert.NotNil(t, tarData)
	assert.True(t, len(tarData) > 0)
	mockDB.AssertExpectations(t)
}

func TestConfirmationService_ExportConfirmationsAsTar_NoTradeIDs(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	cs := NewConfirmationService(mockDB)

	tarData, err := cs.ExportConfirmationsAsTar([]string{})

	assert.Error(t, err)
	assert.Nil(t, tarData)
	assert.Equal(t, "no trade IDs provided", err.Error())
}

func TestConfirmationService_ExportConfirmationsAsTar_NoConfirmationsFound(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	cs := NewConfirmationService(mockDB)

	tradeIDs := []string{"trade-1", "trade-2"}

	mockDB.On("Get", mock.Anything, mock.Anything).Return(assert.AnError)

	tarData, err := cs.ExportConfirmationsAsTar(tradeIDs)

	assert.Error(t, err)
	assert.Nil(t, tarData)
	assert.Contains(t, err.Error(), "no confirmations found")
	mockDB.AssertExpectations(t)
}

func TestConfirmationService_ImportConfirmationsFromTar(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	cs := NewConfirmationService(mockDB)

	// Create a simple tar with empty content
	var buf bytes.Buffer
	
	count, err := cs.ImportConfirmationsFromTar(buf.Bytes(), time.Now().Format(time.RFC3339))

	// Empty tar should not cause error, just return 0 count
	assert.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestConfirmationService_GetConfirmationsMap(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	cs := NewConfirmationService(mockDB)

	tradeIDs := []string{"trade-1", "trade-2", "trade-3"}
	
	// Mock only trade-1 and trade-2 having confirmations
	mockDB.On("Get", "CONFIRMATION:trade-1", mock.Anything).Return(nil)
	mockDB.On("Get", "CONFIRMATION:trade-2", mock.Anything).Return(nil)
	mockDB.On("Get", "CONFIRMATION:trade-3", mock.Anything).Return(assert.AnError)

	result := cs.GetConfirmationsMap(tradeIDs)

	assert.Equal(t, 3, len(result))
	assert.True(t, result["trade-1"])
	assert.True(t, result["trade-2"])
	assert.False(t, result["trade-3"])
	mockDB.AssertExpectations(t)
}

func TestGenerateConfirmationKey(t *testing.T) {
	tradeID := "test-trade-id"
	key := generateConfirmationKey(tradeID)
	
	assert.Equal(t, "CONFIRMATION:test-trade-id", key)
}
