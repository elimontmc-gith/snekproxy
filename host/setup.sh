#!/usr/bin/env bash
set -euo pipefail

if [ "$(id -u)" -ne 0 ]; then
  echo "Run this setup script as root (for example: su -c 'bash ./host/setup.sh')." >&2
  exit 1
fi

if ! command -v apt-get >/dev/null 2>&1; then
  echo "This setup helper currently supports Debian-based hosts only." >&2
  exit 1
fi

apt-get update
DEBIAN_FRONTEND=noninteractive apt-get install -y openssh-server
systemctl enable --now ssh

echo
echo "OpenSSH server is enabled."
echo "Use a regular, non-root account from the client."
echo "Configure host firewall access for the SSH port you intend to use."
echo "Do not enable the SnekProxy blocking-only nftables policy on this host."
