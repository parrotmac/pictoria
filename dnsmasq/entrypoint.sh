#!/usr/bin/env bash
set -euo pipefail

DNSMASQ_CONTAINER_ENABLED="${DNSMASQ_CONTAINER_ENABLED:-true}"

if [ "$DNSMASQ_CONTAINER_ENABLED" != "true" ]; then
    echo "DNSMASQ_CONTAINER_ENABLED is not set to 'true'. Sleeping forever."
    while true; do sleep 1000; done
fi

BRIDGE_WAN_INTERFACE="${BRIDGE_WAN_INTERFACE:-eth0}"
BRIDGE_LAN_INTERFACE="${BRIDGE_LAN_INTERFACE:-eth1}"

echo "Bridging ${BRIDGE_WAN_INTERFACE} [WAN] to ${BRIDGE_LAN_INTERFACE} [LAN]"

/bridge.sh setup "${BRIDGE_WAN_INTERFACE}" "${BRIDGE_LAN_INTERFACE}"
/bridge.sh status

if [ -z "$DNSMASQ_BIND_INTERFACE" ]; then
    echo "ERROR: DNSMASQ_BIND_INTERFACE environment variable is not defined"
    echo "Please set DNSMASQ_BIND_INTERFACE to the network interface name (e.g., eth0)"
    exit 1
fi

echo "Starting dnsmasq DHCP server..."
echo "Binding to interface: $DNSMASQ_BIND_INTERFACE"

sed -i "s/^interface=.*$/interface=${DNSMASQ_BIND_INTERFACE}/" /etc/dnsmasq.conf

if ! ip link show "$DNSMASQ_BIND_INTERFACE" >/dev/null 2>&1; then
    echo "WARNING: Interface $DNSMASQ_BIND_INTERFACE does not exist or is not available"
    echo "Available interfaces:"
    ip link show | grep '^[0-9]' | awk '{print $2}' | sed 's/://'
fi

ip addr del 10.25.7.25 dev "${DNSMASQ_BIND_INTERFACE}" || true
ip addr add 10.25.7.25/24 dev "${DNSMASQ_BIND_INTERFACE}"

exec dnsmasq --no-daemon --log-queries
