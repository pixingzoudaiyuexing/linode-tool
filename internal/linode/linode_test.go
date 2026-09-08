package linode

import (
	"bytes"
	"context"
	"errors"
	"net"
	"strings"
	"testing"

	"github.com/linode/linodego"
)

type fakeClient struct {
	regions          []linodego.Region
	instances        []linodego.Instance
	createOptions    []linodego.InstanceCreateOptions
	firewallOptions  []linodego.FirewallCreateOptions
	deletedIDs       []int
	createErr        error
	firewallErr      error
	deleteErr        error
	listRegionsErr   error
	listInstancesErr error
}

func (f *fakeClient) ListRegions(context.Context, *linodego.ListOptions) ([]linodego.Region, error) {
	return f.regions, f.listRegionsErr
}

func (f *fakeClient) ListInstances(context.Context, *linodego.ListOptions) ([]linodego.Instance, error) {
	return append([]linodego.Instance(nil), f.instances...), f.listInstancesErr
}

func (f *fakeClient) CreateInstance(_ context.Context, options linodego.InstanceCreateOptions) (*linodego.Instance, error) {
	f.createOptions = append(f.createOptions, options)
	if f.createErr != nil {
		return nil, f.createErr
	}
	return &linodego.Instance{ID: 100 + len(f.createOptions), Label: options.Label, Region: options.Region}, nil
}

func (f *fakeClient) DeleteInstance(_ context.Context, id int) error {
	f.deletedIDs = append(f.deletedIDs, id)
	return f.deleteErr
}

func (f *fakeClient) CreateFirewall(_ context.Context, options linodego.FirewallCreateOptions) (*linodego.Firewall, error) {
	f.firewallOptions = append(f.firewallOptions, options)
	if f.firewallErr != nil {
		return nil, f.firewallErr
	}
	return &linodego.Firewall{ID: 900 + len(f.firewallOptions), Label: options.Label}, nil
}

func TestSelectRegionUsesDynamicAPIResults(t *testing.T) {
	client := &fakeClient{regions: []linodego.Region{
		{ID: "jp-tyo-3", Country: "jp", Label: "Tokyo 3, JP", Status: "ok", Capabilities: []string{linodego.CapabilityLinodes}},
		{ID: "jp-osa", Country: "jp", Label: "Osaka, JP", Status: "ok", Capabilities: []string{linodego.CapabilityLinodes}},
		{ID: "sg-disabled", Country: "sg", Label: "Disabled", Status: "offline", Capabilities: []string{linodego.CapabilityLinodes}},
		{ID: "us-storage", Country: "us", Label: "Storage only", Status: "ok", Capabilities: []string{linodego.CapabilityObjectStorage}},
	}}
	var output bytes.Buffer
	prompt := NewPrompter(strings.NewReader("1\n2\n"), &output)

	region, err := SelectRegion(context.Background(), client, prompt)
	if err != nil {
		t.Fatalf("SelectRegion() error = %v", err)
	}
	if region != "jp-tyo-3" {
		t.Fatalf("SelectRegion() = %q, want jp-tyo-3", region)
	}
	for _, expected := range []string{"亚洲 Asia", "欧洲 Europe", "美洲 America", "日本大阪", "jp-osa", "日本东京3", "jp-tyo-3"} {
		if !strings.Contains(output.String(), expected) {
			t.Errorf("menu output does not contain %q:\n%s", expected, output.String())
		}
	}
	if strings.Contains(output.String(), "sg-disabled") || strings.Contains(output.String(), "us-storage") {
		t.Errorf("menu output contains unavailable region:\n%s", output.String())
	}
}

func TestCreateInstancesUsesFixedProfileAndFirewall(t *testing.T) {
	client := &fakeClient{instances: []linodego.Instance{{ID: 1, Label: "cg-node-001"}}}
	var output bytes.Buffer
	config := CreateConfig{Region: "jp-tyo-3", RootPassword: "secret-password", Count: 2}

	if err := CreateInstances(context.Background(), client, config, &output); err != nil {
		t.Fatalf("CreateInstances() error = %v", err)
	}
	if len(client.createOptions) != 2 || len(client.firewallOptions) != 2 {
		t.Fatalf("created %d instances and %d firewalls, want 2 each", len(client.createOptions), len(client.firewallOptions))
	}
	for index, options := range client.createOptions {
		wantLabel := []string{"cg-node-002", "cg-node-003"}[index]
		if options.Label != wantLabel || options.Region != config.Region || options.Type != DefaultType || options.Image != DefaultImage || options.RootPass != config.RootPassword {
			t.Errorf("create options = %+v, want fixed profile with label %s", options, wantLabel)
		}

		firewall := client.firewallOptions[index]
		if len(firewall.Devices.Linodes) != 1 || firewall.Devices.Linodes[0] != 101+index {
			t.Errorf("firewall devices = %v, want instance %d", firewall.Devices.Linodes, 101+index)
		}
		assertFirewallRules(t, firewall.Rules)
	}
	for _, expected := range []string{"[1/2] 创建中...", "[2/2] 创建中...", "创建成功"} {
		if !strings.Contains(output.String(), expected) {
			t.Errorf("create output does not contain %q:\n%s", expected, output.String())
		}
	}
}

func TestCreateInstancesCleansUpWhenFirewallFails(t *testing.T) {
	client := &fakeClient{firewallErr: errors.New("firewall unavailable")}
	var output bytes.Buffer
	err := CreateInstances(context.Background(), client, CreateConfig{Region: "jp-osa", RootPassword: "secret-password", Count: 1}, &output)

	if err == nil {
		t.Fatal("CreateInstances() error = nil, want batch failure")
	}
	if len(client.deletedIDs) != 1 || client.deletedIDs[0] != 101 {
		t.Fatalf("deleted IDs = %v, want [101]", client.deletedIDs)
	}
	if !strings.Contains(output.String(), "已删除刚创建的实例 101") {
		t.Fatalf("cleanup output missing:\n%s", output.String())
	}
}

func TestDeleteInteractiveRequiresConfirmation(t *testing.T) {
	client := &fakeClient{instances: []linodego.Instance{
		{ID: 2, Label: "cg-node-002", Region: "jp-osa"},
		{ID: 1, Label: "cg-node-001", Region: "jp-tyo-3"},
	}}
	var output bytes.Buffer
	prompt := NewPrompter(strings.NewReader("2\nyes\n"), &output)

	if err := DeleteInteractive(context.Background(), client, prompt); err != nil {
		t.Fatalf("DeleteInteractive() error = %v", err)
	}
	if len(client.deletedIDs) != 1 || client.deletedIDs[0] != 2 {
		t.Fatalf("deleted IDs = %v, want [2]", client.deletedIDs)
	}
	if !strings.Contains(output.String(), "已删除: cg-node-002") {
		t.Fatalf("delete output missing:\n%s", output.String())
	}
}

func TestPrintInstancesIncludesRequiredFields(t *testing.T) {
	ip := net.ParseIP("203.0.113.10")
	client := &fakeClient{instances: []linodego.Instance{{
		ID: 123456, Label: "cg-node-001", Region: "jp-tyo-3", IPv4: []*net.IP{&ip}, Status: linodego.InstanceRunning,
	}}}
	var output bytes.Buffer

	if err := PrintInstances(context.Background(), client, &output); err != nil {
		t.Fatalf("PrintInstances() error = %v", err)
	}
	for _, expected := range []string{"ID", "名称", "区域", "IP", "状态", "123456", "cg-node-001", "jp-tyo-3", "203.0.113.10", "running"} {
		if !strings.Contains(output.String(), expected) {
			t.Errorf("list output does not contain %q:\n%s", expected, output.String())
		}
	}
}

func assertFirewallRules(t *testing.T, rules linodego.FirewallRuleSet) {
	t.Helper()
	if rules.InboundPolicy != "DROP" || rules.OutboundPolicy != "ACCEPT" || len(rules.Outbound) != 0 {
		t.Errorf("unexpected firewall policies: %+v", rules)
	}
	if len(rules.Inbound) != 2 {
		t.Fatalf("inbound rules = %d, want 2", len(rules.Inbound))
	}
	for index, rule := range rules.Inbound {
		wantProtocol := []linodego.NetworkProtocol{linodego.TCP, linodego.UDP}[index]
		if rule.Action != "ACCEPT" || rule.Protocol != wantProtocol || rule.Ports != "1-65535" {
			t.Errorf("inbound rule = %+v, want ACCEPT %s 1-65535", rule, wantProtocol)
		}
		if rule.Addresses.IPv4 == nil || len(*rule.Addresses.IPv4) != 1 || (*rule.Addresses.IPv4)[0] != "0.0.0.0/0" {
			t.Errorf("IPv4 addresses = %v, want 0.0.0.0/0", rule.Addresses.IPv4)
		}
		if rule.Addresses.IPv6 == nil || len(*rule.Addresses.IPv6) != 1 || (*rule.Addresses.IPv6)[0] != "::/0" {
			t.Errorf("IPv6 addresses = %v, want ::/0", rule.Addresses.IPv6)
		}
	}
}
