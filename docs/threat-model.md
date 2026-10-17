# Threat model

## Adversaries

- User-controlled browser or local process that tries to bypass the proxy
- Malicious or compromised local application
- Hostile local HTTP requests to the control plane
- DNS or address spoofing attempts
- Firewall policy tampering or route manipulation
- Service crashes, interrupted updates, and partial recovery states

## Primary threats

1. Traffic escape through direct interface access.
2. DNS exfiltration through ordinary resolver settings.
3. Partial activation with a stale or false health state.
4. UI or API compromise translating to privileged configuration changes.
5. Resource exhaustion causing packet loss or fallback under load.
6. Tampered configuration or operating-system state being trusted without reconciliation.

## Mitigations

- Fail-closed enforcement via platform firewall rules
- Independent privileged authority for network actions
- Strict loopback-only API binding with host/origin validation
- Bounded flow and packet-processing limits
- Reconciliation against actual OS state after restarts and reboots
- Conservative blocked state whenever validation fails

## Security posture

The project considers bypass, fallback, and misconfiguration to be the highest-risk outcomes. No component may report a protected state without verifying actual enforcement.
