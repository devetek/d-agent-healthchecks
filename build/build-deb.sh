#!/bin/bash
set -e

APP_NAME="d-agent-healthchecks"
VERSION="0.1.0"
ARCH="amd64"
BUILD_DIR="dist"

echo "🛠️ Building $APP_NAME (DEB package)..."

# Cleanup sebelumnya
rm -rf "$BUILD_DIR"
mkdir -p "$BUILD_DIR/DEBIAN"
mkdir -p "$BUILD_DIR/usr/local/bin"
mkdir -p "$BUILD_DIR/etc/$APP_NAME"
mkdir -p "$BUILD_DIR/etc/systemd/system"

# 🧱 Build binary dari cmd/agent
go build -o "$BUILD_DIR/usr/local/bin/$APP_NAME" ./cmd/agent

# 📄 Salin file konfigurasi & service
cp configs/agent.yml "$BUILD_DIR/etc/$APP_NAME/agent.yml"
cp systemd/$APP_NAME.service "$BUILD_DIR/etc/systemd/system/$APP_NAME.service"

# 📦 File kontrol DEB
cat > "$BUILD_DIR/DEBIAN/control" <<EOF
Package: $APP_NAME
Version: $VERSION
Section: base
Priority: optional
Architecture: $ARCH
Maintainer: Devetek
Description: Healthchecks.io Agent for system monitoring
EOF

# 📦 Build file .deb
dpkg-deb --build "$BUILD_DIR" "${APP_NAME}_${VERSION}_${ARCH}.deb"

echo "✅ DEB package built: ${APP_NAME}_${VERSION}_${ARCH}.deb"
