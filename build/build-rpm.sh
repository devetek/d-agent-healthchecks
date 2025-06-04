#!/bin/bash

set -e

APP_NAME="d-agent-healthchecks"
VERSION="0.1.0"
ARCH="x86_64"
BUILD_ROOT="$(pwd)/rpmbuild"
SRC_DIR="$BUILD_ROOT/SOURCES"
SPEC_DIR="$BUILD_ROOT/SPECS"
TMP_BUILD_DIR="$BUILD_ROOT/tmp"
BIN_NAME="$APP_NAME"

echo "🛠️ Building $APP_NAME (RPM package)..."

# 🔍 Pastikan Go tersedia
if ! command -v go &> /dev/null; then
  echo "❌ ERROR: Go is not installed or not in PATH."
  exit 1
fi

# ✅ Pastikan configs/agent.yml ada
if [ ! -f "configs/agent.yml" ]; then
  echo "❌ ERROR: configs/agent.yml not found. Please create the config before building."
  exit 1
fi

# 📁 Siapkan direktori build
mkdir -p "$SRC_DIR" "$SPEC_DIR" "$TMP_BUILD_DIR"
TMP_SRC_DIR="$TMP_BUILD_DIR/$APP_NAME-$VERSION"
mkdir -p "$TMP_SRC_DIR"
rsync -a ./ "$TMP_SRC_DIR/" --exclude=rpmbuild --exclude=dist

# 🎁 Buat tarball source
tar -czf "$SRC_DIR/$APP_NAME-$VERSION.tar.gz" -C "$TMP_BUILD_DIR" "$APP_NAME-$VERSION"

# 📝 Generate file SPEC
cat > "$SPEC_DIR/$APP_NAME.spec" <<EOF
Name:           $APP_NAME
Version:        $VERSION
Release:        1%{?dist}
Summary:        Healthchecks.io agent for Linux

License:        MIT
Source0:        %{name}-%{version}.tar.gz

%global debug_package %{nil}

%description
Agent to run service checks and send pings to Healthchecks.io.

%prep
%setup -q

%build
go build -ldflags="-s -w" -o $APP_NAME ./cmd/agent

%install
mkdir -p %{buildroot}/usr/local/bin
mkdir -p %{buildroot}/etc/%{name}
mkdir -p %{buildroot}/etc/systemd/system

cp $APP_NAME %{buildroot}/usr/local/bin/
cp ./configs/agent.yml %{buildroot}/etc/%{name}/agent.yml
cp systemd/$APP_NAME.service %{buildroot}/etc/systemd/system/

%files
/usr/local/bin/$APP_NAME
/etc/%{name}/agent.yml
/etc/systemd/system/$APP_NAME.service

%changelog
* Tue Jun 03 2025 Devetek <dev@devetek.com> - $VERSION-1
- Initial RPM build
EOF

# 🧱 Jalankan rpmbuild
rpmbuild --define "_topdir $BUILD_ROOT" -ba "$SPEC_DIR/$APP_NAME.spec"

RPM_OUTPUT=$(find "$BUILD_ROOT/RPMS" -name "*.rpm")
echo "✅ RPM package built:"
echo "$RPM_OUTPUT"

# 🧱 Jalankan rpmbuild
rpmbuild --define "_topdir $BUILD_ROOT" -ba "$SPEC_DIR/$APP_NAME.spec"

# 🗂️ Temukan hasil RPM
RPM_OUTPUT=$(find "$BUILD_ROOT/RPMS" -name "*.rpm")

# 📦 Salin ke root direktori
cp "$RPM_OUTPUT" .

echo "✅ RPM package built and copied to root:"
echo "$(basename "$RPM_OUTPUT")"

# 🧹 Bersihkan build temporary (opsional tapi disarankan)
echo "🧹 Cleaning up temporary build directory..."
rm -rf "$TMP_BUILD_DIR"
