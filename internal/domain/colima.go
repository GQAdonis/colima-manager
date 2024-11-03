package domain

import (
	"context"
	"fmt"
)

// DependencyStatus represents the status of required dependencies
type DependencyStatus struct {
	Homebrew      bool   `json:"homebrew"`
	HomebrewPath  string `json:"homebrew_path,omitempty"`
	Colima        bool   `json:"colima"`
	ColimaVersion string `json:"colima_version,omitempty"`
	ColimaPath    string `json:"colima_path,omitempty"`
	Lima          bool   `json:"lima"`
	LimaVersion   string `json:"lima_version,omitempty"`
}

// ColimaStatus represents the status of Colima
type ColimaStatus struct {
	Status     string `json:"status"`
	CPUs       int    `json:"cpus"`
	Memory     int    `json:"memory"`
	DiskSize   int    `json:"disk_size"`
	Kubernetes bool   `json:"kubernetes"`
	Profile    string `json:"profile"`
}

// CleanRequest represents the clean operation parameters
type CleanRequest struct {
	Profile string `json:"profile"` // empty string means clean all
}

// DockerContext represents a Docker context configuration
type DockerContext struct {
	Name    string `json:"name"`
	Profile string `json:"profile"`
	Socket  string `json:"socket"`
}

// Custom error types
type ProfileNotFoundError struct {
	Profile string
}

func (e *ProfileNotFoundError) Error() string {
	return fmt.Sprintf("profile '%s' does not exist", e.Profile)
}

type ProfileNotStartedError struct {
	Profile string
}

func (e *ProfileNotStartedError) Error() string {
	return fmt.Sprintf("profile '%s' is not started", e.Profile)
}

type ProfileUnreachableError struct {
	Profile string
	Reason  string
}

func (e *ProfileUnreachableError) Error() string {
	return fmt.Sprintf("profile '%s' is unreachable: %s", e.Profile, e.Reason)
}

type ProfileMalfunctionError struct {
	Profile string
	Reason  string
}

func (e *ProfileMalfunctionError) Error() string {
	return fmt.Sprintf("profile '%s' is malfunctioning: %s", e.Profile, e.Reason)
}

type DependencyError struct {
	Dependency string
	Reason     string
}

func (e *DependencyError) Error() string {
	return fmt.Sprintf("%s dependency error: %s", e.Dependency, e.Reason)
}

type DockerContextError struct {
	Operation string
	Profile   string
	Reason    string
}

func (e *DockerContextError) Error() string {
	return fmt.Sprintf("docker context %s failed for profile '%s': %s", e.Operation, e.Profile, e.Reason)
}

// ColimaRepository defines the interface for Colima operations
type ColimaRepository interface {
	Start(ctx context.Context, config ColimaConfig) error
	Stop(ctx context.Context, profile string) error
	StopDaemon(ctx context.Context) error
	Status(ctx context.Context, profile string) (*ColimaStatus, error)
	GetKubeConfig(ctx context.Context, profile string) (string, error)
	Clean(ctx context.Context, req CleanRequest) error
	CheckDependencies(ctx context.Context) (*DependencyStatus, error)
	UpdateDependencies(ctx context.Context) error
	CreateDockerContext(ctx context.Context, profile string) error
	RemoveDockerContext(ctx context.Context, profile string) error
	ListDockerContexts(ctx context.Context) ([]DockerContext, error)
}

// ColimaConfig represents the configuration for starting Colima
type ColimaConfig struct {
	CPUs           int    `json:"cpus,omitempty"`
	Memory         int    `json:"memory,omitempty"`
	DiskSize       int    `json:"disk_size,omitempty"`
	VMType         string `json:"vm_type,omitempty"`
	Runtime        string `json:"runtime,omitempty"`
	NetworkAddress bool   `json:"network_address"`
	Kubernetes     bool   `json:"kubernetes"`
	Profile        string `json:"profile,omitempty"`
}

// DefaultColimaConfig returns a configuration with default values
func DefaultColimaConfig() ColimaConfig {
	return ColimaConfig{
		CPUs:           12,
		Memory:         32,
		DiskSize:       100,
		VMType:         "vz",
		Runtime:        "containerd",
		NetworkAddress: true,
		Kubernetes:     true,
		Profile:        "default",
	}
}
