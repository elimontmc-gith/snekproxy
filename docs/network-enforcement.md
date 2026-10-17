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

## Implemented Linux foundation

The current foundation creates only `inet snekproxy` with an output base chain whose policy is `drop`, plus a loopback exception. It does not provide any forwarding or service egress exception. Therefore, explicit activation blocks all non-loopback locally generated traffic; it is not a usable proxy mode and must not be reported as `PROTECTED`. The service leaves the policy installed on process exit. `--disable` removes it only when its live rules exactly match the policy SnekProxy created; unknown or modified state is left untouched.

## Desktop SOCKS client

The separate desktop client uses SSH dynamic forwarding with a loopback-only SOCKS5 listener and manages the current user's system proxy settings on supported Linux and Windows desktops. It does not configure firewall rules or guarantee that all application traffic uses the proxy. Do not enable the unrelated blocking-only nftables foundation on the SSH host; it prevents SSH replies.
