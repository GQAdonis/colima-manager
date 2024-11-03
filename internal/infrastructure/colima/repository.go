package colima

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gqadonis/colima-manager/internal/domain"
	"github.com/gqadonis/colima-manager/internal/pkg/logger"
)

type ColimaRepository struct {
	homeDir string
	log     *logger.Logger
	exec    Executor
}

func NewColimaRepository() (*ColimaRepository, error) {
	log := logger.GetLogger()
	log.Info("Initializing Colima repository")

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, log.LogError(err, "failed to get home directory")
	}

	repo := &ColimaRepository{
		homeDir: homeDir,
		log:     log,
		exec:    NewRealExecutor(),
	}

	log.Info("Colima repository initialized with home directory: %s", homeDir)
	return repo, nil
}

func (r *ColimaRepository) CheckDependencies(ctx context.Context) (*domain.DependencyStatus, error) {
	r.log.Info("Checking dependencies")
	status := &domain.DependencyStatus{}

	// Check Homebrew
	brewPath, err := r.exec.Command("brew", "--prefix").Output()
	if err == nil {
		status.Homebrew = true
		status.HomebrewPath = strings.TrimSpace(string(brewPath))
		r.log.Debug("Homebrew found at: %s", status.HomebrewPath)
	} else {
		r.log.Error("Homebrew not found: %v", err)
	}

	if !status.Homebrew {
		return status, r.log.LogError(&domain.DependencyError{
			Dependency: "homebrew",
			Reason:     "not installed or not in PATH",
		}, "homebrew dependency check failed")
	}

	// Check Colima
	colimaPath, err := r.exec.Command("which", "colima").Output()
	if err == nil {
		status.Colima = true
		status.ColimaPath = strings.TrimSpace(string(colimaPath))
		r.log.Debug("Colima found at: %s", status.ColimaPath)

		// Get Colima version
		if out, err := r.exec.Command("colima", "version").Output(); err == nil {
			status.ColimaVersion = strings.TrimSpace(string(out))
			r.log.Debug("Colima version: %s", status.ColimaVersion)
		} else {
			r.log.Error("Failed to get Colima version: %v", err)
		}
	} else {
		r.log.Error("Colima not found: %v", err)
	}

	// Check Lima version using brew
	cmd := r.exec.Command("brew", "list", "--versions", "lima")
	if out, err := cmd.Output(); err == nil {
		parts := strings.Fields(string(out))
		if len(parts) >= 2 {
			status.Lima = true
			status.LimaVersion = parts[1]
			r.log.Debug("Lima version: %s", status.LimaVersion)
		}
	} else {
		r.log.Error("Failed to get Lima version: %v", err)
	}

	r.log.Info("Dependency check completed - Homebrew: %v, Colima: %v, Lima: %v",
		status.Homebrew, status.Colima, status.Lima)
	return status, nil
}

func (r *ColimaRepository) UpdateDependencies(ctx context.Context) error {
	r.log.Info("Updating dependencies")

	// Update Homebrew first
	r.log.Debug("Updating Homebrew")
	cmd := r.exec.Command("brew", "update")
	if err := cmd.Run(); err != nil {
		return r.log.LogError(&domain.DependencyError{
			Dependency: "homebrew",
			Reason:     fmt.Sprintf("failed to update: %v", err),
		}, "homebrew update failed")
	}

	// Upgrade Colima and Lima
	r.log.Debug("Upgrading Colima and Lima")
	cmd = r.exec.Command("brew", "upgrade", "colima", "lima")
	if err := cmd.Run(); err != nil {
		return r.log.LogError(&domain.DependencyError{
			Dependency: "colima/lima",
			Reason:     fmt.Sprintf("failed to upgrade: %v", err),
		}, "colima/lima upgrade failed")
	}

	r.log.Info("Dependencies updated successfully")
	return nil
}

func (r *ColimaRepository) Start(ctx context.Context, config domain.ColimaConfig) error {
	r.log.Info("Starting Colima with config: %+v", config)

	args := []string{
		"start",
		"--cpu", fmt.Sprintf("%d", config.CPUs),
		"--memory", fmt.Sprintf("%d", config.Memory),
		"--disk", fmt.Sprintf("%d", config.DiskSize),
		"--vm-type", config.VMType,
		"--runtime", config.Runtime,
	}

	if config.NetworkAddress {
		args = append(args, "--network-address")
	}

	if config.Kubernetes {
		args = append(args, "--kubernetes")
	}

	if config.Profile != "" && config.Profile != "default" {
		args = append(args, "-p", config.Profile)
	}

	r.log.Debug("Executing colima command with args: %v", args)
	cmd := r.exec.Command("colima", args...)

	if output, err := cmd.CombinedOutput(); err != nil {
		return r.log.LogError(err, "failed to start colima: %s", string(output))
	}

	r.log.Info("Colima started successfully - Profile: %s", config.Profile)
	return nil
}

func (r *ColimaRepository) Stop(ctx context.Context, profile string) error {
	r.log.Info("Stopping Colima profile: %s", profile)

	if !r.checkProfileExists(profile) {
		return r.log.LogError(&domain.ProfileNotFoundError{Profile: profile},
			"profile not found during stop")
	}

	args := []string{"stop"}
	if profile != "" && profile != "default" {
		args = append(args, "-p", profile)
	}

	r.log.Debug("Executing colima stop command with args: %v", args)
	cmd := r.exec.Command("colima", args...)

	if output, err := cmd.CombinedOutput(); err != nil {
		return r.log.LogError(err, "failed to stop colima: %s", string(output))
	}

	r.log.Info("Colima stopped successfully - Profile: %s", profile)
	return nil
}

func (r *ColimaRepository) Status(ctx context.Context, profile string) (*domain.ColimaStatus, error) {
	r.log.Info("Checking status for profile: %s", profile)

	if !r.checkProfileExists(profile) {
		return nil, r.log.LogError(&domain.ProfileNotFoundError{Profile: profile},
			"profile not found during status check")
	}

	args := []string{"status", "-e"}
	if profile != "" && profile != "default" {
		args = append(args, "-p", profile)
	}

	r.log.Debug("Executing colima status command with args: %v", args)
	cmd := r.exec.Command("colima", args...)
	output, err := cmd.CombinedOutput()

	outputStr := string(output)
	r.log.Debug("Colima status output: %s", outputStr)

	if err != nil {
		if strings.Contains(outputStr, "is not running") {
			return nil, r.log.LogError(&domain.ProfileNotStartedError{Profile: profile},
				"profile is not running")
		}

		if strings.Contains(outputStr, "connection refused") ||
			strings.Contains(outputStr, "cannot connect") {
			return nil, r.log.LogError(&domain.ProfileUnreachableError{
				Profile: profile,
				Reason:  "connection to VM failed",
			}, "profile is unreachable")
		}

		return nil, r.log.LogError(&domain.ProfileMalfunctionError{
			Profile: profile,
			Reason:  outputStr,
		}, "profile malfunction")
	}

	// Parse the output to create ColimaStatus
	status := &domain.ColimaStatus{
		Profile: profile,
	}

	// Basic parsing of the output
	lines := strings.Split(outputStr, "\n")
	for _, line := range lines {
		if strings.Contains(line, "running") {
			status.Status = "running"
		}
		if strings.Contains(line, "CPU:") {
			fmt.Sscanf(line, "CPU: %d", &status.CPUs)
		}
		if strings.Contains(line, "Memory:") {
			fmt.Sscanf(line, "Memory: %d", &status.Memory)
		}
		if strings.Contains(line, "Disk:") {
			fmt.Sscanf(line, "Disk: %d", &status.DiskSize)
		}
		if strings.Contains(line, "kubernetes") {
			status.Kubernetes = true
		}
	}

	r.log.Info("Status check completed successfully - Profile: %s, Status: %+v", profile, status)
	return status, nil
}

func (r *ColimaRepository) GetKubeConfig(ctx context.Context, profile string) (string, error) {
	r.log.Info("Getting kubeconfig for profile: %s", profile)

	if !r.checkProfileExists(profile) {
		return "", r.log.LogError(&domain.ProfileNotFoundError{Profile: profile},
			"profile not found during kubeconfig retrieval")
	}

	configName := "colima.kubeconfig"
	if profile != "" && profile != "default" {
		configName = fmt.Sprintf("colima-%s.kubeconfig", profile)
	}

	colimaKubeConfig := filepath.Join(r.homeDir, ".colima", configName)
	r.log.Debug("Reading kubeconfig from: %s", colimaKubeConfig)

	data, err := os.ReadFile(colimaKubeConfig)
	if err != nil {
		return "", r.log.LogError(err, "failed to read kubeconfig")
	}

	r.log.Info("Kubeconfig retrieved successfully - Profile: %s", profile)
	return string(data), nil
}

func (r *ColimaRepository) Clean(ctx context.Context, req domain.CleanRequest) error {
	r.log.Info("Starting cleanup - Profile: %s", req.Profile)

	// If cleaning specific profile
	if req.Profile != "" {
		if !r.checkProfileExists(req.Profile) {
			return r.log.LogError(&domain.ProfileNotFoundError{Profile: req.Profile},
				"profile not found during cleanup")
		}

		// Stop the specific profile
		cmd := r.exec.Command("colima", "stop", "-p", req.Profile)
		if output, err := cmd.CombinedOutput(); err != nil {
			r.log.Debug("Error stopping profile (non-fatal): %s", string(output))
		}

		// Delete the specific profile
		cmd = r.exec.Command("colima", "delete", "-p", req.Profile, "-f")
		if output, err := cmd.CombinedOutput(); err != nil {
			return r.log.LogError(err, "failed to delete profile %s: %s", req.Profile, string(output))
		}

		// Clean up profile-specific directories
		profileDirs := []string{
			filepath.Join(r.homeDir, ".colima", req.Profile),
			filepath.Join(r.homeDir, ".lima", req.Profile),
		}

		for _, dir := range profileDirs {
			r.log.Debug("Removing directory: %s", dir)
			if err := os.RemoveAll(dir); err != nil {
				return r.log.LogError(err, "failed to remove directory: %s", dir)
			}
		}

		r.log.Info("Profile cleaned successfully: %s", req.Profile)
		return nil
	}

	// Cleaning all profiles
	r.log.Debug("Cleaning all profiles")

	// Stop all running instances
	cmd := r.exec.Command("colima", "stop")
	if output, err := cmd.CombinedOutput(); err != nil {
		r.log.Debug("Error stopping instances (non-fatal): %s", string(output))
	}

	// Delete all instances
	cmd = r.exec.Command("colima", "delete", "-f")
	if output, err := cmd.CombinedOutput(); err != nil {
		return r.log.LogError(err, "failed to delete all instances: %s", string(output))
	}

	// Clean up all colima-related directories
	dirsToDelete := []string{
		filepath.Join(r.homeDir, ".colima"),
		filepath.Join(r.homeDir, ".lima"),
	}

	for _, dir := range dirsToDelete {
		r.log.Debug("Removing directory: %s", dir)
		if err := os.RemoveAll(dir); err != nil {
			return r.log.LogError(err, "failed to remove directory: %s", dir)
		}
	}

	r.log.Info("All profiles cleaned successfully")
	return nil
}

func (r *ColimaRepository) checkProfileExists(profile string) bool {
	profilePath := filepath.Join(r.homeDir, ".colima", profile)
	_, err := os.Stat(profilePath)
	exists := err == nil
	r.log.Debug("Checking profile existence - Profile: %s, Path: %s, Exists: %v",
		profile, profilePath, exists)
	return exists
}

func (r *ColimaRepository) CreateDockerContext(ctx context.Context, profile string) error {
	r.log.Info("Creating Docker context for profile: %s", profile)

	// Determine socket path based on profile
	var socketPath string
	if profile == "default" {
		socketPath = "/var/run/docker.sock"
	} else {
		socketPath = fmt.Sprintf("/tmp/colima-%s.sock", profile)
	}

	// Create context name
	contextName := "colima"
	if profile != "default" {
		contextName = fmt.Sprintf("colima-%s", profile)
	}

	// Create new context
	cmd := r.exec.Command("docker", "context", "create",
		contextName,
		"--docker", fmt.Sprintf("host=unix://%s", socketPath))

	output, err := cmd.CombinedOutput()
	if err != nil {
		return r.log.LogError(&domain.DockerContextError{
			Operation: "create",
			Profile:   profile,
			Reason:    fmt.Sprintf("failed to create context: %v - %s", err, string(output)),
		}, "docker context creation failed")
	}

	r.log.Info("Docker context created successfully - Profile: %s, Context: %s", profile, contextName)
	return nil
}

func (r *ColimaRepository) RemoveDockerContext(ctx context.Context, profile string) error {
	r.log.Info("Removing Docker context for profile: %s", profile)

	contextName := "colima"
	if profile != "default" {
		contextName = fmt.Sprintf("colima-%s", profile)
	}

	cmd := r.exec.Command("docker", "context", "rm", "-f", contextName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// If the context doesn't exist, we don't treat it as an error
		if strings.Contains(string(output), "not found") {
			r.log.Debug("Docker context %s not found, skipping removal", contextName)
			return nil
		}
		return r.log.LogError(&domain.DockerContextError{
			Operation: "remove",
			Profile:   profile,
			Reason:    fmt.Sprintf("failed to remove context: %v - %s", err, string(output)),
		}, "docker context removal failed")
	}

	r.log.Info("Docker context removed successfully - Profile: %s, Context: %s", profile, contextName)
	return nil
}

func (r *ColimaRepository) ListDockerContexts(ctx context.Context) ([]domain.DockerContext, error) {
	r.log.Info("Listing Docker contexts")

	cmd := r.exec.Command("docker", "context", "ls", "--format", "{{.Name}}")
	output, err := cmd.Output()
	if err != nil {
		return nil, r.log.LogError(err, "failed to list Docker contexts")
	}

	contexts := []domain.DockerContext{}
	for _, line := range strings.Split(string(output), "\n") {
		if strings.HasPrefix(line, "colima") {
			profile := "default"
			if line != "colima" {
				profile = strings.TrimPrefix(line, "colima-")
			}

			socketPath := "/var/run/docker.sock"
			if profile != "default" {
				socketPath = fmt.Sprintf("/tmp/colima-%s.sock", profile)
			}

			contexts = append(contexts, domain.DockerContext{
				Name:    line,
				Profile: profile,
				Socket:  socketPath,
			})
		}
	}

	r.log.Info("Found %d Colima Docker contexts", len(contexts))
	return contexts, nil
}
