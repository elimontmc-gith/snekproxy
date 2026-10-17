# Architecture

## Overview

SnekProxy separates the user experience, privileged enforcement, and packet processing by clear trust boundaries. The design is intentionally fail-closed: the enforcement plane remains authoritative even if the UI or forwarding engine is unhealthy.

The current repository includes an OpenSSH host setup helper and a Qt desktop client that manages an SSH dynamic SOCKS5 tunnel, alongside an experimental Linux blocking-only enforcement foundation. The SSH tunnel is not integrated with the firewall policy. There is no TUN interface or packet-processing engine.

## Components

### Dashboard

The dashboard is unprivileged. It displays state, accepts user requests, and provides diagnostics. It never configures firewall rules or routes directly.

### Local HTTP API

The local API binds to loopback only, validates host and origin headers, and enforces authentication for all state-changing actions. The API is restricted to diagnostics and authorized control requests.

### Privileged networking service

The planned privileged service owns:

- TUN interface lifecycle
- route configuration
- DNS policy and policy rollback
- nftables or equivalent OS enforcement
- lifecycle state and recovery data

The implemented Linux foundation currently owns only an explicit, narrowly named nftables output table. It permits loopback and drops other locally generated traffic. It does not alter routes, DNS, interfaces, or other firewall tables.

### Packet-processing engine

The forwarding engine handles packet parsing, TCP/UDP flow handling, DNS translation, and backpressure management. It is isolated from privileged routing decisions and never serves as the sole authority for protection.

### Desktop proxy client

The C++/Qt client uses OpenSSH for authenticated dynamic SOCKS5 forwarding. After validating the local SOCKS listener, it updates and later restores current-user proxy settings: GNOME desktop settings on Linux and WinINET settings on Windows. It keeps a recovery record for restoration after a client crash. Applications that ignore these settings may bypass the tunnel; this component is not a system-wide kill switch.

## Trust boundaries

```text
User / Browser
  -> Dashboard
  -> Local loopback API
  -> Privileged service
  -> Enforcement policy (nftables / WFP)
  -> TUN / forwarding engine
  -> Physical network
```

## State model

The project defines the following protection states:

- DISABLED
- ACTIVATING
- PROTECTED
- DEGRADED_BLOCKED
- STOPPING
- RECOVERING
- ERROR_BLOCKED

The enforcement state is authoritative. A service may report healthy status, but the operating system enforcement policy remains the deciding factor.

## Failure model

Traffic is blocked when any of the following happens:

- forwarding engine fails
- required firewall rules disappear
- interfaces are removed or route state is inconsistent
- DNS policy cannot be enforced
- activation is partial or interrupted

This project prefers a safe blocked state over a degraded or insecure direct path.
