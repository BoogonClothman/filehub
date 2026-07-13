#!/usr/bin/env bash
# FileHub — Install to /opt/filehub and register systemd service
# Usage: sudo bash install.sh

set -euo pipefail

APP_DIR="/opt/filehub"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

echo "============================================"
echo "  FileHub Linux Installation"
echo "============================================"
echo ""

# Stop & disable if already installed
if systemctl is-active --quiet filehub 2>/dev/null; then
    echo "[INFO] Stopping existing service..."
    systemctl stop filehub
fi

# Create app directory
mkdir -p "$APP_DIR"
mkdir -p "$APP_DIR/data"

# Copy binary
if [[ -f "$PROJECT_DIR/filehub" ]]; then
    cp "$PROJECT_DIR/filehub" "$APP_DIR/filehub"
    chmod +x "$APP_DIR/filehub"
    echo "[OK] Binary installed to $APP_DIR/filehub"
else
    echo "[ERROR] filehub binary not found in project root."
    echo "        Run: go build -o filehub ."
    exit 1
fi

# Copy config if not exists
if [[ ! -f "$APP_DIR/config.json" ]]; then
    if [[ -f "$PROJECT_DIR/config.json" ]]; then
        cp "$PROJECT_DIR/config.json" "$APP_DIR/config.json"
    fi
fi

# Install systemd unit
cp "$SCRIPT_DIR/filehub.service" /etc/systemd/system/filehub.service
systemctl daemon-reload
echo "[OK] systemd unit installed"

# Enable and start
systemctl enable filehub
systemctl start filehub
echo "[OK] Service started"

# Health check
sleep 1
if systemctl is-active --quiet filehub; then
    echo ""
    echo "[SUCCESS] FileHub is running!"
    echo "  http://$(hostname -I | awk '{print $1}'):5000"
else
    echo "[WARNING] Service may not have started. Check: journalctl -u filehub -n 20"
fi

echo ""
echo "Manage: systemctl {start,stop,restart,status} filehub"
echo "Logs:   journalctl -u filehub -f"
