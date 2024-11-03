# Colima Manager

A Go-based management tool for Colima that handles container runtime and Kubernetes cluster operations. It provides automated startup, health monitoring, and Kubernetes configuration management.

## Features

- Automated Colima startup with configurable resources
- Docker socket monitoring
- Kubernetes cluster health checks
- Automatic kubeconfig merging
- Continuous health monitoring
- Graceful shutdown handling
- Multiple profile support with auto-start capability
- System service integration for automatic startup

## Prerequisites

- Go 1.22.0 or later
- Colima installed
- kubectl installed (for Kubernetes operations)
- Mage installed (`go install github.com/magefile/mage@latest`)

## Building and Installing

This project uses Mage for build automation. Here are the available commands:

```bash
# Install dependencies
mage installdeps

# Build the binary locally
mage build

# Build and install to /usr/local/bin
mage install

# Run tests
mage test

# Run tests with coverage report
mage testcoverage

# Clean build artifacts
mage clean
```

When using `mage install`, the binary will be built and installed to `/usr/local/bin/colima-manager`, making it globally accessible from your terminal.

## Running

After installing, you can run the manager from anywhere:

```bash
colima-manager
```

To automatically start a profile on launch:
```bash
colima-manager -a
```

The manager will:
1. Start Colima with the configured profile settings (or default if not specified)
2. Wait for the Docker socket to become available
3. Configure Kubernetes if enabled
4. Begin continuous health monitoring

## System Service Installation

### macOS

To install Colima Manager as a system service that starts automatically on boot:

```bash
# Navigate to the services/macos directory
cd services/macos

# Run the installation script
./install.sh
```

The installation script will:
1. Create necessary log directories
2. Install the service for the current user
3. Configure the service to start on system boot
4. Start the service immediately

Logs will be available at `/var/log/colima-manager/`:
- `output.log`: Standard output logs
- `error.log`: Error logs

To manage the service:
```bash
# Stop the service
launchctl stop com.tribemedia.colima-manager

# Start the service
launchctl start com.tribemedia.colima-manager

# Unload the service (prevent it from starting on boot)
launchctl unload ~/Library/LaunchAgents/com.tribemedia.colima-manager.plist

# Load the service (enable it to start on boot)
launchctl load ~/Library/LaunchAgents/com.tribemedia.colima-manager.plist
```

## Configuration

The application can be configured using a `config.yaml` file in the root directory. Here's a sample configuration:

```yaml
# Server configuration
server:
  # The port number the HTTP server will listen on
  # Default: 8080 if not specified
  port: 8080
  
  # Auto-start configuration
  auto:
    # Whether to automatically start a profile on server startup
    enabled: true
    # The profile to start automatically (must exist in profiles section)
    default: "default"

# Colima profiles configuration
profiles:
  # Default profile with recommended settings
  default:
    cpus: 12
    memory: 32
    disk_size: 100
    vm_type: "vz"
    runtime: "containerd"
    network_address: true
    kubernetes: true
```

### Configuration Options

#### Server Section
- `port`: The HTTP server port (default: 8080)
- `auto.enabled`: Enable automatic profile startup
- `auto.default`: The profile name to start automatically

#### Profiles Section
Each profile can have the following settings:
- `cpus`: Number of CPUs to allocate
- `memory`: Amount of memory in GB
- `disk_size`: Disk size in GB
- `vm_type`: VM type (e.g., "vz")
- `runtime`: Container runtime (e.g., "containerd")
- `network_address`: Enable network address
- `kubernetes`: Enable Kubernetes support

## Testing Strategy

The project follows a comprehensive testing strategy:

1. **Unit Tests**: Each package contains unit tests that verify the behavior of individual components in isolation. Key areas covered include:
   - Domain logic in `internal/domain`
   - Use cases in `internal/usecase`
   - HTTP handlers in `internal/interface/http/handler`
   - Infrastructure components in `internal/infrastructure`
   - Configuration loading and validation in `internal/config`

2. **Integration Tests**: Tests that verify the interaction between different components, particularly focusing on:
   - Colima operations
   - HTTP endpoint functionality
   - Configuration loading
   - Profile management

3. **Test Coverage**: The project maintains test coverage through:
   - Regular test execution via `mage test`
   - Coverage reporting via `mage testcoverage`
   - Coverage reports are generated in HTML format at `coverage/coverage.html`

## Logs

Logs are stored in:
- Runtime logs: `~/colima-monitor/logs/`
  - `colima-monitor.log`: General operation logs
  - `colima-monitor.err`: Error logs
- Service logs (when installed as system service): `/var/log/colima-manager/`
  - `output.log`: Standard output logs
  - `error.log`: Error logs

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
