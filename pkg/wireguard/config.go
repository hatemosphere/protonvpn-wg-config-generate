// Package wireguard generates WireGuard configuration files.
package wireguard

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"text/template"

	"protonvpn-wg-config-generate/internal/api"
	"protonvpn-wg-config-generate/internal/config"
	"protonvpn-wg-config-generate/internal/constants"
)

// wireguardConfigTemplate is the template for generating WireGuard configuration
const wireguardConfigTemplate = `[Interface]
PrivateKey = {{.PrivateKey}}
{{.AddressLine}}
DNS = {{.DNS}}

[Peer]
PublicKey = {{.PublicKey}}
AllowedIPs = {{.AllowedIPs}}
Endpoint = {{.Endpoint}}:{{.Port}}
`

// configData holds the data for the WireGuard config template
type configData struct {
	PrivateKey  string
	AddressLine string
	DNS         string
	PublicKey   string
	AllowedIPs  string
	Endpoint    string
	Port        int
}

// ConfigGenerator generates WireGuard configuration files
type ConfigGenerator struct {
	config   *config.Config
	template *template.Template
}

// NewConfigGenerator creates a new configuration generator
func NewConfigGenerator(cfg *config.Config) *ConfigGenerator {
	tmpl := template.Must(template.New("wireguard").Parse(wireguardConfigTemplate))
	return &ConfigGenerator{
		config:   cfg,
		template: tmpl,
	}
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
	data := configData{
		PrivateKey:  privateKey,
		AddressLine: g.buildAddressLine(),
		DNS:         strings.Join(g.config.DNSServers, ", "),
		PublicKey:   physicalServer.X25519PublicKey,
		AllowedIPs:  strings.Join(g.config.AllowedIPs, ", "),
		Endpoint:    physicalServer.EntryIP,
		Port:        constants.WireGuardPort,
	}

	var buf bytes.Buffer
	if err := g.template.Execute(&buf, data); err != nil {
		// This should never happen with a valid template
		panic(fmt.Sprintf("failed to execute template: %v", err))
	}

	return buf.String()
}

func (g *ConfigGenerator) buildAddressLine() string {
	if g.config.EnableIPv6 {
		return fmt.Sprintf("Address = %s, %s", constants.WireGuardIPv4, constants.WireGuardIPv6)
	}
	return fmt.Sprintf("Address = %s", constants.WireGuardIPv4)
}
