# Platform support

## Reference platform

Linux is the reference platform for this project.

## Linux support

The Linux implementation target includes:

- TUN interface creation and lifecycle
- nftables enforcement
- NetworkManager coordination
- route and resolver policy management
- service hardening and recovery

Status: `NOT VERIFIED` in this environment because the sandbox is not a live Linux host with actual firewall and network privileges.

## Windows support

The Windows model is designed separately with:

- Wintun or equivalent transport
- Windows service hardening
- WFP enforcement
- route and DNS handling

Status: `NOT VERIFIED` in this environment.

## Current repository state

This scaffold documents the architecture and programmatic state model but does not claim production-grade verified enforcement from a real host. The project is intentionally conservative and safe-by-default.
