// Package wireguard generates WireGuard configuration files.
package wireguard

import (
	"fmt"
	"os"
	"strings"

	"protonvpn-wg-config-generate/internal/api"
	"protonvpn-wg-config-generate/internal/config"
)

// ConfigGenerator generates WireGuard configuration files
type ConfigGenerator struct {
	config *config.Config
}

// NewConfigGenerator creates a new configuration generator
func NewConfigGenerator(cfg *config.Config) *ConfigGenerator {
	return &ConfigGenerator{config: cfg}
}

// Generate creates a WireGuard configuration file
func (g *ConfigGenerator) Generate(server *api.LogicalServer, physicalServer *api.PhysicalServer, privateKey string) error {
	content := g.buildConfig(server, physicalServer, privateKey)

	if err := os.WriteFile(g.config.OutputFile, []byte(content), 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func (g *ConfigGenerator) buildConfig(_ *api.LogicalServer, physicalServer *api.PhysicalServer, privateKey string) string {
	addressLine := g.buildAddressLine()

	return fmt.Sprintf(`[Interface]
PrivateKey = %s
%s
DNS = %s

[Peer]
PublicKey = %s
AllowedIPs = %s
Endpoint = %s:51820
`,
		privateKey,
		addressLine,
		strings.Join(g.config.DNSServers, ", "),
		physicalServer.X25519PublicKey,
		strings.Join(g.config.AllowedIPs, ", "),
		physicalServer.EntryIP,
	)
}

func (g *ConfigGenerator) buildAddressLine() string {
	// Check if IPv6 is included in allowed IPs
	hasIPv6 := false
	for _, ip := range g.config.AllowedIPs {
		if strings.Contains(ip, "::/0") {
			hasIPv6 = true
			break
		}
	}

	if hasIPv6 {
		return "Address = 10.2.0.2/32, fd00::2/128"
	}
	return "Address = 10.2.0.2/32"
}
