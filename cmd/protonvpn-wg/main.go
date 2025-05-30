package main

import (
	"fmt"
	"os"

	"github.com/ProtonVPN/go-vpn-lib/ed25519"
	"protonvpn-wg-config-generate/internal/auth"
	"protonvpn-wg-config-generate/internal/config"
	"protonvpn-wg-config-generate/internal/vpn"
	"protonvpn-wg-config-generate/pkg/wireguard"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Parse configuration
	cfg, err := config.Parse()
	if err != nil {
		config.PrintUsage()
		return err
	}

	// Authenticate
	authClient := auth.NewClient(cfg)
	session, err := authClient.Authenticate()
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}
	fmt.Println("Authentication successful!")

	// Generate key pair
	keyPair, err := ed25519.NewKeyPair()
	if err != nil {
		return fmt.Errorf("failed to generate key pair: %w", err)
	}
	cfg.ClientPrivateKey = keyPair.ToX25519Base64()

	// Create VPN client
	vpnClient := vpn.NewClient(cfg, session)

	// Get VPN certificate
	vpnInfo, err := vpnClient.GetCertificate(keyPair)
	if err != nil {
		return fmt.Errorf("failed to get VPN certificate: %w", err)
	}

	// Get server list
	servers, err := vpnClient.GetServers()
	if err != nil {
		return fmt.Errorf("failed to get servers: %w", err)
	}

	// Select best server
	selector := vpn.NewServerSelector(cfg)
	server, err := selector.SelectBest(servers)
	if err != nil {
		return err
	}

	fmt.Printf("Selected server: %s (Country: %s, City: %s, Load: %d%%, Score: %.2f)\n",
		server.Name, server.ExitCountry, server.City, server.Load, server.Score)

	// Get best physical server
	physicalServer := vpn.GetBestPhysicalServer(server)
	if physicalServer == nil {
		return fmt.Errorf("no physical servers available")
	}

	// Generate WireGuard configuration
	generator := wireguard.NewConfigGenerator(cfg)
	if err := generator.Generate(server, physicalServer, cfg.ClientPrivateKey); err != nil {
		return fmt.Errorf("failed to generate WireGuard config: %w", err)
	}

	fmt.Printf("WireGuard configuration written to: %s\n", cfg.OutputFile)

	// Note about persistence
	if vpnInfo.DeviceName != "" {
		fmt.Printf("Device name: %s (visible in ProtonVPN dashboard)\n", vpnInfo.DeviceName)
	}

	return nil
}
