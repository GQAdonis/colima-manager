# Colima Manager

[Previous sections unchanged until Running section...]

## Running

The manager uses explicit commands for starting and stopping:

```bash
# Start the manager and all configured profiles
colima-manager start

# Stop all running instances and exit
colima-manager stop
```

### Startup Sequence with Auto Profile (-a flag)

When starting with the -a flag, the manager follows this sequence:

1. Configuration Loading:
   ```
   Starting Colima Manager
   Loading configuration...
   Configuration loaded successfully
   ```

2. Initialization:
   ```
   Initializing Colima repository...
   Colima repository initialized successfully
   Initializing Colima use case...
   Colima use case initialized successfully
   ```

3. Profile Setup (with -a flag):
   ```
   Auto flag detected, preparing to start default profile
   Loading configuration for profile: default
   Profile configuration: CPUs=4, Memory=8, DiskSize=60...
   Starting Colima profile 'default'...
   ```

4. Profile Status Monitoring:
   ```
   Waiting for profile 'default' to be fully ready...
   Profile 'default' status: Starting, waiting...
   Profile 'default' is now running with: CPUs=4, Memory=8...
   ```

5. Kubernetes Verification (if enabled):
   ```
   Verifying Kubernetes configuration...
   Kubernetes configuration verified successfully
   Profile 'default' is fully ready
   ```

6. HTTP Server Startup:
   ```
   Initializing HTTP server...
   Starting HTTP server at localhost:8080
   ```

The API server will only become available after the default profile is fully running and verified. This ensures that all services are ready before accepting requests.

Command-line flags can be combined:

```bash
# Start with automatic profile creation and specific config
colima-manager start -a -c /path/to/config.yaml

# Start with automatic profile in daemon mode
colima-manager start -a -d
```

[Rest of the README remains unchanged...]
