package usecase

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/gqadonis/colima-manager/internal/domain"
)

type mockRepository struct {
	startCalled       bool
	startConfig       domain.ColimaConfig
	statusCalled      bool
	statusProfile     string
	kubeConfigCalled  bool
	kubeConfigProfile string
	mockStatus        *domain.ColimaStatus
	mockError         error
	mu                sync.Mutex // protect concurrent access to mock fields
}

func (m *mockRepository) Start(ctx context.Context, config domain.ColimaConfig) error {
	m.mu.Lock()
	m.startCalled = true
	m.startConfig = config
	m.mu.Unlock()
	// Simulate some work
	time.Sleep(100 * time.Millisecond)
	return m.mockError
}

func (m *mockRepository) Stop(ctx context.Context, profile string) error {
	time.Sleep(100 * time.Millisecond)
	return m.mockError
}

func (m *mockRepository) StopDaemon(ctx context.Context) error {
	return m.mockError
}

func (m *mockRepository) Status(ctx context.Context, profile string) (*domain.ColimaStatus, error) {
	m.mu.Lock()
	m.statusCalled = true
	m.statusProfile = profile
	m.mu.Unlock()
	return m.mockStatus, m.mockError
}

func (m *mockRepository) GetKubeConfig(ctx context.Context, profile string) (string, error) {
	m.mu.Lock()
	m.kubeConfigCalled = true
	m.kubeConfigProfile = profile
	m.mu.Unlock()
	return "", m.mockError
}

func (m *mockRepository) Clean(ctx context.Context, req domain.CleanRequest) error {
	time.Sleep(100 * time.Millisecond)
	return m.mockError
}

func (m *mockRepository) CheckDependencies(ctx context.Context) (*domain.DependencyStatus, error) {
	return &domain.DependencyStatus{
		Homebrew: true,
		Colima:   true,
		Lima:     true,
	}, m.mockError
}

func (m *mockRepository) UpdateDependencies(ctx context.Context) error {
	return m.mockError
}

func (m *mockRepository) CreateDockerContext(ctx context.Context, profile string) error {
	return m.mockError
}

func (m *mockRepository) RemoveDockerContext(ctx context.Context, profile string) error {
	return m.mockError
}

func (m *mockRepository) ListDockerContexts(ctx context.Context) ([]domain.DockerContext, error) {
	return nil, m.mockError
}

func TestStartupSequence(t *testing.T) {
	// Reset profile lock before test
	domain.ResetProfileLock()

	// Create mock repository
	mockRepo := &mockRepository{
		mockStatus: &domain.ColimaStatus{
			Status:     "Running",
			CPUs:       4,
			Memory:     8,
			DiskSize:   60,
			Kubernetes: true,
			Profile:    "default",
		},
	}

	// Create use case with mock repository
	useCase := NewColimaUseCase(mockRepo)

	// Test configuration
	config := domain.ColimaConfig{
		CPUs:           4,
		Memory:         8,
		DiskSize:       60,
		VMType:         "vz",
		Runtime:        "containerd",
		NetworkAddress: true,
		Kubernetes:     true,
		Profile:        "default",
	}

	// Start the profile
	err := useCase.Start(context.Background(), config)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify Start was called with correct config
	mockRepo.mu.Lock()
	if !mockRepo.startCalled {
		t.Error("Expected Start to be called")
	}
	if mockRepo.startConfig != config {
		t.Errorf("Expected config %+v, got %+v", config, mockRepo.startConfig)
	}
	mockRepo.mu.Unlock()

	// Check status
	status, err := useCase.Status(context.Background(), config.Profile)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify Status was called with correct profile
	mockRepo.mu.Lock()
	if !mockRepo.statusCalled {
		t.Error("Expected Status to be called")
	}
	if mockRepo.statusProfile != config.Profile {
		t.Errorf("Expected profile %s, got %s", config.Profile, mockRepo.statusProfile)
	}
	mockRepo.mu.Unlock()

	// Verify status values
	if status.Status != "Running" {
		t.Errorf("Expected status Running, got %s", status.Status)
	}
	if status.CPUs != config.CPUs {
		t.Errorf("Expected CPUs %d, got %d", config.CPUs, status.CPUs)
	}
	if status.Memory != config.Memory {
		t.Errorf("Expected Memory %d, got %d", config.Memory, status.Memory)
	}
	if status.DiskSize != config.DiskSize {
		t.Errorf("Expected DiskSize %d, got %d", config.DiskSize, status.DiskSize)
	}
	if status.Kubernetes != config.Kubernetes {
		t.Errorf("Expected Kubernetes %v, got %v", config.Kubernetes, status.Kubernetes)
	}

	// Verify kubeconfig check for Kubernetes-enabled profile
	if config.Kubernetes {
		_, err = useCase.GetKubeConfig(context.Background(), config.Profile)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		mockRepo.mu.Lock()
		if !mockRepo.kubeConfigCalled {
			t.Error("Expected GetKubeConfig to be called for Kubernetes-enabled profile")
		}
		if mockRepo.kubeConfigProfile != config.Profile {
			t.Errorf("Expected profile %s, got %s", config.Profile, mockRepo.kubeConfigProfile)
		}
		mockRepo.mu.Unlock()
	}
}

func TestProfileLocking(t *testing.T) {
	// Reset profile lock before test
	domain.ResetProfileLock()

	mockRepo := &mockRepository{}
	useCase1 := NewColimaUseCase(mockRepo)
	useCase2 := NewColimaUseCase(mockRepo)

	// Test configuration
	config := domain.ColimaConfig{
		Profile: "test-profile",
		CPUs:    4,
		Memory:  8,
	}

	// Use a WaitGroup to coordinate goroutines
	var wg sync.WaitGroup
	wg.Add(2)

	// Channel to collect errors
	errChan := make(chan error, 2)

	// Start operation with first use case
	go func() {
		defer wg.Done()
		if err := useCase1.Start(context.Background(), config); err != nil {
			errChan <- err
		}
	}()

	// Small delay to ensure first operation has started
	time.Sleep(50 * time.Millisecond)

	// Attempt concurrent operation with second use case
	go func() {
		defer wg.Done()
		if err := useCase2.Start(context.Background(), config); err != nil {
			errChan <- err
		}
	}()

	// Wait for both operations to complete
	wg.Wait()
	close(errChan)

	// Collect errors
	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}

	// Verify that exactly one error occurred (the blocked operation)
	if len(errors) != 1 {
		t.Errorf("Expected exactly one error, got %d", len(errors))
	}

	// Verify that the error is ProfileBusyError
	if len(errors) > 0 {
		if _, ok := errors[0].(*domain.ProfileBusyError); !ok {
			t.Errorf("Expected ProfileBusyError, got %T", errors[0])
		}
	}

	// Test read operations (should work even when locked)
	_, err := useCase2.Status(context.Background(), config.Profile)
	if err != nil {
		t.Errorf("Expected no error for Status, got %v", err)
	}

	_, err = useCase2.GetKubeConfig(context.Background(), config.Profile)
	if err != nil {
		t.Errorf("Expected no error for GetKubeConfig, got %v", err)
	}

	// Reset lock and verify we can start again
	domain.ResetProfileLock()

	err = useCase2.Start(context.Background(), config)
	if err != nil {
		t.Errorf("Expected no error after lock release, got %v", err)
	}
}
