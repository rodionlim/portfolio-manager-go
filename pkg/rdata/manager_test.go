package rdata_test

import (
	"errors"
	"testing"

	"portfolio-manager/pkg/rdata"
	"portfolio-manager/pkg/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockDatabase struct {
	mock.Mock
}

func (m *MockDatabase) GetAllKeysWithPrefix(prefix string) ([]string, error) {
	args := m.Called(prefix)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockDatabase) Put(key string, value interface{}) error {
	args := m.Called(key, value)
	return args.Error(0)
}

func (m *MockDatabase) Get(key string, value interface{}) error {
	args := m.Called(key, value)
	return args.Error(0)
}

func (m *MockDatabase) Delete(key string) error {
	args := m.Called(key)
	return args.Error(0)
}

func (m *MockDatabase) Close() error { return nil }

var seedFilePath = "../../seed/refdata.yaml"

func TestNewManager_DatabaseEmpty(t *testing.T) {
	mockDB := new(MockDatabase)
	mockDB.On("GetAllKeysWithPrefix", string(types.ReferenceDataKeyPrefix)).Return([]string{}, nil)
	mockDB.On("Put", mock.Anything, mock.Anything).Return(nil)

	rm, err := rdata.NewManager(mockDB, seedFilePath)
	assert.NoError(t, err)
	assert.NotNil(t, rm)
	mockDB.AssertExpectations(t)
}

func TestNewManager_DatabaseNotEmpty(t *testing.T) {
	mockDB := new(MockDatabase)
	mockDB.On("GetAllKeysWithPrefix", string(types.ReferenceDataKeyPrefix)).Return([]string{"key1"}, nil)

	rm, err := rdata.NewManager(mockDB, seedFilePath)
	assert.NoError(t, err)
	assert.NotNil(t, rm)
	mockDB.AssertExpectations(t)
}

func TestNewManager_ErrorSeedingDatabase(t *testing.T) {
	mockDB := new(MockDatabase)
	mockDB.On("GetAllKeysWithPrefix", string(types.ReferenceDataKeyPrefix)).Return([]string{}, nil)
	mockDB.On("Put", mock.Anything, mock.Anything).Return(errors.New("db error"))

	rm, err := rdata.NewManager(mockDB, seedFilePath)
	assert.Error(t, err)
	assert.Nil(t, rm)
	mockDB.AssertExpectations(t)
}
