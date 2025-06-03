#!/bin/bash

set -e

REPO="devetek/d-agent-healthchecks"
VERSION="v0.1.0"
ARCH="$(uname -m)"

echo "📦 Installing d-agent-healthchecks ($VERSION)..."

# Detect OS
if [ -f /etc/debian_version ]; then
  OS="debian"
elif [ -f /etc/redhat-release ]; then
  OS="redhat"
else
  echo "❌ Unsupported OS"
  exit 1
fi

# File path mapping
if [[ "$ARCH" == "x86_64" ]]; then
  if [[ "$OS" == "debian" ]]; then
    FILE="d-agent-healthchecks-0.1.0-1.x86_64.rpm"
    CMD="sudo dpkg -i $FILE"
  else
    FILE="d-agent-healthchecks-0.1.0-1.x86_64.rpm"
    CMD="sudo rpm -ivh $FILE"
  fi
else
  echo "❌ Unsupported architecture: $ARCH"
  exit 1
fi

# Install from local build folder
if [ ! -f "$FILE" ]; then
  echo "❌ File not found: $FILE"
  exit 1
fi

echo "⚙️ Installing $FILE..."
$CMD

echo "✅ Installed successfully!"

echo "🔧 To start the agent:"
echo "  sudo systemctl daemon-reload"
echo "  sudo systemctl enable d-agent-healthchecks"
echo "  sudo systemctl start d-agent-healthchecks"
