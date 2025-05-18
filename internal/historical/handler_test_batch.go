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
type mockHistoricalDeleteService struct {
mock.Mock
}

func (m *mockHistoricalDeleteService) DeleteMetrics(timestamps []string) (DeleteMetricsResponse, error) {
args := m.Called(timestamps)
return args.Get(0).(DeleteMetricsResponse), args.Error(1)
}

func TestHandleDeleteMetrics_Success(t *testing.T) {
mockSvc := new(mockHistoricalDeleteService)
request := DeleteMetricsRequest{
Timestamps: []string{"2024-01-01T00:00:00Z", "2024-02-01T00:00:00Z"},
}

response := DeleteMetricsResponse{
Deleted:  2,
Failed:   0,
Failures: []string{},
}

mockSvc.On("DeleteMetrics", request.Timestamps).Return(response, nil)

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
mockSvc := new(mockHistoricalDeleteService)

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
mockSvc := new(mockHistoricalDeleteService)
request := DeleteMetricsRequest{
Timestamps: []string{"2024-01-01T00:00:00Z", "2024-02-01T00:00:00Z"},
}

mockSvc.On("DeleteMetrics", request.Timestamps).Return(DeleteMetricsResponse{}, errors.New("database error"))

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
mockSvc := new(mockHistoricalDeleteService)

req := httptest.NewRequest("POST", "/api/v1/historical/metrics/delete", bytes.NewBuffer([]byte("invalid json")))
req.Header.Set("Content-Type", "application/json")
rr := httptest.NewRecorder()

handler := HandleDeleteMetrics(mockSvc)
handler.ServeHTTP(rr, req)

assert.Equal(t, http.StatusBadRequest, rr.Code)
mockSvc.AssertNotCalled(t, "DeleteMetrics")
}
