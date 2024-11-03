#!/bin/bash

# Exit on any error
set -e

# Get the current user
CURRENT_USER=$(whoami)
HOME_DIR=$(eval echo ~$CURRENT_USER)

echo "Installing Colima Manager service for user: $CURRENT_USER"

# Create log directory
sudo mkdir -p /var/log/colima-manager
sudo chown $CURRENT_USER:staff /var/log/colima-manager

# Get the directory where the script is located
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# Create temporary plist file with replaced username
TEMP_PLIST=$(mktemp)
cat "$SCRIPT_DIR/com.tribemedia.colima-manager.plist" | \
    sed "s/gqadonis/$CURRENT_USER/g" | \
    sed "s|/Users/gqadonis|$HOME_DIR|g" > "$TEMP_PLIST"

# Create LaunchAgents directory if it doesn't exist
mkdir -p "$HOME_DIR/Library/LaunchAgents"

# Copy the modified plist file
cp "$TEMP_PLIST" "$HOME_DIR/Library/LaunchAgents/com.tribemedia.colima-manager.plist"
rm "$TEMP_PLIST"

# Set correct permissions
chmod 644 "$HOME_DIR/Library/LaunchAgents/com.tribemedia.colima-manager.plist"

# Unload the service if it's already loaded
launchctl unload "$HOME_DIR/Library/LaunchAgents/com.tribemedia.colima-manager.plist" 2>/dev/null || true

# Load the service
launchctl load "$HOME_DIR/Library/LaunchAgents/com.tribemedia.colima-manager.plist"

# Start the service
launchctl start com.tribemedia.colima-manager

echo "Colima Manager service has been installed and started."
echo "The service will automatically start on system boot."
echo "Logs can be found in /var/log/colima-manager/"
