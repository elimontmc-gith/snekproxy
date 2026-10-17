# Network enforcement

## Coverage matrix

| Category | Handling | Policy |
|---|---|---|
| IPv4 | Proxied or blocked | Must be constrained to protected path |
| IPv6 | Proxied or blocked | Must be constrained to protected path |
| TCP | Proxied | Bounded flow and timeout enforcement |
| UDP | Proxied or blocked | Bounded flow, expiry, and validation |
| ICMP / ICMPv6 | Blocked unless explicitly supported | Default is deny |
| DNS (UDP/TCP) | Proxied or blocked | Must not bypass resolver policy |
| DHCP / DHCPv6 | Local exception if required | Narrowly scoped and documented |
| Link-local / multicast / broadcast | Blocked unless required | Default deny |
| Loopback control traffic | Local exception | Only explicitly authorized |
| Unsupported IP protocols | Blocked | Never silently bypass |
| Service-generated egress | Explicit exception | Limited and traceable |

## Activation lifecycle

1. Validate environment and configuration.
2. Install blocking policy.
3. Initialize TUN and forwarding support.
4. Apply route and DNS policy.
5. Validate actual enforcement.
6. Transition to PROTECTED only after validations pass.

## Protection requirements

- No direct fallback on failure
- No broad port, UID, or destination exemptions
- No silent removal of firewall state
- Recovery must reconcile against actual OS state
- Any physical egress must be attributable to authorized SnekProxy operations

## Egress handling

Service-generated egress is allowed only for explicitly authorized SnekProxy control or forwarding traffic. This project expects those paths to be minimal and separately defined by the privileged service. The general default remains: no direct traffic outside the protected path while protected.
