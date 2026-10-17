# Security model

## Primary invariant

SnekProxy protection is a kill-switch policy, not a best-effort service. When protection is enabled, ordinary network traffic must not be allowed to bypass the proxy path.

If the forwarding path is unavailable, the enforcement policy blocks traffic. Direct connectivity is never treated as a safe fallback.

## Security assumptions

- The privileged service is the root of trust for enforcement state.
- The dashboard is not trusted to enforce policy.
- All packets entering the TUN or physical interfaces are treated as untrusted input.
- The operating system firewall must provide the actual protection guarantee.
- A health endpoint alone is not evidence of protection.

## Protection contract

The system must satisfy all of the following:

1. No traffic bypasses the proxy path while protected.
2. IPv4 and IPv6 are both covered.
3. DNS is protected and must not escape through direct resolvers.
4. Unsupported traffic is blocked.
5. A service crash or dashboard crash cannot disable the enforcement policy.
6. User-requested stop actions restore ordinary networking only through a controlled cleanup path.

## Default fail-closed rule

```text
if protection is enabled and forwarding is not validated:
    block all non-authorized traffic
```

## Out-of-scope claims

This scaffold does not claim live kernel-level enforcement has been proven against a real Linux host or a real Windows host in this environment. Those are tracked as `NOT VERIFIED` in the testing documentation.
