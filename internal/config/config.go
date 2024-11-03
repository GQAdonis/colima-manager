package config

import (
	"flag"
	"os"

	"gopkg.in/yaml.v2"
)

type ProfileConfig struct {
	CPUs           int    `yaml:"cpus"`
	Memory         int    `yaml:"memory"`
	DiskSize       int    `yaml:"disk_size"`
	VMType         string `yaml:"vm_type"`
	Runtime        string `yaml:"runtime"`
	NetworkAddress bool   `yaml:"network_address"`
	Kubernetes     bool   `yaml:"kubernetes"`
}

type AutoConfig struct {
	Enabled bool   `yaml:"enabled"`
	Default string `yaml:"default"`
}

type Config struct {
	Server struct {
		Port   int        `yaml:"port"`
		Host   string     `yaml:"host"`
		Daemon bool       `yaml:"daemon"`
		Auto   AutoConfig `yaml:"auto"`
	} `yaml:"server"`
	Profiles map[string]ProfileConfig `yaml:"profiles"`
}

func LoadConfig() (*Config, error) {
	config := &Config{}
	config.Server.Port = 8080        // Default port
	config.Server.Host = "localhost" // Default host

	// Define command line flags
	daemon := flag.Bool("d", false, "Run in daemon mode")
	daemonLong := flag.Bool("daemon", false, "Run in daemon mode")
	host := flag.String("h", "", "Server host address")
	hostLong := flag.String("host", "", "Server host address")

	flag.Parse()

	// Try to load from config file
	data, err := os.ReadFile("config.yaml")
	if err == nil {
		if err := yaml.Unmarshal(data, config); err != nil {
			return nil, err
		}
	}

	// Override with command line flags if provided
	if *daemon || *daemonLong {
		config.Server.Daemon = true
	}

	if *host != "" || *hostLong != "" {
		hostVal := *host
		if hostVal == "" {
			hostVal = *hostLong
		}
		config.Server.Host = hostVal
	}

	return config, nil
}
