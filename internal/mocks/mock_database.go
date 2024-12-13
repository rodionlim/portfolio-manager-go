package mocks

import "github.com/stretchr/testify/mock"

// MockDatabase implements dal.Database for testing
type MockDatabase struct {
	mock.Mock
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

func (m *MockDatabase) GetAllKeysWithPrefix(prefix string) ([]string, error) {
	args := m.Called(prefix)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockDatabase) Close() error { return nil }
