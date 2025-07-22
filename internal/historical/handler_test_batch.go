// filepath: /Users/rodionlim/workspace/portfolio-manager-go/internal/historical/handler_test_batch.go
package historical

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock service for testing metrics deletion
type mockHistoricalMetricsSetter struct {
	mock.Mock
}

func (m *mockHistoricalMetricsSetter) DeleteMetrics(timestamps []string, bookFilter string) (DeleteMetricsResponse, error) {
	args := m.Called(timestamps, bookFilter)
	return args.Get(0).(DeleteMetricsResponse), args.Error(1)
}

func (m *mockHistoricalMetricsSetter) DeleteMetric(timestamp string, bookFilter string) error {
	args := m.Called(timestamp, bookFilter)
	return args.Error(0)
}

func (m *mockHistoricalMetricsSetter) UpsertMetric(metric TimestampedMetrics, bookFilter string) error {
	args := m.Called(metric, bookFilter)
	return args.Error(0)
}

func (m *mockHistoricalMetricsSetter) StoreCurrentMetrics(bookFilter string) error {
	args := m.Called(bookFilter)
	return args.Error(0)
}

// Mock service for testing metrics scheduler operations
type mockHistoricalMetricsScheduler struct {
	mock.Mock
}

func (m *mockHistoricalMetricsScheduler) CreateMetricsJob(cronExpr string, bookFilter string) (*MetricsJob, error) {
	args := m.Called(cronExpr, bookFilter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*MetricsJob), args.Error(1)
}

func (m *mockHistoricalMetricsScheduler) DeleteMetricsJob(bookFilter string) error {
	args := m.Called(bookFilter)
	return args.Error(0)
}

func (m *mockHistoricalMetricsScheduler) ListMetricsJobs() ([]MetricsJob, error) {
	args := m.Called()
	return args.Get(0).([]MetricsJob), args.Error(1)
}

func (m *mockHistoricalMetricsScheduler) ListAllMetricsJobsIncludingPortfolio() ([]MetricsJob, error) {
	args := m.Called()
	return args.Get(0).([]MetricsJob), args.Error(1)
}

func (m *mockHistoricalMetricsScheduler) StartMetricsCollection(cronExpr string, bookFilter string) func() {
	args := m.Called(cronExpr, bookFilter)
	return args.Get(0).(func())
}

func (m *mockHistoricalMetricsScheduler) StopMetricsCollection() {
	m.Called()
}

// Tests for HandleDeleteMetrics endpoint
func TestHandleDeleteMetrics_Success(t *testing.T) {
	mockSvc := new(mockHistoricalMetricsSetter)
	request := DeleteMetricsRequest{
		Timestamps: []string{"2024-01-01T00:00:00Z", "2024-02-01T00:00:00Z"},
	}

	response := DeleteMetricsResponse{
		Deleted:  2,
		Failed:   0,
		Failures: []string{},
	}

	mockSvc.On("DeleteMetrics", request.Timestamps, "").Return(response, nil)

	reqBody, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", "/api/v1/historical/metrics/delete", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler := HandleDeleteMetrics(mockSvc)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var respData DeleteMetricsResponse
	err := json.Unmarshal(rr.Body.Bytes(), &respData)
	assert.NoError(t, err)
	assert.Equal(t, 2, respData.Deleted)
	assert.Equal(t, 0, respData.Failed)
	assert.Empty(t, respData.Failures)

	mockSvc.AssertExpectations(t)
}

func TestHandleDeleteMetrics_EmptyTimestamps(t *testing.T) {
	mockSvc := new(mockHistoricalMetricsSetter)

	request := DeleteMetricsRequest{
		Timestamps: []string{},
	}

	reqBody, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", "/api/v1/historical/metrics/delete", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler := HandleDeleteMetrics(mockSvc)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	mockSvc.AssertNotCalled(t, "DeleteMetrics")
}

func TestHandleDeleteMetrics_Error(t *testing.T) {
	mockSvc := new(mockHistoricalMetricsSetter)
	request := DeleteMetricsRequest{
		Timestamps: []string{"2024-01-01T00:00:00Z", "2024-02-01T00:00:00Z"},
	}

	mockSvc.On("DeleteMetrics", request.Timestamps, "").Return(DeleteMetricsResponse{}, errors.New("database error"))

	reqBody, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", "/api/v1/historical/metrics/delete", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler := HandleDeleteMetrics(mockSvc)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	mockSvc.AssertExpectations(t)
}

func TestHandleDeleteMetrics_InvalidJSON(t *testing.T) {
	mockSvc := new(mockHistoricalMetricsSetter)

	req := httptest.NewRequest("POST", "/api/v1/historical/metrics/delete", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler := HandleDeleteMetrics(mockSvc)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	mockSvc.AssertNotCalled(t, "DeleteMetrics")
}

// Tests for HandleCreateMetricsJob endpoint
func TestHandleCreateMetricsJob_Success(t *testing.T) {
	mockSvc := new(mockHistoricalMetricsScheduler)
	request := CreateMetricsJobRequest{
		CronExpr:   "0 0 * * *",
		BookFilter: "test-book",
	}

	expectedJob := &MetricsJob{
		BookFilter: "test-book",
		CronExpr:   "0 0 * * *",
		TaskId:     "test-task-id-1",
	}

	mockSvc.On("CreateMetricsJob", "0 0 * * *", "test-book").Return(expectedJob, nil)

	reqBody, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", "/api/v1/historical/metrics/jobs", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler := HandleCreateMetricsJob(mockSvc)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var respData MetricsJob
	err := json.Unmarshal(rr.Body.Bytes(), &respData)
	assert.NoError(t, err)
	assert.Equal(t, "test-book", respData.BookFilter)
	assert.Equal(t, "0 0 * * *", respData.CronExpr)

	mockSvc.AssertExpectations(t)
}

func TestHandleCreateMetricsJob_EmptyBookFilter(t *testing.T) {
	mockSvc := new(mockHistoricalMetricsScheduler)
	request := CreateMetricsJobRequest{
		CronExpr:   "0 0 * * *",
		BookFilter: "",
	}

	reqBody, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", "/api/v1/historical/metrics/jobs", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler := HandleCreateMetricsJob(mockSvc)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	mockSvc.AssertNotCalled(t, "CreateMetricsJob")
}

func TestHandleCreateMetricsJob_ServiceError(t *testing.T) {
	mockSvc := new(mockHistoricalMetricsScheduler)
	request := CreateMetricsJobRequest{
		CronExpr:   "invalid-cron",
		BookFilter: "test-book",
	}

	mockSvc.On("CreateMetricsJob", "invalid-cron", "test-book").Return(nil, errors.New("invalid cron expression"))

	reqBody, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", "/api/v1/historical/metrics/jobs", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler := HandleCreateMetricsJob(mockSvc)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	mockSvc.AssertExpectations(t)
}

func TestHandleCreateMetricsJob_InvalidJSON(t *testing.T) {
	mockSvc := new(mockHistoricalMetricsScheduler)

	req := httptest.NewRequest("POST", "/api/v1/historical/metrics/jobs", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler := HandleCreateMetricsJob(mockSvc)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	mockSvc.AssertNotCalled(t, "CreateMetricsJob")
}

// Tests for HandleDeleteMetricsJob endpoint
func TestHandleDeleteMetricsJob_Success(t *testing.T) {
	mockSvc := new(mockHistoricalMetricsScheduler)
	bookFilter := "test-book"

	mockSvc.On("DeleteMetricsJob", bookFilter).Return(nil)

	req := httptest.NewRequest("DELETE", "/api/v1/historical/metrics/jobs/"+bookFilter, nil)
	rr := httptest.NewRecorder()

	handler := HandleDeleteMetricsJob(mockSvc)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)
	mockSvc.AssertExpectations(t)
}

func TestHandleDeleteMetricsJob_NotFound(t *testing.T) {
	mockSvc := new(mockHistoricalMetricsScheduler)
	bookFilter := "nonexistent-book"

	mockSvc.On("DeleteMetricsJob", bookFilter).Return(errors.New("metrics job not found for book_filter: " + bookFilter))

	req := httptest.NewRequest("DELETE", "/api/v1/historical/metrics/jobs/"+bookFilter, nil)
	rr := httptest.NewRecorder()

	handler := HandleDeleteMetricsJob(mockSvc)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	mockSvc.AssertExpectations(t)
}

func TestHandleDeleteMetricsJob_EmptyBookFilter(t *testing.T) {
	mockSvc := new(mockHistoricalMetricsScheduler)

	req := httptest.NewRequest("DELETE", "/api/v1/historical/metrics/jobs/", nil)
	rr := httptest.NewRecorder()

	handler := HandleDeleteMetricsJob(mockSvc)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	mockSvc.AssertNotCalled(t, "DeleteMetricsJob")
}

func TestHandleDeleteMetricsJob_ServiceError(t *testing.T) {
	mockSvc := new(mockHistoricalMetricsScheduler)
	bookFilter := "test-book"

	mockSvc.On("DeleteMetricsJob", bookFilter).Return(errors.New("database error"))

	req := httptest.NewRequest("DELETE", "/api/v1/historical/metrics/jobs/"+bookFilter, nil)
	rr := httptest.NewRecorder()

	handler := HandleDeleteMetricsJob(mockSvc)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	mockSvc.AssertExpectations(t)
}

// Tests for HandleListMetricsJobs endpoint
func TestHandleListMetricsJobs_Success(t *testing.T) {
	mockSvc := new(mockHistoricalMetricsScheduler)
	expectedJobs := []MetricsJob{
		{BookFilter: "book1", CronExpr: "0 0 * * *", TaskId: "test-task-id-1"},
		{BookFilter: "book2", CronExpr: "0 12 * * *", TaskId: "test-task-id-2"},
	}

	mockSvc.On("ListMetricsJobs").Return(expectedJobs, nil)

	req := httptest.NewRequest("GET", "/api/v1/historical/metrics/jobs", nil)
	rr := httptest.NewRecorder()

	handler := HandleListMetricsJobs(mockSvc)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var respData []MetricsJob
	err := json.Unmarshal(rr.Body.Bytes(), &respData)
	assert.NoError(t, err)
	assert.Len(t, respData, 2)
	assert.Equal(t, "book1", respData[0].BookFilter)
	assert.Equal(t, "book2", respData[1].BookFilter)

	mockSvc.AssertExpectations(t)
}

func TestHandleListMetricsJobs_EmptyList(t *testing.T) {
	mockSvc := new(mockHistoricalMetricsScheduler)
	expectedJobs := []MetricsJob{}

	mockSvc.On("ListMetricsJobs").Return(expectedJobs, nil)

	req := httptest.NewRequest("GET", "/api/v1/historical/metrics/jobs", nil)
	rr := httptest.NewRecorder()

	handler := HandleListMetricsJobs(mockSvc)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var respData []MetricsJob
	err := json.Unmarshal(rr.Body.Bytes(), &respData)
	assert.NoError(t, err)
	assert.Len(t, respData, 0)

	mockSvc.AssertExpectations(t)
}

func TestHandleListMetricsJobs_ServiceError(t *testing.T) {
	mockSvc := new(mockHistoricalMetricsScheduler)

	mockSvc.On("ListMetricsJobs").Return([]MetricsJob{}, errors.New("database error"))

	req := httptest.NewRequest("GET", "/api/v1/historical/metrics/jobs", nil)
	rr := httptest.NewRecorder()

	handler := HandleListMetricsJobs(mockSvc)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	mockSvc.AssertExpectations(t)
}

// Tests for HandleDeleteMetrics with book_filter parameter
func TestHandleDeleteMetrics_WithBookFilter(t *testing.T) {
	mockSvc := new(mockHistoricalMetricsSetter)
	request := DeleteMetricsRequest{
		Timestamps: []string{"2024-01-01T00:00:00Z", "2024-02-01T00:00:00Z"},
	}

	response := DeleteMetricsResponse{
		Deleted:  2,
		Failed:   0,
		Failures: []string{},
	}

	mockSvc.On("DeleteMetrics", request.Timestamps, "tactical").Return(response, nil)

	reqBody, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", "/api/v1/historical/metrics/delete?book_filter=tactical", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler := HandleDeleteMetrics(mockSvc)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var respData DeleteMetricsResponse
	err := json.Unmarshal(rr.Body.Bytes(), &respData)
	assert.NoError(t, err)
	assert.Equal(t, 2, respData.Deleted)
	assert.Equal(t, 0, respData.Failed)
	assert.Empty(t, respData.Failures)

	mockSvc.AssertExpectations(t)
}
