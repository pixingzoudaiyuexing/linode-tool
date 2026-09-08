package linode

import (
	"context"

	"github.com/linode/linodego"
)

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

func DefaultFirewallRules() linodego.FirewallRuleSet {
	return linodego.FirewallRuleSet{
		InboundPolicy:  "ACCEPT",
		Inbound:        []linodego.FirewallRule{},
		OutboundPolicy: "ACCEPT",
		Outbound:       []linodego.FirewallRule{},
	}
}

func CreateFirewallForInstance(ctx context.Context, client Client, instance linodego.Instance) (*linodego.Firewall, error) {
	return client.CreateFirewall(ctx, linodego.FirewallCreateOptions{
		Label: instance.Label + "-firewall",
		Rules: DefaultFirewallRules(),
		Devices: linodego.DevicesCreationOptions{
			Linodes: []int{instance.ID},
		},
	})
}
