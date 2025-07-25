#!/usr/bin/env bash
set -euo pipefail

MOUNT_EXTERNAL_STORAGE_DEVICE="${MOUNT_EXTERNAL_STORAGE_DEVICE:-}"
MOUNT_EXTERNAL_STORAGE_TARGET="${MOUNT_EXTERNAL_STORAGE_TARGET:-/mnt/external}"
MOUNT_EXTERNAL_STORAGE_FSTYPE="${MOUNT_EXTERNAL_STORAGE_FSTYPE:-ext4}"

if [ "$MOUNT_EXTERNAL_STORAGE_DEVICE" ]; then
    if [ ! -d "$MOUNT_EXTERNAL_STORAGE_TARGET" ]; then
        echo "Creating mount point at $MOUNT_EXTERNAL_STORAGE_TARGET"
        mkdir -p "$MOUNT_EXTERNAL_STORAGE_TARGET"
    fi

    echo "Mounting external storage device $MOUNT_EXTERNAL_STORAGE_DEVICE to $MOUNT_EXTERNAL_STORAGE_TARGET"
    
    if ! mount | grep -q "$MOUNT_EXTERNAL_STORAGE_TARGET"; then
        mount -t "$MOUNT_EXTERNAL_STORAGE_FSTYPE" "$MOUNT_EXTERNAL_STORAGE_DEVICE" "$MOUNT_EXTERNAL_STORAGE_TARGET"
        echo "Mounted successfully."
    else
        echo "Device is already mounted."
    fi
else
    echo "No external storage device specified. Skipping mount."
fi

echo "Starting Pictoria application..."
exec ./pictoria $@
