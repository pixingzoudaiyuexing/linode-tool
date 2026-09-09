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

type firewallDeviceCall struct {
	firewallID int
	options    linodego.FirewallDeviceCreateOptions
}

type fakeClient struct {
	regions            []linodego.Region
	instances          []linodego.Instance
	firewalls          []linodego.Firewall
	createOptions      []linodego.InstanceCreateOptions
	firewallOptions    []linodego.FirewallCreateOptions
	firewallDeviceCalls []firewallDeviceCall
	deletedIDs         []int
	createErr          error
	firewallErr        error
	firewallDeviceErr  error
	deleteErr          error
	listRegionsErr     error
	listInstancesErr   error
	listFirewallsErr   error
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

func (f *fakeClient) ListFirewalls(context.Context, *linodego.ListOptions) ([]linodego.Firewall, error) {
	return append([]linodego.Firewall(nil), f.firewalls...), f.listFirewallsErr
}

func (f *fakeClient) CreateFirewall(_ context.Context, options linodego.FirewallCreateOptions) (*linodego.Firewall, error) {
	f.firewallOptions = append(f.firewallOptions, options)
	if f.firewallErr != nil {
		return nil, f.firewallErr
	}
	return &linodego.Firewall{ID: 900 + len(f.firewallOptions), Label: options.Label, Rules: options.Rules}, nil
}

func (f *fakeClient) CreateFirewallDevice(_ context.Context, firewallID int, options linodego.FirewallDeviceCreateOptions) (*linodego.FirewallDevice, error) {
	f.firewallDeviceCalls = append(f.firewallDeviceCalls, firewallDeviceCall{firewallID: firewallID, options: options})
	if f.firewallDeviceErr != nil {
		return nil, f.firewallDeviceErr
	}
	return &linodego.FirewallDevice{ID: 1000 + len(f.firewallDeviceCalls)}, nil
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

func TestCreateInstancesUsesFixedProfileAndSharedFirewall(t *testing.T) {
	client := &fakeClient{instances: []linodego.Instance{{ID: 1, Label: "cg-node-001"}}}
	var output bytes.Buffer
	config := CreateConfig{Region: "jp-tyo-3", RootPassword: "secret-password", Count: 2}

	if err := CreateInstances(context.Background(), client, config, &output); err != nil {
		t.Fatalf("CreateInstances() error = %v", err)
	}
	if len(client.createOptions) != 2 {
		t.Fatalf("created %d instances, want 2", len(client.createOptions))
	}
	if len(client.firewallOptions) != 1 {
		t.Fatalf("created %d firewalls, want 1 shared firewall", len(client.firewallOptions))
	}
	if len(client.firewallDeviceCalls) != 1 {
		t.Fatalf("created %d firewall device bindings, want 1 for the second instance", len(client.firewallDeviceCalls))
	}

	for index, options := range client.createOptions {
		wantLabel := []string{"cg-node-002", "cg-node-003"}[index]
		if options.Label != wantLabel || options.Region != config.Region || options.Type != DefaultType || options.Image != DefaultImage || options.RootPass != config.RootPassword {
			t.Errorf("create options = %+v, want fixed profile with label %s", options, wantLabel)
		}
	}

	firewall := client.firewallOptions[0]
	if firewall.Label != SharedFirewallLabel {
		t.Errorf("firewall label = %q, want %q", firewall.Label, SharedFirewallLabel)
	}
	if len(firewall.Devices.Linodes) != 1 || firewall.Devices.Linodes[0] != 101 {
		t.Errorf("initial firewall devices = %v, want [101]", firewall.Devices.Linodes)
	}
	assertFirewallRules(t, firewall.Rules)

	binding := client.firewallDeviceCalls[0]
	if binding.firewallID != 901 || binding.options.ID != 102 || binding.options.Type != linodego.FirewallDeviceLinode {
		t.Errorf("binding = %+v, want firewall 901 -> linode 102", binding)
	}

	for _, expected := range []string{"[1/2] 创建中...", "[2/2] 创建中...", "创建成功", "Firewall ID: 901"} {
		if !strings.Contains(output.String(), expected) {
			t.Errorf("create output does not contain %q:\n%s", expected, output.String())
		}
	}
}

func TestCreateInstancesReusesExistingSharedFirewall(t *testing.T) {
	client := &fakeClient{firewalls: []linodego.Firewall{{ID: 777, Label: SharedFirewallLabel}}}
	var output bytes.Buffer

	if err := CreateInstances(context.Background(), client, CreateConfig{Region: "sg-sin-2", RootPassword: "secret-password", Count: 2}, &output); err != nil {
		t.Fatalf("CreateInstances() error = %v", err)
	}
	if len(client.firewallOptions) != 0 {
		t.Fatalf("created %d new firewalls, want 0", len(client.firewallOptions))
	}
	if len(client.firewallDeviceCalls) != 2 {
		t.Fatalf("firewall bindings = %d, want 2", len(client.firewallDeviceCalls))
	}
	for index, binding := range client.firewallDeviceCalls {
		if binding.firewallID != 777 || binding.options.ID != 101+index {
			t.Errorf("binding[%d] = %+v, want firewall 777 -> linode %d", index, binding, 101+index)
		}
	}
}

func TestCreateInstancesReusesLegacyManagedFirewall(t *testing.T) {
	client := &fakeClient{firewalls: []linodego.Firewall{{ID: 888, Label: "cg-node-001-fw-104674547"}}}
	var output bytes.Buffer

	if err := CreateInstances(context.Background(), client, CreateConfig{Region: "sg-sin-2", RootPassword: "secret-password", Count: 1}, &output); err != nil {
		t.Fatalf("CreateInstances() error = %v", err)
	}
	if len(client.firewallOptions) != 0 {
		t.Fatalf("created %d new firewalls, want 0 because legacy firewall should be reused", len(client.firewallOptions))
	}
	if len(client.firewallDeviceCalls) != 1 || client.firewallDeviceCalls[0].firewallID != 888 || client.firewallDeviceCalls[0].options.ID != 101 {
		t.Fatalf("legacy firewall binding = %+v, want firewall 888 -> linode 101", client.firewallDeviceCalls)
	}
}

func TestFindReusableFirewallPrefersSharedFirewall(t *testing.T) {
	client := &fakeClient{firewalls: []linodego.Firewall{
		{ID: 888, Label: "cg-node-001-fw-104674547"},
		{ID: 777, Label: SharedFirewallLabel},
	}}

	firewall, err := FindReusableFirewall(context.Background(), client)
	if err != nil {
		t.Fatalf("FindReusableFirewall() error = %v", err)
	}
	if firewall == nil || firewall.ID != 777 {
		t.Fatalf("FindReusableFirewall() = %+v, want shared firewall ID 777", firewall)
	}
}

func TestCreateInstancesCleansUpWhenFirewallCreateFails(t *testing.T) {
	client := &fakeClient{firewallErr: errors.New("firewall unavailable")}
	var output bytes.Buffer
	err := CreateInstances(context.Background(), client, CreateConfig{Region: "jp-osa", RootPassword: "secret-password", Count: 1}, &output)

	if err == nil {
		t.Fatal("CreateInstances() error = nil, want batch failure")
	}
	if len(client.deletedIDs) != 1 || client.deletedIDs[0] != 101 {
		t.Fatalf("deleted IDs = %v, want [101]", client.deletedIDs)
	}
	if !strings.Contains(output.String(), "Firewall 创建失败") || !strings.Contains(output.String(), "已删除刚创建的实例 101") {
		t.Fatalf("cleanup output missing:\n%s", output.String())
	}
}

func TestCreateInstancesCleansUpWhenFirewallAttachFails(t *testing.T) {
	client := &fakeClient{
		firewalls:         []linodego.Firewall{{ID: 777, Label: SharedFirewallLabel}},
		firewallDeviceErr: errors.New("attach unavailable"),
	}
	var output bytes.Buffer
	err := CreateInstances(context.Background(), client, CreateConfig{Region: "jp-osa", RootPassword: "secret-password", Count: 1}, &output)

	if err == nil {
		t.Fatal("CreateInstances() error = nil, want batch failure")
	}
	if len(client.deletedIDs) != 1 || client.deletedIDs[0] != 101 {
		t.Fatalf("deleted IDs = %v, want [101]", client.deletedIDs)
	}
	if len(client.firewallOptions) != 0 {
		t.Fatalf("created %d firewalls, want 0", len(client.firewallOptions))
	}
	if !strings.Contains(output.String(), "Firewall 绑定失败") || !strings.Contains(output.String(), "已删除刚创建的实例 101") {
		t.Fatalf("cleanup output missing:\n%s", output.String())
	}
}

func TestCreateSharedFirewallForInstanceUsesStableLabel(t *testing.T) {
	client := &fakeClient{}
	firewall, err := CreateSharedFirewallForInstance(context.Background(), client, linodego.Instance{ID: 101, Label: "cg-node-001"})
	if err != nil {
		t.Fatalf("CreateSharedFirewallForInstance() error = %v", err)
	}
	if firewall.Label != SharedFirewallLabel {
		t.Fatalf("firewall label = %q, want %q", firewall.Label, SharedFirewallLabel)
	}
	if len(client.firewallOptions) != 1 || len(client.firewallOptions[0].Devices.Linodes) != 1 || client.firewallOptions[0].Devices.Linodes[0] != 101 {
		t.Fatalf("firewall create options = %+v, want first instance 101 attached", client.firewallOptions)
	}
}

func TestDeleteInteractiveRequiresConfirmation(t *testing.T) {
	client := &fakeClient{instances: []linodego.Instance{
		{ID: 2, Label: "cg-node-002", Region: "jp-osa"},
		{ID: 1, Label: "cg-node-001", Region: "jp-tyo-3"},
	}}
	var output bytes.Buffer
	prompt := NewPrompter(strings.NewReader("2\n\n"), &output)

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

func TestDeleteInteractiveAllRequiresExplicitYes(t *testing.T) {
	client := &fakeClient{instances: []linodego.Instance{
		{ID: 2, Label: "cg-node-002", Region: "jp-osa"},
		{ID: 1, Label: "cg-node-001", Region: "jp-tyo-3"},
	}}
	var output bytes.Buffer
	prompt := NewPrompter(strings.NewReader("3\nyes\n"), &output)

	if err := DeleteInteractive(context.Background(), client, prompt); err != nil {
		t.Fatalf("DeleteInteractive() error = %v", err)
	}
	if len(client.deletedIDs) != 2 || client.deletedIDs[0] != 1 || client.deletedIDs[1] != 2 {
		t.Fatalf("deleted IDs = %v, want [1 2]", client.deletedIDs)
	}
	if !strings.Contains(output.String(), "3. 全部删除") || !strings.Contains(output.String(), "已删除: cg-node-001") {
		t.Fatalf("delete-all output missing:\n%s", output.String())
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
	if rules.InboundPolicy != "ACCEPT" || len(rules.Inbound) != 0 || rules.OutboundPolicy != "ACCEPT" || len(rules.Outbound) != 0 {
		t.Errorf("unexpected firewall policies: %+v", rules)
	}
}
