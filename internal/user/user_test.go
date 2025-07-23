package user

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockDatabase is a mock implementation of dal.Database for testing
type MockDatabase struct {
	mock.Mock
}

func (m *MockDatabase) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockDatabase) Get(key string, v interface{}) error {
	args := m.Called(key, v)
	return args.Error(0)
}

func (m *MockDatabase) Put(key string, v interface{}) error {
	args := m.Called(key, v)
	return args.Error(0)
}

func (m *MockDatabase) Delete(key string) error {
	args := m.Called(key)
	return args.Error(0)
}

func (m *MockDatabase) GetAllKeysWithPrefix(prefix string) ([]string, error) {
	args := m.Called(prefix)
	return args.Get(0).([]string), args.Error(1)
}

func TestService_GetProfile_Default(t *testing.T) {
	mockDB := new(MockDatabase)
	mockDB.On("Get", UserProfileKey, mock.Anything).Return(assert.AnError)

	service := NewService(mockDB)
	profile, err := service.GetProfile()

	assert.NoError(t, err)
	assert.Equal(t, "User", profile.Username)
	assert.Equal(t, "user@example.com", profile.Email)
	assert.Equal(t, "", profile.Avatar)
	mockDB.AssertExpectations(t)
}

func TestService_UpdateProfile_Success(t *testing.T) {
	mockDB := new(MockDatabase)
	mockDB.On("Put", UserProfileKey, mock.Anything).Return(nil)

	service := NewService(mockDB)
	profile := &Profile{
		Username: "TestUser",
		Email:    "test@example.com",
		Avatar:   "https://example.com/avatar.png",
	}

	err := service.UpdateProfile(profile)

	assert.NoError(t, err)
	mockDB.AssertExpectations(t)
}

func TestService_UpdateProfile_EmptyUsername(t *testing.T) {
	mockDB := new(MockDatabase)
	service := NewService(mockDB)
	
	profile := &Profile{
		Username: "",
		Email:    "test@example.com",
		Avatar:   "",
	}

	err := service.UpdateProfile(profile)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "username cannot be empty")
}

func TestService_UpdateProfile_EmptyEmail(t *testing.T) {
	mockDB := new(MockDatabase)
	service := NewService(mockDB)
	
	profile := &Profile{
		Username: "TestUser",
		Email:    "",
		Avatar:   "",
	}

	err := service.UpdateProfile(profile)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "email cannot be empty")
}

func TestService_UpdateProfile_NilProfile(t *testing.T) {
	mockDB := new(MockDatabase)
	service := NewService(mockDB)

	err := service.UpdateProfile(nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "profile cannot be nil")
}