# Colima Manager

A Go-based management tool for Colima that handles container runtime and Kubernetes cluster operations. It provides automated startup, health monitoring, and Kubernetes configuration management.

## Features

- Automated Colima startup with configurable resources
- Docker socket monitoring
- Kubernetes cluster health checks
- Automatic kubeconfig merging
- Continuous health monitoring
- Graceful shutdown handling

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

The manager will:
1. Start Colima with predefined resources (12 CPUs, 32GB RAM, 100GB disk)
2. Wait for the Docker socket to become available
3. Configure Kubernetes if enabled
4. Begin continuous health monitoring

## Configuration

The application can be configured using a `config.yaml` file in the root directory. Here's a sample configuration:

```yaml
# Server configuration
server:
  # The port number the HTTP server will listen on
  # Default: 8080 if not specified
  port: 8080
```

## Testing Strategy

The project follows a comprehensive testing strategy:

1. **Unit Tests**: Each package contains unit tests that verify the behavior of individual components in isolation. Key areas covered include:
   - Domain logic in `internal/domain`
   - Use cases in `internal/usecase`
   - HTTP handlers in `internal/interface/http/handler`
   - Infrastructure components in `internal/infrastructure`

2. **Integration Tests**: Tests that verify the interaction between different components, particularly focusing on:
   - Colima operations
   - HTTP endpoint functionality
   - Configuration loading

3. **Test Coverage**: The project maintains test coverage through:
   - Regular test execution via `mage test`
   - Coverage reporting via `mage testcoverage`
   - Coverage reports are generated in HTML format at `coverage/coverage.html`

## Logs

Logs are stored in `~/colima-monitor/logs/`:
- `colima-monitor.log`: General operation logs
- `colima-monitor.err`: Error logs

## License

[Add your license information here]
