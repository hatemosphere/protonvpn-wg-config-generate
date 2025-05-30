# ProtonVPN WireGuard Config Generate

A Go program that generates WireGuard configuration files for ProtonVPN servers with automatic selection of the best servers from specified countries with support of generic filters.

## Features

- Authenticates with ProtonVPN using username/password
- Supports 2FA authentication
- Creates persistent WireGuard configurations (visible in ProtonVPN dashboard)
- Automatically selects the best server (highest score, lowest load) from specified countries
- Filters servers by tier (Plus/Free) and features (P2P support)
- Generates WireGuard configuration files
- Supports VPN accelerator feature
- IPv6 support

## Installation

1. Clone the repository:
```bash
git clone <your-repo-url>
cd protonvpn-wg-config-generate
```

2. Build the program:
```bash
go build -o protonvpn-wg-config-generate
```

## Usage

```bash
./protonvpn-wg-config-generate -username <username> -countries <country-codes> [options]
```

### Options

- `-username`: ProtonVPN username (optional, will prompt if not provided)
- `-countries`: Comma-separated list of country codes (e.g., US,NL,CH) **[Required]**
- `-output`: Output WireGuard configuration file (default: protonvpn.conf)
- `-ipv6`: Enable IPv6 support (default: false)
- `-dns`: Comma-separated list of DNS servers (defaults based on IPv6 setting)
- `-allowed-ips`: Comma-separated list of allowed IPs (defaults based on IPv6 setting)
- `-accelerator`: Enable VPN accelerator (default: true)
- `-api-url`: ProtonVPN API URL (default: https://vpn-api.proton.me)
- `-plus-only`: Use only Plus servers (default: true)
- `-p2p-only`: Use only P2P-enabled servers (default: true)
- `-device-name`: Device name for WireGuard config (auto-generated if empty)
- `-duration`: Certificate duration (default: 365d). Examples: 30m, 24h, 7d, 1h30m. Maximum: 365d
- `-clear-session`: Clear saved session and force re-authentication
- `-no-session`: Don't save or use session persistence
- `-force-refresh`: Force session refresh even if not close to expiration (requires re-authentication)
- `-session-duration`: Session cache duration (default: 0 = use API expiration). Examples: 12h, 24h, 7d. Max: 30d

### Examples

1. Generate config for best P2P server in US or Netherlands:
```bash
./protonvpn-wg-config-generate -username myusername -countries US,NL
```

2. Generate config with custom DNS and output file:
```bash
./protonvpn-wg-config-generate -username myusername -countries CH,DE -dns 1.1.1.1,8.8.8.8 -output switzerland.conf
```

3. Disable VPN accelerator:
```bash
./protonvpn-wg-config-generate -username myusername -countries US -accelerator=false
```

4. Generate config with 30-day duration:
```bash
./protonvpn-wg-config-generate -username myusername -countries US -duration 30d
```

5. Generate config without saving session (always prompt for password):
```bash
./protonvpn-wg-config-generate -username myusername -countries US -no-session
```

6. Use session with 24-hour expiration:
```bash
./protonvpn-wg-config-generate -username myusername -countries US -session-duration 24h
```

7. Enable IPv6 support:
```bash
./protonvpn-wg-config-generate -username myusername -countries US -ipv6
```

## IPv6 Support

By default, the tool generates IPv4-only configurations. When you enable IPv6 with the `-ipv6` flag:

- **Interface Address**: Both IPv4 (10.2.0.2/32) and IPv6 (2a07:b944::2:2/128) addresses are assigned
- **DNS Servers**: IPv4 DNS (10.2.0.1) and IPv6 DNS (2a07:b944::2:1) - both ProtonVPN's internal DNS servers
- **Allowed IPs**: Both IPv4 (0.0.0.0/0) and IPv6 (::/0) routes are included

You can override the defaults by explicitly specifying `-dns` and `-allowed-ips` flags.

## Authentication

The program supports the following authentication methods:

1. **Username/Password**: Enter your ProtonVPN credentials
2. **2FA**: If enabled, you'll be prompted for your 2FA code
3. **Mailbox Password**: If you have a second password, you'll be prompted for it

### Session Persistence

The program saves your authentication session to avoid re-entering credentials:
- Sessions are stored in `~/.protonvpn-session.json` with secure permissions (0600)
- ProtonVPN sessions expire after 30 days (from API `ExpiresIn` field)
- Session duration is configurable with `-session-duration` (default: 0 = use API's 30 days)
- Custom durations are capped at the API's expiration time
- Sessions show time until expiration when reused
- Sessions are automatically verified before use
- Sessions automatically refresh when less than 7 days remain
- Use `-clear-session` flag to force re-authentication
- Use `-force-refresh` flag to force refresh even if not expiring soon
- Use `-no-session` flag to disable session persistence entirely
- Sessions are user-specific and won't be used for different usernames

## Using the Generated Configuration

Once you have the WireGuard configuration file, you can use it with any WireGuard client:

### Linux
```bash
sudo wg-quick up ./protonvpn.conf
```

### macOS (with WireGuard installed)
```bash
sudo wg-quick up ./protonvpn.conf
```

### Windows/GUI clients
Import the configuration file into your WireGuard client.

## Requirements

- Go 1.24 or higher
- ProtonVPN account with Plus subscription (for P2P servers)

## Security Notes

- The program generates a new WireGuard private key for each run
- Configuration files contain sensitive information and are saved with 0600 permissions
- Never share your WireGuard configuration files
- Persistent configurations appear in your ProtonVPN dashboard and can be revoked there
- Certificates are valid for the specified duration (default: 365 days, max: 365 days)

## Project Structure

```
.
├── cmd/
│   └── protonvpn-wg/      # Main application entry point
├── internal/              # Private application code
│   ├── api/              # API types and constants
│   ├── auth/             # Authentication logic
│   ├── config/           # Configuration handling
│   └── vpn/              # VPN client and server selection
├── pkg/                  # Public packages
│   └── wireguard/        # WireGuard configuration generation
├── vendor/               # Vendored dependencies
├── Makefile              # Build automation
└── README.md             # This file
```

## Development

```bash
# Format code
make fmt

# Run tests
make test

# Run linter (requires golangci-lint)
make lint

# Build for multiple platforms
make build-all

# Clean build artifacts
make clean
```

## Troubleshooting

### CAPTCHA Verification Required (Error 9001)

If you encounter "CAPTCHA verification required" error:

1. **Login via ProtonVPN website first**: This can help establish your account as legitimate
2. **Try from a different IP**: VPN or residential IPs may work better than datacenter IPs
3. **Wait and retry**: Sometimes waiting a few hours helps
4. **Use saved sessions**: Once authenticated, sessions are saved to avoid repeated CAPTCHA challenges

### App Version Errors (Error 5003)

If you see "This version of the app is no longer supported":
- The app version headers are hardcoded and may become outdated
- Check ProtonVPN forums or GitHub for current working versions
- The tool currently uses `linux-vpn@4.2.0`

## License

This project uses code from ProtonVPN's go-vpn-lib which is licensed under GPLv3.
