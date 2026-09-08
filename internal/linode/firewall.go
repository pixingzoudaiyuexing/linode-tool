package linode

// FirewallConfig defines the default firewall policy for created nodes.
// The MVP intentionally opens all ports as requested.
type FirewallConfig struct {
	TCPPorts string
	UDPPorts string
	IPv4     string
	IPv6     string
}

func DefaultFirewallConfig() FirewallConfig {
	return FirewallConfig{
		TCPPorts: "1-65535",
		UDPPorts: "1-65535",
		IPv4:     "0.0.0.0/0",
		IPv6:     "::/0",
	}
}
