# Testing and verification

## Verification status model

Every requirement must be classified as one of:

- PASS
- FAIL
- NOT VERIFIED
- NOT APPLICABLE

The project does not treat skipped tests as passing.

## Current repository status

The implementation is intentionally incomplete with respect to production verification. Unit tests cover config validation, lifecycle ordering, fail-closed state reporting, and nftables command behavior through a fake runner. They do not change host networking or prove kernel enforcement.

### Linux implementation

- Config loading and lifecycle unit tests: `PASS`
- nftables live enforcement on a host: `NOT VERIFIED`
- TUN lifecycle design: `NOT VERIFIED`
- NetworkManager integration: `NOT VERIFIED`
- DNS policy enforcement: `NOT VERIFIED`
- Failure and recovery testing: `NOT VERIFIED`

### Windows implementation

- WFP enforcement design: `NOT VERIFIED`
- Wintun integration: `NOT VERIFIED`
- Windows failure injection: `NOT VERIFIED`

## Mandatory test matrix

| Area | Test | Expected result | Status |
|---|---|---|---|
| Activation | Service starts with blocked default | No direct bypass | NOT VERIFIED |
| Linux foundation | Exact owned nftables policy only | Unknown policy is never overwritten or deleted | PASS (unit tests); NOT VERIFIED (live host) |
| TUN | Interface creation and packet receive | Safe and bounded | NOT VERIFIED |
| TCP | Bidirectional forwarding | Functional and bounded | NOT VERIFIED |
| UDP | Datagram forwarding | Bounded and safe | NOT VERIFIED |
| DNS | DNS leak prevention | No direct resolver fallback | NOT VERIFIED |
| Recovery | Crash and restart handling | Reconcile actual OS state | NOT VERIFIED |
| Security | Authorization and host validation | Must reject hostile requests | NOT VERIFIED |
| Platform | Windows verification | Must not claim PASS without host evidence | NOT VERIFIED |

## Build status

This repository is in early scaffold state and does not yet have a complete multi-platform build or test harness verified on real hosts.
