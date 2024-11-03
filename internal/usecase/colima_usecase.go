package usecase

import (
	"context"

	"github.com/gqadonis/colima-manager/internal/domain"
	"github.com/gqadonis/colima-manager/internal/pkg/logger"
)

type ColimaUseCaseInterface interface {
	CheckDependencies(ctx context.Context) (*domain.DependencyStatus, error)
	UpdateDependencies(ctx context.Context) error
	Start(ctx context.Context, config domain.ColimaConfig) error
	Stop(ctx context.Context, profile string) error
	Status(ctx context.Context, profile string) (*domain.ColimaStatus, error)
	GetKubeConfig(ctx context.Context, profile string) (string, error)
	Clean(ctx context.Context, req domain.CleanRequest) error
}

type ColimaUseCase struct {
	repo domain.ColimaRepository
	log  *logger.Logger
}

func NewColimaUseCase(repo domain.ColimaRepository) ColimaUseCaseInterface {
	return &ColimaUseCase{
		repo: repo,
		log:  logger.GetLogger(),
	}
}

func (uc *ColimaUseCase) CheckDependencies(ctx context.Context) (*domain.DependencyStatus, error) {
	uc.log.Info("Checking dependencies in usecase")
	status, err := uc.repo.CheckDependencies(ctx)
	if err != nil {
		return nil, uc.log.LogError(err, "dependency check failed in usecase")
	}
	uc.log.Info("Dependencies checked successfully - Homebrew: %v, Colima: %v, Lima: %v",
		status.Homebrew, status.Colima, status.Lima)
	return status, nil
}

func (uc *ColimaUseCase) UpdateDependencies(ctx context.Context) error {
	uc.log.Info("Updating dependencies in usecase")
	if err := uc.repo.UpdateDependencies(ctx); err != nil {
		return uc.log.LogError(err, "failed to update dependencies in usecase")
	}
	uc.log.Info("Dependencies updated successfully")
	return nil
}

func (uc *ColimaUseCase) Start(ctx context.Context, config domain.ColimaConfig) error {
	uc.log.Info("Starting Colima instance with config: %+v", config)

	// Apply defaults if not set
	defaults := domain.DefaultColimaConfig()

	if config.Profile == "" {
		config.Profile = defaults.Profile
		uc.log.Debug("Using default profile: %s", config.Profile)
	}
	if config.CPUs == 0 {
		config.CPUs = defaults.CPUs
		uc.log.Debug("Using default CPUs: %d", config.CPUs)
	}
	if config.Memory == 0 {
		config.Memory = defaults.Memory
		uc.log.Debug("Using default Memory: %d", config.Memory)
	}
	if config.DiskSize == 0 {
		config.DiskSize = defaults.DiskSize
		uc.log.Debug("Using default DiskSize: %d", config.DiskSize)
	}
	if config.VMType == "" {
		config.VMType = defaults.VMType
		uc.log.Debug("Using default VMType: %s", config.VMType)
	}
	if config.Runtime == "" {
		config.Runtime = defaults.Runtime
		uc.log.Debug("Using default Runtime: %s", config.Runtime)
	}

	// Check dependencies before starting
	uc.log.Debug("Checking dependencies before start")
	status, err := uc.repo.CheckDependencies(ctx)
	if err != nil {
		return uc.log.LogError(err, "dependency check failed before start")
	}

	// If dependencies are missing or outdated, try to update them
	if !status.Colima || !status.Lima {
		uc.log.Info("Missing dependencies detected, attempting update")
		if err := uc.repo.UpdateDependencies(ctx); err != nil {
			return uc.log.LogError(err, "failed to update dependencies before start")
		}

		// Check again after update
		uc.log.Debug("Verifying dependencies after update")
		status, err = uc.repo.CheckDependencies(ctx)
		if err != nil {
			return uc.log.LogError(err, "dependency check failed after update")
		}
		if !status.Colima || !status.Lima {
			return uc.log.LogError(&domain.DependencyError{
				Dependency: "colima/lima",
				Reason:     "failed to install required dependencies",
			}, "dependencies still missing after update")
		}
	}

	if err := uc.repo.Start(ctx, config); err != nil {
		return uc.log.LogError(err, "failed to start Colima instance")
	}

	uc.log.Info("Colima instance started successfully - Profile: %s", config.Profile)
	return nil
}

func (uc *ColimaUseCase) Stop(ctx context.Context, profile string) error {
	uc.log.Info("Stopping Colima instance - Profile: %s", profile)

	if profile == "" {
		profile = domain.DefaultColimaConfig().Profile
		uc.log.Debug("Using default profile: %s", profile)
	}

	if err := uc.repo.Stop(ctx, profile); err != nil {
		return uc.log.LogError(err, "failed to stop Colima instance")
	}

	uc.log.Info("Colima instance stopped successfully - Profile: %s", profile)
	return nil
}

func (uc *ColimaUseCase) Status(ctx context.Context, profile string) (*domain.ColimaStatus, error) {
	uc.log.Info("Checking Colima status - Profile: %s", profile)

	if profile == "" {
		profile = domain.DefaultColimaConfig().Profile
		uc.log.Debug("Using default profile: %s", profile)
	}

	status, err := uc.repo.Status(ctx, profile)
	if err != nil {
		return nil, uc.log.LogError(err, "failed to get Colima status")
	}

	uc.log.Info("Colima status retrieved successfully - Profile: %s, Status: %+v", profile, status)
	return status, nil
}

func (uc *ColimaUseCase) GetKubeConfig(ctx context.Context, profile string) (string, error) {
	uc.log.Info("Getting kubeconfig - Profile: %s", profile)

	if profile == "" {
		profile = domain.DefaultColimaConfig().Profile
		uc.log.Debug("Using default profile: %s", profile)
	}

	kubeconfig, err := uc.repo.GetKubeConfig(ctx, profile)
	if err != nil {
		return "", uc.log.LogError(err, "failed to get kubeconfig")
	}

	uc.log.Info("Kubeconfig retrieved successfully - Profile: %s", profile)
	return kubeconfig, nil
}

func (uc *ColimaUseCase) Clean(ctx context.Context, req domain.CleanRequest) error {
	uc.log.Info("Cleaning Colima resources - Profile: %s", req.Profile)

	if err := uc.repo.Clean(ctx, req); err != nil {
		return uc.log.LogError(err, "failed to clean Colima resources")
	}

	if req.Profile == "" {
		uc.log.Info("All Colima resources cleaned successfully")
	} else {
		uc.log.Info("Colima resources cleaned successfully - Profile: %s", req.Profile)
	}
	return nil
}
