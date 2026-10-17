# Testing and verification

## Verification status model

Every requirement must be classified as one of:

- PASS
- FAIL
- NOT VERIFIED
- NOT APPLICABLE

The project does not treat skipped tests as passing.

## Current repository status

This scaffold is intentionally conservative and incomplete with respect to production verification. The following statuses are recorded for the repository as it exists now.

### Linux implementation

- TUN lifecycle design: `NOT VERIFIED`
- nftables enforcement: `NOT VERIFIED`
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
| TUN | Interface creation and packet receive | Safe and bounded | NOT VERIFIED |
| TCP | Bidirectional forwarding | Functional and bounded | NOT VERIFIED |
| UDP | Datagram forwarding | Bounded and safe | NOT VERIFIED |
| DNS | DNS leak prevention | No direct resolver fallback | NOT VERIFIED |
| Recovery | Crash and restart handling | Reconcile actual OS state | NOT VERIFIED |
| Security | Authorization and host validation | Must reject hostile requests | NOT VERIFIED |
| Platform | Windows verification | Must not claim PASS without host evidence | NOT VERIFIED |

## Build status

This repository is in early scaffold state and does not yet have a complete multi-platform build or test harness verified on real hosts.
