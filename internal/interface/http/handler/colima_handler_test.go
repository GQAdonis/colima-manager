package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gqadonis/colima-manager/internal/domain"
	"github.com/gqadonis/colima-manager/internal/pkg/logger"
	"github.com/gqadonis/colima-manager/internal/usecase"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockColimaUseCase is a mock implementation of the usecase interface
type MockColimaUseCase struct {
	mock.Mock
}

func (m *MockColimaUseCase) CheckDependencies(ctx context.Context) (*domain.DependencyStatus, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.DependencyStatus), args.Error(1)
}

func (m *MockColimaUseCase) UpdateDependencies(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockColimaUseCase) Start(ctx context.Context, config domain.ColimaConfig) error {
	args := m.Called(ctx, config)
	return args.Error(0)
}

func (m *MockColimaUseCase) Stop(ctx context.Context, profile string) error {
	args := m.Called(ctx, profile)
	return args.Error(0)
}

func (m *MockColimaUseCase) Status(ctx context.Context, profile string) (*domain.ColimaStatus, error) {
	args := m.Called(ctx, profile)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ColimaStatus), args.Error(1)
}

func (m *MockColimaUseCase) GetKubeConfig(ctx context.Context, profile string) (string, error) {
	args := m.Called(ctx, profile)
	return args.String(0), args.Error(1)
}

func (m *MockColimaUseCase) Clean(ctx context.Context, req domain.CleanRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func setupTest() (*echo.Echo, usecase.ColimaUseCaseInterface) {
	e := echo.New()
	mockUseCase := new(MockColimaUseCase)
	return e, mockUseCase
}

func TestCheckDependencies(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*MockColimaUseCase)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name: "successful dependency check",
			setupMock: func(m *MockColimaUseCase) {
				m.On("CheckDependencies", mock.Anything).Return(&domain.DependencyStatus{
					Homebrew: true,
					Colima:   true,
					Lima:     true,
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"homebrew": true,
				"colima":   true,
				"lima":     true,
			},
		},
		{
			name: "dependency error",
			setupMock: func(m *MockColimaUseCase) {
				m.On("CheckDependencies", mock.Anything).Return(nil,
					&domain.DependencyError{Dependency: "homebrew", Reason: "not installed"})
			},
			expectedStatus: http.StatusServiceUnavailable,
			expectedBody: map[string]interface{}{
				"error": "homebrew dependency error: not installed",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e, mockUseCase := setupTest()
			tt.setupMock(mockUseCase.(*MockColimaUseCase))

			h := &ColimaHandler{
				useCase: mockUseCase,
				log:     logger.GetLogger(),
			}

			req := httptest.NewRequest(http.MethodGet, "/dependencies", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := h.CheckDependencies(c)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			var response map[string]interface{}
			err = json.Unmarshal(rec.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedBody, response)
		})
	}
}

func TestStart(t *testing.T) {
	tests := []struct {
		name           string
		config         domain.ColimaConfig
		setupMock      func(*MockColimaUseCase)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name: "successful start",
			config: domain.ColimaConfig{
				CPUs:       4,
				Memory:     8,
				DiskSize:   60,
				Profile:    "default",
				Kubernetes: true,
			},
			setupMock: func(m *MockColimaUseCase) {
				m.On("Start", mock.Anything, mock.AnythingOfType("domain.ColimaConfig")).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"status": "started",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e, mockUseCase := setupTest()
			tt.setupMock(mockUseCase.(*MockColimaUseCase))

			h := &ColimaHandler{
				useCase: mockUseCase,
				log:     logger.GetLogger(),
			}

			jsonBody := strings.NewReader(`{"cpus":4,"memory":8,"disk_size":60,"profile":"default","kubernetes":true}`)
			req := httptest.NewRequest(http.MethodPost, "/start", jsonBody)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := h.Start(c)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			var response map[string]interface{}
			err = json.Unmarshal(rec.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedBody, response)
		})
	}
}
