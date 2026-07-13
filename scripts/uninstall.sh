#!/usr/bin/env bash
# FileHub — Remove systemd service and optionally /opt/filehub
# Usage: sudo bash uninstall.sh

set -euo pipefail

APP_DIR="/opt/filehub"

echo "============================================"
echo "  FileHub Linux Uninstall"
echo "============================================"
echo ""

if systemctl is-active --quiet filehub 2>/dev/null; then
    echo "Stopping service..."
    systemctl stop filehub
fi

if systemctl is-enabled --quiet filehub 2>/dev/null; then
    echo "Disabling service..."
    systemctl disable filehub
fi

if [[ -f /etc/systemd/system/filehub.service ]]; then
    rm /etc/systemd/system/filehub.service
    systemctl daemon-reload
    echo "[OK] systemd unit removed"
else
    echo "[INFO] No systemd unit found"
fi

if [[ -d "$APP_DIR" ]]; then
    echo ""
    read -rp "Remove $APP_DIR? [y/N] " answer
    if [[ "$answer" =~ ^[Yy]$ ]]; then
        rm -rf "$APP_DIR"
        echo "[OK] $APP_DIR removed"
    else
        echo "[INFO] Kept $APP_DIR"
    fi
fi

echo ""
echo "[SUCCESS] FileHub uninstalled."
