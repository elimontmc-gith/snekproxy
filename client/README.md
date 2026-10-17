# SnekProxy client

The client is a Qt 6 desktop application. Its window collects the SSH host address, account, SSH port, and local SOCKS5 port; starts and stops the tunnel; displays SSH output; and automatically applies/restores the current user's system proxy settings. It uses OpenSSH dynamic forwarding, binds SOCKS5 to loopback only, and never asks for or stores passwords.

## Build

Requirements: C++17, CMake 3.16+, Qt 6 Network and Widgets development files, and the OpenSSH `ssh` client. On Debian or Ubuntu, install the build dependencies as root:

```sh
su -c "apt-get update && apt-get install -y build-essential cmake qt6-base-dev"
```

From this directory:

```sh
make
```

The application is created at `build/snekproxy-client` (`build/snekproxy-client.exe` on Windows). Build separately on the target OS; the client chooses Windows or Linux system proxy integration at compile time.

## Use

```sh
./build/snekproxy-client
```

Use a regular host account, not root. The client requires a previously verified SSH host key and a key available to `ssh-agent`; it runs SSH in batch mode rather than presenting an unsafe password prompt. First connect in a terminal with `ssh -p PORT USER@HOST`, verify the host fingerprint with the host administrator, and load the key into your agent if needed.

Once the local tunnel is verified, the app updates the GNOME desktop proxy on Linux or current-user WinINET proxy settings on Windows. **Disable Tunnel** stops SSH and restores the previously saved values; the saved recovery record is also restored at next startup after a crash. Linux support currently requires GNOME and `gsettings`; unsupported desktop environments fail without changing settings. Some applications ignore system proxy settings. This is not a system-wide kill switch and does not guarantee that every application or DNS request is proxied.
