package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"portfolio-manager/pkg/logging"
)

// TestUpcheckHandler tests the upcheckHandler function.
func TestUpcheckHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatalf("could not create request: %v", err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(upcheckHandler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expected := "I'm up!"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}

// TestStart tests the Start function.
func TestStart(t *testing.T) {
	logger, err := logging.InitializeLogger(true, "")
	if err != nil {
		t.Fatalf("could not initialize logger: %v", err)
	}

	ctx := context.WithValue(context.Background(), "logger", logger)
	srv := NewServer(":0") // Use port 0 to get an available port

	go func() {
		if err := srv.Start(ctx); err != nil && err != http.ErrServerClosed {
			t.Fatalf("could not start server: %v", err)
		}
	}()

	// Give the server a moment to start
	<-time.After(time.Millisecond * 100)

	req, err := http.NewRequest("GET", "http://"+srv.Addr+"/", nil)
	if err != nil {
		t.Fatalf("could not create request: %v", err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(upcheckHandler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expected := "I'm up!"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}