package usecase

import (
	"context"
	"fmt"
	"testing"

	"github.com/gqadonis/colima-manager/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockColimaRepository is a mock implementation of the repository
type MockColimaRepository struct {
	mock.Mock
}

func (m *MockColimaRepository) Start(ctx context.Context, config domain.ColimaConfig) error {
	args := m.Called(ctx, config)
	return args.Error(0)
}

func (m *MockColimaRepository) Stop(ctx context.Context, profile string) error {
	args := m.Called(ctx, profile)
	return args.Error(0)
}

func (m *MockColimaRepository) StopDaemon(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockColimaRepository) Status(ctx context.Context, profile string) (*domain.ColimaStatus, error) {
	args := m.Called(ctx, profile)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ColimaStatus), args.Error(1)
}

func (m *MockColimaRepository) GetKubeConfig(ctx context.Context, profile string) (string, error) {
	args := m.Called(ctx, profile)
	return args.String(0), args.Error(1)
}

func (m *MockColimaRepository) Clean(ctx context.Context, req domain.CleanRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockColimaRepository) CheckDependencies(ctx context.Context) (*domain.DependencyStatus, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.DependencyStatus), args.Error(1)
}

func (m *MockColimaRepository) UpdateDependencies(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockColimaRepository) CreateDockerContext(ctx context.Context, profile string) error {
	args := m.Called(ctx, profile)
	return args.Error(0)
}

func (m *MockColimaRepository) RemoveDockerContext(ctx context.Context, profile string) error {
	args := m.Called(ctx, profile)
	return args.Error(0)
}

func (m *MockColimaRepository) ListDockerContexts(ctx context.Context) ([]domain.DockerContext, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.DockerContext), args.Error(1)
}

func TestStop(t *testing.T) {
	tests := []struct {
		name          string
		profile       string
		setupMock     func(*MockColimaRepository)
		expectedError error
	}{
		{
			name:    "successful stop with daemon",
			profile: "default",
			setupMock: func(m *MockColimaRepository) {
				m.On("Stop", mock.Anything, "default").Return(nil)
				m.On("StopDaemon", mock.Anything).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:    "successful stop with custom profile",
			profile: "test-profile",
			setupMock: func(m *MockColimaRepository) {
				m.On("Stop", mock.Anything, "test-profile").Return(nil)
				m.On("StopDaemon", mock.Anything).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:    "failed to stop colima but daemon stops",
			profile: "default",
			setupMock: func(m *MockColimaRepository) {
				m.On("Stop", mock.Anything, "default").Return(&domain.ProfileNotFoundError{Profile: "default"})
				// The daemon stop should still be attempted even if colima stop fails
				m.On("StopDaemon", mock.Anything).Return(nil).Once()
			},
			expectedError: &domain.ProfileNotFoundError{},
		},
		{
			name:    "failed to stop daemon but colima stops",
			profile: "default",
			setupMock: func(m *MockColimaRepository) {
				m.On("Stop", mock.Anything, "default").Return(nil)
				m.On("StopDaemon", mock.Anything).Return(fmt.Errorf("failed to stop daemon"))
			},
			expectedError: nil, // We don't return daemon stop errors
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockColimaRepository)
			tt.setupMock(mockRepo)

			uc := NewColimaUseCase(mockRepo)
			err := uc.Stop(context.Background(), tt.profile)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.IsType(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestStart(t *testing.T) {
	tests := []struct {
		name          string
		config        domain.ColimaConfig
		setupMock     func(*MockColimaRepository)
		expectedError error
	}{
		{
			name:   "successful start with defaults",
			config: domain.ColimaConfig{},
			setupMock: func(m *MockColimaRepository) {
				m.On("CheckDependencies", mock.Anything).Return(&domain.DependencyStatus{
					Homebrew: true,
					Colima:   true,
					Lima:     true,
				}, nil)
				m.On("Start", mock.Anything, mock.MatchedBy(func(config domain.ColimaConfig) bool {
					defaults := domain.DefaultColimaConfig()
					return config.CPUs == defaults.CPUs &&
						config.Memory == defaults.Memory &&
						config.DiskSize == defaults.DiskSize &&
						config.VMType == defaults.VMType &&
						config.Runtime == defaults.Runtime &&
						config.Profile == defaults.Profile
				})).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:   "missing dependencies",
			config: domain.ColimaConfig{},
			setupMock: func(m *MockColimaRepository) {
				m.On("CheckDependencies", mock.Anything).Return(&domain.DependencyStatus{
					Homebrew: true,
					Colima:   false,
					Lima:     false,
				}, nil)
				m.On("UpdateDependencies", mock.Anything).Return(nil)
				m.On("CheckDependencies", mock.Anything).Return(&domain.DependencyStatus{
					Homebrew: true,
					Colima:   false,
					Lima:     false,
				}, nil)
			},
			expectedError: &domain.DependencyError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockColimaRepository)
			tt.setupMock(mockRepo)

			uc := NewColimaUseCase(mockRepo)
			err := uc.Start(context.Background(), tt.config)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.IsType(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestStatus(t *testing.T) {
	tests := []struct {
		name           string
		profile        string
		setupMock      func(*MockColimaRepository)
		expectedStatus *domain.ColimaStatus
		expectedError  error
	}{
		{
			name:    "successful status check",
			profile: "default",
			setupMock: func(m *MockColimaRepository) {
				m.On("Status", mock.Anything, "default").Return(&domain.ColimaStatus{
					Status:     "running",
					CPUs:       4,
					Memory:     8,
					DiskSize:   60,
					Kubernetes: true,
					Profile:    "default",
				}, nil)
			},
			expectedStatus: &domain.ColimaStatus{
				Status:     "running",
				CPUs:       4,
				Memory:     8,
				DiskSize:   60,
				Kubernetes: true,
				Profile:    "default",
			},
			expectedError: nil,
		},
		{
			name:    "profile not found",
			profile: "non-existent",
			setupMock: func(m *MockColimaRepository) {
				m.On("Status", mock.Anything, "non-existent").Return(nil,
					&domain.ProfileNotFoundError{Profile: "non-existent"})
			},
			expectedStatus: nil,
			expectedError:  &domain.ProfileNotFoundError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockColimaRepository)
			tt.setupMock(mockRepo)

			uc := NewColimaUseCase(mockRepo)
			status, err := uc.Status(context.Background(), tt.profile)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.IsType(t, tt.expectedError, err)
				assert.Nil(t, status)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedStatus, status)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestClean(t *testing.T) {
	tests := []struct {
		name          string
		request       domain.CleanRequest
		setupMock     func(*MockColimaRepository)
		expectedError error
	}{
		{
			name: "clean all profiles",
			request: domain.CleanRequest{
				Profile: "",
			},
			setupMock: func(m *MockColimaRepository) {
				m.On("Clean", mock.Anything, mock.MatchedBy(func(req domain.CleanRequest) bool {
					return req.Profile == ""
				})).Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "clean specific profile",
			request: domain.CleanRequest{
				Profile: "test-profile",
			},
			setupMock: func(m *MockColimaRepository) {
				m.On("Clean", mock.Anything, mock.MatchedBy(func(req domain.CleanRequest) bool {
					return req.Profile == "test-profile"
				})).Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "profile not found",
			request: domain.CleanRequest{
				Profile: "non-existent",
			},
			setupMock: func(m *MockColimaRepository) {
				m.On("Clean", mock.Anything, mock.MatchedBy(func(req domain.CleanRequest) bool {
					return req.Profile == "non-existent"
				})).Return(&domain.ProfileNotFoundError{Profile: "non-existent"})
			},
			expectedError: &domain.ProfileNotFoundError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockColimaRepository)
			tt.setupMock(mockRepo)

			uc := NewColimaUseCase(mockRepo)
			err := uc.Clean(context.Background(), tt.request)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.IsType(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestCheckDependencies(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*MockColimaRepository)
		expectedStatus *domain.DependencyStatus
		expectedError  error
	}{
		{
			name: "all dependencies present",
			setupMock: func(m *MockColimaRepository) {
				m.On("CheckDependencies", mock.Anything).Return(&domain.DependencyStatus{
					Homebrew: true,
					Colima:   true,
					Lima:     true,
				}, nil)
			},
			expectedStatus: &domain.DependencyStatus{
				Homebrew: true,
				Colima:   true,
				Lima:     true,
			},
			expectedError: nil,
		},
		{
			name: "missing dependencies",
			setupMock: func(m *MockColimaRepository) {
				m.On("CheckDependencies", mock.Anything).Return(nil,
					&domain.DependencyError{Dependency: "homebrew", Reason: "not installed"})
			},
			expectedStatus: nil,
			expectedError:  &domain.DependencyError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockColimaRepository)
			tt.setupMock(mockRepo)

			uc := NewColimaUseCase(mockRepo)
			status, err := uc.CheckDependencies(context.Background())

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.IsType(t, tt.expectedError, err)
				assert.Nil(t, status)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedStatus, status)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}
