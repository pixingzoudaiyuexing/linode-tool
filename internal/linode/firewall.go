package linode

import (
	"context"
	"fmt"
	"regexp"

	"github.com/linode/linodego"
)

const SharedFirewallLabel = "cg-allow-all"

var legacyFirewallLabelPattern = regexp.MustCompile(`^cg-node-[0-9]+-fw-[0-9]+$`)

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

// FindReusableFirewall returns the shared firewall created by the current
// version. For accounts upgraded from the old per-instance implementation it
// also reuses the first legacy cg-node-*-fw-* firewall, avoiding another
// active-service allocation just to migrate.
func FindReusableFirewall(ctx context.Context, client Client) (*linodego.Firewall, error) {
	firewalls, err := client.ListFirewalls(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("获取 Firewall 列表失败: %w", err)
	}

	var legacy *linodego.Firewall
	for i := range firewalls {
		firewall := &firewalls[i]
		if firewall.Label == SharedFirewallLabel {
			return firewall, nil
		}
		if legacy == nil && legacyFirewallLabelPattern.MatchString(firewall.Label) {
			legacy = firewall
		}
	}
	return legacy, nil
}

// CreateSharedFirewallForInstance creates the single shared allow-all firewall
// and binds the first instance in the same API call.
func CreateSharedFirewallForInstance(ctx context.Context, client Client, instance linodego.Instance) (*linodego.Firewall, error) {
	return client.CreateFirewall(ctx, linodego.FirewallCreateOptions{
		Label: SharedFirewallLabel,
		Rules: DefaultFirewallRules(),
		Devices: linodego.DevicesCreationOptions{
			Linodes: []int{instance.ID},
		},
	})
}

// AttachFirewallToInstance binds another Linode to the already existing shared
// firewall without creating a new Firewall service.
func AttachFirewallToInstance(ctx context.Context, client Client, firewallID, instanceID int) (*linodego.FirewallDevice, error) {
	return client.CreateFirewallDevice(ctx, firewallID, linodego.FirewallDeviceCreateOptions{
		ID:   instanceID,
		Type: linodego.FirewallDeviceLinode,
	})
}
