package config

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary config file
	content := []byte(`
server:
  port: 9090
  auto:
    enabled: true
    default: "test-profile"
profiles:
  test-profile:
    cpus: 4
    memory: 8
    disk_size: 60
    vm_type: "vz"
    runtime: "containerd"
    network_address: true
    kubernetes: true
`)
	tmpfile, err := os.CreateTemp("", "config.*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write(content); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	// Temporarily replace config.yaml with our test file
	if err := os.Rename("config.yaml", "config.yaml.bak"); err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	if err := os.Rename(tmpfile.Name(), "config.yaml"); err != nil {
		t.Fatal(err)
	}
	defer func() {
		os.Remove("config.yaml")
		if _, err := os.Stat("config.yaml.bak"); err == nil {
			os.Rename("config.yaml.bak", "config.yaml")
		}
	}()

	// Test loading the config
	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test server configuration
	if config.Server.Port != 9090 {
		t.Errorf("Expected port 9090, got %d", config.Server.Port)
	}

	// Test auto configuration
	if !config.Server.Auto.Enabled {
		t.Error("Expected auto.enabled to be true")
	}
	if config.Server.Auto.Default != "test-profile" {
		t.Errorf("Expected default profile 'test-profile', got '%s'", config.Server.Auto.Default)
	}

	// Test profile configuration
	profile, exists := config.Profiles["test-profile"]
	if !exists {
		t.Fatal("Expected test-profile to exist")
	}

	tests := []struct {
		name     string
		got      interface{}
		expected interface{}
	}{
		{"CPUs", profile.CPUs, 4},
		{"Memory", profile.Memory, 8},
		{"DiskSize", profile.DiskSize, 60},
		{"VMType", profile.VMType, "vz"},
		{"Runtime", profile.Runtime, "containerd"},
		{"NetworkAddress", profile.NetworkAddress, true},
		{"Kubernetes", profile.Kubernetes, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, tt.got)
			}
		})
	}
}

func TestLoadConfigDefaults(t *testing.T) {
	// Temporarily move any existing config file
	if err := os.Rename("config.yaml", "config.yaml.bak"); err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	defer func() {
		if _, err := os.Stat("config.yaml.bak"); err == nil {
			os.Rename("config.yaml.bak", "config.yaml")
		}
	}()

	// Test loading with no config file
	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test default values
	if config.Server.Port != 8080 {
		t.Errorf("Expected default port 8080, got %d", config.Server.Port)
	}
}
