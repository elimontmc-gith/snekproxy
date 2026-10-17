# SnekProxy host

The host side uses OpenSSH for encrypted and authenticated SOCKS5 forwarding. No custom proxy daemon or unauthenticated public proxy port is exposed. The SSH server listens on the host; each client creates its own loopback-only SOCKS5 endpoint through an SSH dynamic forward.

## Install on Debian or Ubuntu

From the repository root, install and start OpenSSH with `su` (the script never reads or stores the root password):

```sh
su -c "bash ./host/setup.sh"
```

The helper installs `openssh-server` and enables its `ssh` service. It does not change firewall rules or SSH authentication policy.

## Configure access

1. Create or choose a regular host user. Do not use root for proxy access.
2. Find the host's LAN address with `ip -brief address`.
3. Add the client's public key to that user's `~/.ssh/authorized_keys` (recommended). Keep the private key on the client.
4. Allow the SSH TCP port (normally 22) through any host or network firewall.
5. From another machine, confirm ordinary SSH access first: `ssh -p 22 USER@HOST_IP`.

The C++ desktop client uses SSH's existing host-key verification. Before connecting in the app, verify and record the host key from a terminal with the host administrator. The client does not disable strict checking or accept an unknown key for you.

## Important

The existing SnekProxy nftables foundation is blocking-only, not a proxy. Do not activate it on this host: its output drop policy prevents the SSH server from replying to clients. The OpenSSH tunnel is the proxy path in this milestone; TUN, DNS enforcement, and kill-switch integration are not implemented.
