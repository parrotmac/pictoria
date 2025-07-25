#!/bin/bash

# iptables Router Script
# Creates NAT routing between WAN and LAN interfaces
# Usage: ./iptables-router.sh [setup|teardown] [WAN_INTERFACE] [LAN_INTERFACE]

set -euo pipefail

ACTION="${1:-setup}"
WAN_INTERFACE="${2:-eth0}"
LAN_INTERFACE="${3:-eth1}"

# Custom chain names to identify our rules
CUSTOM_FORWARD_CHAIN="ROUTER_FORWARD"
CUSTOM_INPUT_CHAIN="ROUTER_INPUT"
CUSTOM_OUTPUT_CHAIN="ROUTER_OUTPUT"

log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" >&2
}

check_root() {
    if [[ $EUID -ne 0 ]]; then
        log "ERROR: This script must be run as root"
        exit 1
    fi
}

check_interfaces() {
    if ! ip link show "$WAN_INTERFACE" >/dev/null 2>&1; then
        log "ERROR: WAN interface $WAN_INTERFACE does not exist"
        exit 1
    fi
    
    if ! ip link show "$LAN_INTERFACE" >/dev/null 2>&1; then
        log "ERROR: LAN interface $LAN_INTERFACE does not exist"
        exit 1
    fi
}

enable_ip_forwarding() {
    log "Enabling IP forwarding"
    echo 1 > /proc/sys/net/ipv4/ip_forward
    
    # Make it persistent across reboots
    if ! grep -q "^net.ipv4.ip_forward=1" /etc/sysctl.conf 2>/dev/null; then
        echo "net.ipv4.ip_forward=1" >> /etc/sysctl.conf
    fi
}

disable_ip_forwarding() {
    log "Disabling IP forwarding"
    echo 0 > /proc/sys/net/ipv4/ip_forward
    
    # Remove from sysctl.conf
    sed -i '/^net.ipv4.ip_forward=1/d' /etc/sysctl.conf 2>/dev/null || true
}

create_custom_chains() {
    log "Creating custom iptables chains"
    
    # Create custom chains if they don't exist
    iptables-legacy -t filter -N "$CUSTOM_FORWARD_CHAIN" 2>/dev/null || true
    iptables-legacy -t filter -N "$CUSTOM_INPUT_CHAIN" 2>/dev/null || true
    iptables-legacy -t filter -N "$CUSTOM_OUTPUT_CHAIN" 2>/dev/null || true
}

remove_custom_chains() {
    log "Removing custom iptables chains"
    
    # Remove jumps to custom chains
    iptables-legacy -t filter -D FORWARD -j "$CUSTOM_FORWARD_CHAIN" 2>/dev/null || true
    iptables-legacy -t filter -D INPUT -j "$CUSTOM_INPUT_CHAIN" 2>/dev/null || true
    iptables-legacy -t filter -D OUTPUT -j "$CUSTOM_OUTPUT_CHAIN" 2>/dev/null || true
    
    # Flush and delete custom chains
    iptables-legacy -t filter -F "$CUSTOM_FORWARD_CHAIN" 2>/dev/null || true
    iptables-legacy -t filter -X "$CUSTOM_FORWARD_CHAIN" 2>/dev/null || true
    iptables-legacy -t filter -F "$CUSTOM_INPUT_CHAIN" 2>/dev/null || true
    iptables-legacy -t filter -X "$CUSTOM_INPUT_CHAIN" 2>/dev/null || true
    iptables-legacy -t filter -F "$CUSTOM_OUTPUT_CHAIN" 2>/dev/null || true
    iptables-legacy -t filter -X "$CUSTOM_OUTPUT_CHAIN" 2>/dev/null || true
}

setup_nat_rules() {
    log "Setting up NAT rules for WAN: $WAN_INTERFACE, LAN: $LAN_INTERFACE"
    
    # Clear existing NAT rules for our interfaces
    teardown_nat_rules
    
    # Enable masquerading for outgoing traffic on WAN interface
    iptables-legacy -t nat -A POSTROUTING -o "$WAN_INTERFACE" -j MASQUERADE
    
    # Allow forwarding from LAN to WAN
    iptables-legacy -t filter -A "$CUSTOM_FORWARD_CHAIN" -i "$LAN_INTERFACE" -o "$WAN_INTERFACE" -j ACCEPT
    
    # Allow forwarding from WAN to LAN for established/related connections
    iptables-legacy -t filter -A "$CUSTOM_FORWARD_CHAIN" -i "$WAN_INTERFACE" -o "$LAN_INTERFACE" -m state --state RELATED,ESTABLISHED -j ACCEPT
    
    # Allow loopback traffic
    iptables-legacy -t filter -A "$CUSTOM_INPUT_CHAIN" -i lo -j ACCEPT
    iptables-legacy -t filter -A "$CUSTOM_OUTPUT_CHAIN" -o lo -j ACCEPT
    
    # Allow traffic on LAN interface
    iptables-legacy -t filter -A "$CUSTOM_INPUT_CHAIN" -i "$LAN_INTERFACE" -j ACCEPT
    iptables-legacy -t filter -A "$CUSTOM_OUTPUT_CHAIN" -o "$LAN_INTERFACE" -j ACCEPT
    
    # Allow established/related traffic on WAN interface
    iptables-legacy -t filter -A "$CUSTOM_INPUT_CHAIN" -i "$WAN_INTERFACE" -m state --state RELATED,ESTABLISHED -j ACCEPT
    iptables-legacy -t filter -A "$CUSTOM_OUTPUT_CHAIN" -o "$WAN_INTERFACE" -j ACCEPT
    
    # Link custom chains to main chains
    iptables-legacy -t filter -I FORWARD 1 -j "$CUSTOM_FORWARD_CHAIN"
    iptables-legacy -t filter -I INPUT 1 -j "$CUSTOM_INPUT_CHAIN"
    iptables-legacy -t filter -I OUTPUT 1 -j "$CUSTOM_OUTPUT_CHAIN"
}

teardown_nat_rules() {
    log "Removing existing NAT and routing rules"
    
    # Remove NAT rules for our interfaces
    while iptables-legacy -t nat -D POSTROUTING -o "$WAN_INTERFACE" -j MASQUERADE 2>/dev/null; do
        true
    done
    
    # Clear custom chains
    iptables-legacy -t filter -F "$CUSTOM_FORWARD_CHAIN" 2>/dev/null || true
    iptables-legacy -t filter -F "$CUSTOM_INPUT_CHAIN" 2>/dev/null || true
    iptables-legacy -t filter -F "$CUSTOM_OUTPUT_CHAIN" 2>/dev/null || true
}

show_status() {
    log "Current iptables router status:"
    echo
    echo "=== IP Forwarding Status ==="
    echo "Current: $(cat /proc/sys/net/ipv4/ip_forward)"
    echo "Persistent: $(grep -c '^net.ipv4.ip_forward=1' /etc/sysctl.conf 2>/dev/null || echo 0)"
    echo
    echo "=== NAT Rules ==="
    iptables-legacy -t nat -L POSTROUTING -n -v | grep -E "(MASQUERADE|$WAN_INTERFACE)" || echo "No NAT rules found"
    echo
    echo "=== Custom Chains ==="
    for chain in "$CUSTOM_FORWARD_CHAIN" "$CUSTOM_INPUT_CHAIN" "$CUSTOM_OUTPUT_CHAIN"; do
        if iptables-legacy -t filter -L "$chain" >/dev/null 2>&1; then
            echo "$chain: EXISTS ($(iptables-legacy -t filter -L "$chain" --line-numbers | wc -l) rules)"
        else
            echo "$chain: NOT FOUND"
        fi
    done
    echo
    echo "=== Interface Status ==="
    echo "WAN ($WAN_INTERFACE): $(ip link show "$WAN_INTERFACE" 2>/dev/null | grep -o 'state [A-Z]*' || echo 'NOT FOUND')"
    echo "LAN ($LAN_INTERFACE): $(ip link show "$LAN_INTERFACE" 2>/dev/null | grep -o 'state [A-Z]*' || echo 'NOT FOUND')"
}

setup_router() {
    log "Setting up iptables router (WAN: $WAN_INTERFACE, LAN: $LAN_INTERFACE)"
    
    check_interfaces
    enable_ip_forwarding
    create_custom_chains
    setup_nat_rules
    
    log "Router setup completed successfully"
}

teardown_router() {
    log "Tearing down iptables router"
    
    teardown_nat_rules
    remove_custom_chains
    disable_ip_forwarding
    
    log "Router teardown completed successfully"
}

show_help() {
    cat << EOF
Usage: $0 [ACTION] [WAN_INTERFACE] [LAN_INTERFACE]

Actions:
    setup      - Configure iptables rules for routing (default)
    teardown   - Remove all routing rules and disable forwarding
    status     - Show current router configuration status
    help       - Show this help message

Arguments:
    WAN_INTERFACE  - External/internet interface (default: eth0)
    LAN_INTERFACE  - Internal/local network interface (default: eth1)

Examples:
    $0 setup eth0 wlan0       # Setup routing between eth0 (WAN) and wlan0 (LAN)
    $0 teardown               # Remove all routing rules
    $0 status                 # Show current status
    
Note: This script must be run as root and requires iptables to be installed.
EOF
}

main() {
    check_root
    
    case "$ACTION" in
        "setup")
            setup_router
            ;;
        "teardown")
            teardown_router
            ;;
        "status")
            show_status
            ;;
        "help"|"-h"|"--help")
            show_help
            ;;
        *)
            log "ERROR: Unknown action '$ACTION'"
            show_help
            exit 1
            ;;
    esac
}

# Run main function
main "$@"
