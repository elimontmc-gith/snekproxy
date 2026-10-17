# SnekProxy

SnekProxy is a fail-closed network protection project with Linux as the reference implementation target. This repository provides an architectural scaffold, policy model, and enforcement lifecycle for a system-wide proxy and kill switch.

## Current status

The repository now has separate `host/` and `client/` components. The host uses OpenSSH; the Qt 6 C++ desktop client creates an encrypted SOCKS5 tunnel and automatically configures/restores current-user proxy settings on supported Linux and Windows desktops. The Go Linux enforcement foundation remains blocking-only and is not integrated with SSH forwarding. This is not a system-wide kill switch: TUN, DNS enforcement, route management, and universal application proxy support are not implemented.

## Design intent

- Linux-first enforcement using TUN, nftables, and NetworkManager (planned; not implemented)
- Fail-closed policy: if proxying is unavailable, block traffic
- OpenSSH host and Qt desktop client for encrypted SOCKS5 forwarding
- Windows-specific enforcement model documented separately
- Tests and verification captured as `PASS`, `FAIL`, `NOT VERIFIED`, or `NOT APPLICABLE`

## Host and client

Set up the host's OpenSSH server:

```bash
su -c "bash ./host/setup.sh"
```

Build the Qt client from `client/`:

```bash
cd client
make
./build/snekproxy-client
```

The client asks for the host address, regular SSH username, SSH port, and local SOCKS5 port. Once connected, it updates the system proxy automatically; **Disable Tunnel** stops SSH and restores the prior settings. Linux currently supports GNOME desktops, and Windows uses current-user Internet proxy settings. See [host/README.md](host/README.md) and [client/README.md](client/README.md) for setup and limitations.

## Optional blocking-only foundation

The Go scaffold does not forward proxy traffic. Do not enable it on the SSH host: its nftables output drop rule would block SSH replies. If intentionally testing its blocking behavior on a separate Linux machine, activation requires root and will block non-loopback egress:

```bash
su -c "go run ./cmd/snekproxy --enable --config ./config.yaml"
```

Remove only SnekProxy's exact recognized policy with `--disable`.

## Security invariant

The OpenSSH SOCKS tunnel is encrypted and authenticated, but only configured applications use it. It does not guarantee protection from direct traffic or DNS leaks.
