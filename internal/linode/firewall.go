package linode

import "github.com/linode/linodego"

// Firewall rules for the tool are intentionally kept as a single place.
// The default policy for created nodes will be full inbound access.
type FirewallConfig struct {
	TCPPorts string
	UDPPorts string
}

func DefaultFirewallConfig() FirewallConfig {
	return FirewallConfig{
		TCPPorts: "1-65535",
		UDPPorts: "1-65535",
	}
}

// Keep the linodego import here while firewall API integration is finalized.
var _ *linodego.Client
