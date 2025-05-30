package wireguard

import (
	"strings"
	"testing"

	"protonvpn-wg-config-generate/internal/api"
	"protonvpn-wg-config-generate/internal/config"
)

func TestConfigGeneration(t *testing.T) {
	cfg := &config.Config{
		DNSServers: []string{"10.2.0.1"},
		AllowedIPs: []string{"0.0.0.0/0"},
		OutputFile: "test.conf",
	}

	generator := NewConfigGenerator(cfg)

	server := &api.LogicalServer{
		Name: "Test-Server",
	}

	physicalServer := &api.PhysicalServer{
		EntryIP:         "192.168.1.1",
		X25519PublicKey: "testPublicKey123=",
	}

	privateKey := "testPrivateKey456="

	result := generator.buildConfig(server, physicalServer, privateKey)

	// Check that the config has proper structure
	lines := strings.Split(result, "\n")

	// Verify section headers
	if lines[0] != "[Interface]" {
		t.Errorf("Expected first line to be '[Interface]', got '%s'", lines[0])
	}

	// Check for proper indentation (no tabs/spaces before section headers)
	for i, line := range lines {
		if strings.HasPrefix(line, "[") && (strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t")) {
			t.Errorf("Line %d has unexpected indentation before section header: '%s'", i+1, line)
		}
	}

	// Verify content structure
	expectedContent := []string{
		"[Interface]",
		"PrivateKey = testPrivateKey456=",
		"Address = 10.2.0.2/32",
		"DNS = 10.2.0.1",
		"",
		"[Peer]",
		"PublicKey = testPublicKey123=",
		"AllowedIPs = 0.0.0.0/0",
		"Endpoint = 192.168.1.1:51820",
	}

	for i, expected := range expectedContent {
		if i >= len(lines) {
			t.Errorf("Missing line %d: expected '%s'", i+1, expected)
			continue
		}
		if lines[i] != expected {
			t.Errorf("Line %d mismatch:\nExpected: '%s'\nGot:      '%s'", i+1, expected, lines[i])
		}
	}
}

func TestConfigGenerationWithIPv6(t *testing.T) {
	cfg := &config.Config{
		DNSServers: []string{"10.2.0.1", "2a07:b944::2:1"},
		AllowedIPs: []string{"0.0.0.0/0", "::/0"},
		OutputFile: "test.conf",
		EnableIPv6: true,
	}

	generator := NewConfigGenerator(cfg)

	server := &api.LogicalServer{
		Name: "Test-Server",
	}

	physicalServer := &api.PhysicalServer{
		EntryIP:         "192.168.1.1",
		X25519PublicKey: "testPublicKey123=",
	}

	privateKey := "testPrivateKey456="

	result := generator.buildConfig(server, physicalServer, privateKey)

	// Check that IPv6 address is included
	if !strings.Contains(result, "Address = 10.2.0.2/32, 2a07:b944::2:2/128") {
		t.Errorf("Expected IPv6 address in config, got:\n%s", result)
	}

	// Check DNS servers
	if !strings.Contains(result, "DNS = 10.2.0.1, 2a07:b944::2:1") {
		t.Errorf("Expected both DNS servers in config, got:\n%s", result)
	}

	// Check AllowedIPs
	if !strings.Contains(result, "AllowedIPs = 0.0.0.0/0, ::/0") {
		t.Errorf("Expected both IPv4 and IPv6 in AllowedIPs, got:\n%s", result)
	}
}
