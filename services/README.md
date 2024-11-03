# Colima Manager Service Installation

This directory contains service files for running Colima Manager as a system service on MacOS and Linux.

## MacOS Installation

1. Copy the binary to the system bin directory:
```bash
sudo cp ./colima-manager /usr/local/bin/
sudo chmod +x /usr/local/bin/colima-manager
```

2. Create log directory:
```bash
sudo mkdir -p /var/log/colima-manager
```

3. Copy the service file:
```bash
sudo cp ./services/macos/com.tribemedia.colima-manager.plist /Library/LaunchDaemons/
```

4. Load and start the service:
```bash
sudo launchctl load /Library/LaunchDaemons/com.tribemedia.colima-manager.plist
```

## Linux Installation

1. Copy the binary to the system bin directory:
```bash
sudo cp ./colima-manager /usr/local/bin/
sudo chmod +x /usr/local/bin/colima-manager
```

2. Create log directory:
```bash
sudo mkdir -p /var/log/colima-manager
```

3. Copy the service file:
```bash
sudo cp ./services/linux/colima-manager.service /etc/systemd/system/
```

4. Reload systemd and enable the service:
```bash
sudo systemctl daemon-reload
sudo systemctl enable colima-manager
sudo systemctl start colima-manager
```

## Checking Service Status

### MacOS
```bash
sudo launchctl list | grep colima-manager
```

### Linux
```bash
sudo systemctl status colima-manager
```

## Viewing Logs

Logs are stored in `/var/log/colima-manager/` for both platforms:
- output.log: Contains standard output
- error.log: Contains error messages
