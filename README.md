# SnekProxy

SnekProxy is a fail-closed network protection project with Linux as the reference implementation target. This repository provides an architectural scaffold, policy model, and enforcement lifecycle for a system-wide proxy and kill switch.

## Current status

The repository is intentionally scaffolded and security-first. It does not claim verified full-device enforcement on a live Linux host or Windows host in this environment. The code reflects the required architecture, strict state transitions, and a conservative default that blocks traffic when the forwarding path is not ready.

## Design intent

- Linux-first enforcement using TUN, nftables, and NetworkManager
- Fail-closed policy: if proxying is unavailable, block traffic
- Independent privileged enforcement separate from the dashboard
- Windows-specific enforcement model documented separately
- Tests and verification captured as `PASS`, `FAIL`, `NOT VERIFIED`, or `NOT APPLICABLE`

## Run

```bash
go run ./cmd/snekproxy --enable --config ./config.yaml
```

## Security invariant

When protection is enabled, the service must never permit ordinary direct connectivity while the forwarding engine or enforcement policy is unavailable.
