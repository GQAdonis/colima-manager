package config

import (
	"flag"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	// Parse flags before running tests to handle test flags
	flag.Parse()
	os.Exit(m.Run())
}

func createTestConfig(t *testing.T) (string, func()) {
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

	if _, err := tmpfile.Write(content); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	cleanup := func() {
		os.Remove(tmpfile.Name())
	}

	return tmpfile.Name(), cleanup
}

func verifyConfig(t *testing.T, config *Config) {
	if config.Server.Port != 9090 {
		t.Errorf("Expected port 9090, got %d", config.Server.Port)
	}

	if !config.Server.Auto.Enabled {
		t.Error("Expected auto.enabled to be true")
	}
	if config.Server.Auto.Default != "test-profile" {
		t.Errorf("Expected default profile 'test-profile', got '%s'", config.Server.Auto.Default)
	}

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

func TestFlagPatterns(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		envVar   string
		checkFn  func(*Config) bool
		expected bool
	}{
		{
			name: "Short config flag",
			args: []string{"-c", "custom.yaml"},
			checkFn: func(c *Config) bool {
				return flag.Lookup("c") != nil && flag.Lookup("config") != nil
			},
			expected: true,
		},
		{
			name: "Long config flag",
			args: []string{"--config", "custom.yaml"},
			checkFn: func(c *Config) bool {
				return flag.Lookup("c") != nil && flag.Lookup("config") != nil
			},
			expected: true,
		},
		{
			name: "Short daemon flag",
			args: []string{"-d"},
			checkFn: func(c *Config) bool {
				return flag.Lookup("d") != nil && flag.Lookup("daemon") != nil && c.Server.Daemon
			},
			expected: true,
		},
		{
			name: "Long daemon flag",
			args: []string{"--daemon"},
			checkFn: func(c *Config) bool {
				return flag.Lookup("d") != nil && flag.Lookup("daemon") != nil && c.Server.Daemon
			},
			expected: true,
		},
		{
			name: "Short host flag",
			args: []string{"-h", "localhost"},
			checkFn: func(c *Config) bool {
				return flag.Lookup("h") != nil && flag.Lookup("host") != nil && c.Server.Host == "localhost"
			},
			expected: true,
		},
		{
			name: "Long host flag",
			args: []string{"--host", "localhost"},
			checkFn: func(c *Config) bool {
				return flag.Lookup("h") != nil && flag.Lookup("host") != nil && c.Server.Host == "localhost"
			},
			expected: true,
		},
		{
			name: "Short auto flag",
			args: []string{"-a"},
			checkFn: func(c *Config) bool {
				return flag.Lookup("a") != nil && flag.Lookup("auto") != nil &&
					c.Server.Auto.Enabled &&
					c.Server.Auto.Default == "default" &&
					c.Profiles["default"].CPUs == 4
			},
			expected: true,
		},
		{
			name: "Long auto flag",
			args: []string{"--auto"},
			checkFn: func(c *Config) bool {
				return flag.Lookup("a") != nil && flag.Lookup("auto") != nil &&
					c.Server.Auto.Enabled &&
					c.Server.Auto.Default == "default" &&
					c.Profiles["default"].CPUs == 4
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original state
			oldArgs := os.Args
			oldFlagCommandLine := flag.CommandLine
			defer func() {
				os.Args = oldArgs
				flag.CommandLine = oldFlagCommandLine
			}()

			// Reset flags for this test
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
			os.Args = append([]string{"cmd"}, tt.args...)

			if tt.envVar != "" {
				oldEnv, exists := os.LookupEnv("COLIMA_MANAGER_CONFIG")
				os.Setenv("COLIMA_MANAGER_CONFIG", tt.envVar)
				defer func() {
					if exists {
						os.Setenv("COLIMA_MANAGER_CONFIG", oldEnv)
					} else {
						os.Unsetenv("COLIMA_MANAGER_CONFIG")
					}
				}()
			}

			config, err := LoadConfig()
			if err != nil {
				t.Fatalf("Failed to load config: %v", err)
			}

			if got := tt.checkFn(config); got != tt.expected {
				t.Errorf("Test %s failed: expected %v, got %v", tt.name, tt.expected, got)
			}
		})
	}
}

func TestAutoFlagWithExistingProfile(t *testing.T) {
	// Save original state
	oldArgs := os.Args
	oldFlagCommandLine := flag.CommandLine
	defer func() {
		os.Args = oldArgs
		flag.CommandLine = oldFlagCommandLine
	}()

	// Create a config file with an existing default profile
	content := []byte(`
server:
  auto:
    default: "default"
profiles:
  default:
    cpus: 8
    memory: 16
    disk_size: 100
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

	// Reset flags for this test
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	os.Args = []string{"cmd", "-a", "-c", tmpfile.Name()}

	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify that the existing profile wasn't overwritten
	if config.Profiles["default"].CPUs != 8 {
		t.Errorf("Expected CPUs to be 8, got %d", config.Profiles["default"].CPUs)
	}
	if config.Profiles["default"].Memory != 16 {
		t.Errorf("Expected Memory to be 16, got %d", config.Profiles["default"].Memory)
	}
}

func TestLoadConfigWithEnvVar(t *testing.T) {
	configPath, cleanup := createTestConfig(t)
	defer cleanup()

	// Save original state
	oldArgs := os.Args
	oldFlagCommandLine := flag.CommandLine
	oldEnv, envExists := os.LookupEnv("COLIMA_MANAGER_CONFIG")
	defer func() {
		os.Args = oldArgs
		flag.CommandLine = oldFlagCommandLine
		if envExists {
			os.Setenv("COLIMA_MANAGER_CONFIG", oldEnv)
		} else {
			os.Unsetenv("COLIMA_MANAGER_CONFIG")
		}
	}()

	// Reset flags and set environment variable
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	os.Args = []string{"cmd"}
	os.Setenv("COLIMA_MANAGER_CONFIG", configPath)

	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	verifyConfig(t, config)
}

func TestLoadConfigDefaults(t *testing.T) {
	// Save original state
	oldArgs := os.Args
	oldFlagCommandLine := flag.CommandLine
	defer func() {
		os.Args = oldArgs
		flag.CommandLine = oldFlagCommandLine
	}()

	// Reset flags for this test
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	os.Args = []string{"cmd"}

	// Temporarily move any existing config file
	if err := os.Rename("config.yaml", "config.yaml.bak"); err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	defer func() {
		if _, err := os.Stat("config.yaml.bak"); err == nil {
			os.Rename("config.yaml.bak", "config.yaml")
		}
	}()

	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if config.Server.Port != 8080 {
		t.Errorf("Expected default port 8080, got %d", config.Server.Port)
	}
}
