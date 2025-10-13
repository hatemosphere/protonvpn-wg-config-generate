# ProtonVPN WireGuard Config Generate

A Go program that generates WireGuard configuration files for ProtonVPN servers with automatic selection of the best servers from specified countries with support of generic filters.

## Motivation

The motivation to write this was simple - I wanted to automatically rotate VPN servers on my private HTPC Linux host running WireGuard, and since I'm already paying for a Proton bundle subscription, why not just use theirs? Unfortunately, as of the time of writing, they didn't have a good programmatic headless way to generate WireGuard profiles, so I did my research and reverse-engineered their APIs (which was a pain in the butt) and created this. The idea is to run this code as a daemon and restart the WireGuard client on profile file change.

## Features

- Authenticates with ProtonVPN using username/password
- Opnionally persists and refreshes login session, to function in headless mode after entering password once
- Supports 2FA authentication
- Creates persistent WireGuard configurations (visible in ProtonVPN dashboard)
- Automatically selects the best server (highest score, lowest load) from specified countries
- Supports both Free tier and paid tier servers (Plus and ProtonMail)
- Filters servers by features (P2P support, Secure Core)
- Generates WireGuard configuration files
- Supports VPN accelerator feature
- IPv6 support

## Installation

1. Clone the repository:
```bash
git clone https://github.com/hatemosphere/protonvpn-wg-config-generate
cd protonvpn-wg-config-generate
```

2. Build the program:
```bash
make build
```

Or manually with Go:
```bash
go build -o build/protonvpn-wg-config-generate cmd/protonvpn-wg/main.go
```

## Usage

```bash
./build/protonvpn-wg-config-generate -username <username> -countries <country-codes> [options]
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
- `-p2p-only`: Use only P2P-enabled servers (default: true)
- `-secure-core`: Use only Secure Core servers for multi-hop VPN (default: false)
- `-free-only`: Use only Free tier servers (tier 0) (default: false)
- `-device-name`: Device name for WireGuard config (auto-generated if empty)
- `-debug`: Enable debug output showing all filtered servers (default: false)
- `-duration`: Certificate duration (default: 365d). Examples: 30m, 24h, 7d, 1h30m. Maximum: 365d
- `-clear-session`: Clear saved session and force re-authentication
- `-no-session`: Don't save or use session persistence
- `-force-refresh`: Force session refresh even if not close to expiration (requires re-authentication)
- `-session-duration`: Session cache duration (default: 0 = use API expiration). Examples: 12h, 24h, 7d. Max: 30d

### Examples

1. Generate config for best P2P server in US or Netherlands:
```bash
./build/protonvpn-wg-config-generate -username myusername -countries US,NL
```

2. Generate config with custom DNS and output file:
```bash
./build/protonvpn-wg-config-generate -username myusername -countries CH,DE -dns 1.1.1.1,8.8.8.8 -output switzerland.conf
```

3. Disable VPN accelerator:
```bash
./build/protonvpn-wg-config-generate -username myusername -countries US -accelerator=false
```

4. Generate config with 30-day duration:
```bash
./build/protonvpn-wg-config-generate -username myusername -countries US -duration 30d
```

5. Generate config without saving session (always prompt for password):
```bash
./build/protonvpn-wg-config-generate -username myusername -countries US -no-session
```

6. Use session with 24-hour expiration:
```bash
./build/protonvpn-wg-config-generate -username myusername -countries US -session-duration 24h
```

7. Enable IPv6 support:
```bash
./build/protonvpn-wg-config-generate -username myusername -countries US -ipv6
```

8. Use Secure Core servers for enhanced privacy:
```bash
./build/protonvpn-wg-config-generate -username myusername -countries NL,US -secure-core
```

9. Debug mode to see all available servers:
```bash
./build/protonvpn-wg-config-generate -username myusername -countries US -debug
```

10. Use Free tier servers only:
```bash
./build/protonvpn-wg-config-generate -username myusername -countries US,NL -free-only
```

## IPv6 Support

By default, the tool generates IPv4-only configurations. When you enable IPv6 with the `-ipv6` flag:

- **Interface Address**: Both IPv4 (10.2.0.2/32) and IPv6 (2a07:b944::2:2/128) addresses are assigned
- **DNS Servers**: IPv4 DNS (10.2.0.1) and IPv6 DNS (2a07:b944::2:1) - both ProtonVPN's internal DNS servers
- **Allowed IPs**: Both IPv4 (0.0.0.0/0) and IPv6 (::/0) routes are included

You can override the defaults by explicitly specifying `-dns` and `-allowed-ips` flags.

## Secure Core

Secure Core is ProtonVPN's premium feature that routes your traffic through multiple servers before leaving the VPN network:

- First hop: Through secure servers in privacy-friendly countries (Switzerland, Iceland, Sweden)
- Second hop: To your chosen destination country
- Provides additional protection against network-based attacks
- Use the `-secure-core` flag to enable this feature
- Note: Secure Core servers may have higher latency due to the multi-hop routing

**Important Notes:**
- Secure Core servers don't support P2P, so P2P filtering is automatically disabled when using `-secure-core`
- The country filter always applies to **exit countries** - where your traffic appears to come from
- Server names show both entry and exit countries (e.g., "IS-NL#1" = Iceland → Netherlands)
- Entry countries for Secure Core are always privacy-friendly: Switzerland (CH), Iceland (IS), Sweden (SE)

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

## Server Tier Support

By default, the tool excludes Free tier servers and only uses paid tier servers (Plus and ProtonMail):
- **Free tier (tier 0)**: Available with `-free-only` flag. Limited server selection, no P2P support
- **Plus tier (tier 2)**: Default. Full feature support including P2P and Secure Core
- **ProtonMail tier (tier 3)**: Default. Included with Proton bundle subscriptions

**Important Notes:**
- When using `-free-only`, P2P filtering is automatically disabled since Free servers don't support P2P
- Free tier servers have limited availability and may have higher load
- Secure Core requires Plus or higher subscription

## Requirements

- Go 1.25 or higher
- ProtonVPN account (Free tier or paid subscription)

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
│       └── main.go        # CLI entry point
├── internal/              # Private application code
│   ├── api/              # API types and data structures
│   │   └── types.go      # ProtonVPN API response types
│   ├── auth/             # Authentication logic
│   │   ├── auth.go       # SRP authentication implementation
│   │   ├── errors.go     # Custom error types
│   │   └── session.go    # Session management and refresh
│   ├── config/           # Configuration handling
│   │   ├── flags.go      # Command-line flag parsing
│   │   └── types.go      # Config struct and validation
│   ├── constants/        # Application constants
│   │   ├── api.go        # API endpoints and headers
│   │   ├── defaults.go   # Default configuration values
│   │   ├── session.go    # Session-related constants
│   │   └── wireguard.go  # WireGuard network constants
│   └── vpn/              # VPN functionality
│       ├── client.go     # Certificate generation
│       └── servers.go    # Server selection logic
├── pkg/                  # Public packages
│   ├── timeutil/         # Time and duration utilities
│   │   ├── formatter.go  # Duration formatting
│   │   └── parser.go     # Duration parsing
│   ├── validation/       # Input validation
│   │   └── validation.go # Username and country code validation
│   └── wireguard/        # WireGuard configuration
│       ├── config.go     # Config file generation
│       └── config_test.go # Config generation tests
├── vendor/               # Vendored dependencies
├── Makefile              # Build automation
├── go.mod                # Go module definition
├── go.sum                # Module checksums
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
