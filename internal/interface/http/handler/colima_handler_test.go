package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gqadonis/colima-manager/internal/domain"
	"github.com/labstack/echo/v4"
)

type mockUseCase struct {
	mockDependencyStatus *domain.DependencyStatus
	mockColimaStatus     *domain.ColimaStatus
	mockKubeConfig       string
	mockError            error
}

func (m *mockUseCase) CheckDependencies(ctx context.Context) (*domain.DependencyStatus, error) {
	return m.mockDependencyStatus, m.mockError
}

func (m *mockUseCase) UpdateDependencies(ctx context.Context) error {
	return m.mockError
}

func (m *mockUseCase) Start(ctx context.Context, config domain.ColimaConfig) error {
	return m.mockError
}

func (m *mockUseCase) Stop(ctx context.Context, profile string) error {
	return m.mockError
}

func (m *mockUseCase) Status(ctx context.Context, profile string) (*domain.ColimaStatus, error) {
	return m.mockColimaStatus, m.mockError
}

func (m *mockUseCase) GetKubeConfig(ctx context.Context, profile string) (string, error) {
	return m.mockKubeConfig, m.mockError
}

func (m *mockUseCase) Clean(ctx context.Context, req domain.CleanRequest) error {
	return m.mockError
}

func TestHandlerProfileBusy(t *testing.T) {
	// Create mock use case that returns ProfileBusyError
	mockUC := &mockUseCase{
		mockError: &domain.ProfileBusyError{Profile: "test-profile"},
	}

	// Create handler with mock use case
	h := NewColimaHandler(mockUC)

	// Create Echo instance
	e := echo.New()

	// Test Start endpoint
	startConfig := domain.ColimaConfig{
		Profile: "test-profile",
		CPUs:    4,
		Memory:  8,
	}
	startJSON, _ := json.Marshal(startConfig)

	req := httptest.NewRequest(http.MethodPost, "/start", bytes.NewReader(startJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute request
	err := h.Start(c)
	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	// Check response
	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status code %d, got %d", http.StatusServiceUnavailable, rec.Code)
	}

	var response map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["code"] != "profile_busy" {
		t.Errorf("Expected code 'profile_busy', got '%s'", response["code"])
	}

	expectedError := "profile 'test-profile' is currently busy with another operation"
	if response["error"] != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, response["error"])
	}
}

func TestHandlerSuccess(t *testing.T) {
	// Create mock use case with successful responses
	mockUC := &mockUseCase{
		mockDependencyStatus: &domain.DependencyStatus{
			Homebrew: true,
			Colima:   true,
			Lima:     true,
		},
		mockColimaStatus: &domain.ColimaStatus{
			Status: "Running",
			CPUs:   4,
			Memory: 8,
		},
		mockKubeConfig: "test-kubeconfig",
		mockError:      nil,
	}

	// Create handler with mock use case
	h := NewColimaHandler(mockUC)

	// Create Echo instance
	e := echo.New()

	// Test cases for each endpoint
	tests := []struct {
		name           string
		method         string
		path           string
		body           interface{}
		expectedStatus int
	}{
		{
			name:           "CheckDependencies",
			method:         http.MethodGet,
			path:           "/dependencies",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Start",
			method:         http.MethodPost,
			path:           "/start",
			body:           domain.ColimaConfig{Profile: "test"},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Stop",
			method:         http.MethodPost,
			path:           "/stop",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Status",
			method:         http.MethodGet,
			path:           "/status",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "GetKubeConfig",
			method:         http.MethodGet,
			path:           "/kubeconfig",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Clean",
			method:         http.MethodPost,
			path:           "/clean",
			body:           domain.CleanRequest{Profile: "test"},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var reqBody []byte
			var err error
			if tt.body != nil {
				reqBody, err = json.Marshal(tt.body)
				if err != nil {
					t.Fatalf("Failed to marshal request body: %v", err)
				}
			}

			req := httptest.NewRequest(tt.method, tt.path, bytes.NewReader(reqBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Call appropriate handler method based on the test case
			switch tt.name {
			case "CheckDependencies":
				err = h.CheckDependencies(c)
			case "Start":
				err = h.Start(c)
			case "Stop":
				err = h.Stop(c)
			case "Status":
				err = h.Status(c)
			case "GetKubeConfig":
				err = h.GetKubeConfig(c)
			case "Clean":
				err = h.Clean(c)
			}

			if err != nil {
				t.Fatalf("Handler returned error: %v", err)
			}

			if rec.Code != tt.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tt.expectedStatus, rec.Code)
			}
		})
	}
}
