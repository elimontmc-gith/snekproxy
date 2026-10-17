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

Implemented foundation:

- Strict YAML config loading
- Explicit `--enable` activation, root and nftables checks
- An owned `inet snekproxy` output policy that allows loopback and drops other locally generated traffic
- Exact-policy verification and refusal to overwrite or remove an unrecognized table

Not implemented: TUN, upstream proxy forwarding, DNS or route management, NetworkManager coordination, and automatic recovery.

Live-host status: `NOT VERIFIED`. Unit tests use a fake command runner and do not prove kernel enforcement.

## Windows support

The Windows model is designed separately with:

- Wintun or equivalent transport
- Windows service hardening
- WFP enforcement
- route and DNS handling

Status: `NOT VERIFIED` in this environment.

## Current repository state

The Qt desktop client builds per target operating system. Linux integration currently targets GNOME through `gsettings`; Windows integration targets current-user WinINET settings. Other Linux desktop environments are not supported by automatic proxy management. The client is not a kill switch and is not claimed to proxy all applications.

This repository does not claim production-grade verified enforcement from a real host.
