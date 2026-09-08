package linode

import (
	"context"
	"fmt"
	"io"
	"net"
	"sort"
	"text/tabwriter"
)

func PrintInstances(ctx context.Context, client Client, out io.Writer) error {
	instances, err := client.ListInstances(ctx, nil)
	if err != nil {
		return fmt.Errorf("获取实例失败: %w", err)
	}

	writer := tabwriter.NewWriter(out, 0, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\t名称\t区域\tIP\t状态")
	for _, item := range instances {
		fmt.Fprintf(writer, "%d\t%s\t%s\t%s\t%s\n", item.ID, item.Label, item.Region, instanceIPv4(item.IPv4), item.Status)
	}
	return writer.Flush()
}

func DeleteInteractive(ctx context.Context, client Client, prompt *Prompter) error {
	instances, err := client.ListInstances(ctx, nil)
	if err != nil {
		return fmt.Errorf("获取实例失败: %w", err)
	}
	if len(instances) == 0 {
		fmt.Fprintln(prompt.out, "没有可删除的实例。")
		return nil
	}

	sort.Slice(instances, func(i, j int) bool {
		if instances[i].Label == instances[j].Label {
			return instances[i].ID < instances[j].ID
		}
		return instances[i].Label < instances[j].Label
	})
	fmt.Fprintln(prompt.out, "请选择删除:")
	for index, instance := range instances {
		fmt.Fprintf(prompt.out, "%d. %s (ID: %d, %s)\n", index+1, instance.Label, instance.ID, instance.Region)
	}
	choice, err := prompt.ReadChoice("请输入序号: ", len(instances))
	if err != nil {
		return err
	}
	selected := instances[choice-1]
	confirmation, err := prompt.Read(fmt.Sprintf("输入 yes 确认删除 %s (ID: %d): ", selected.Label, selected.ID))
	if err != nil {
		return err
	}
	if confirmation != "yes" {
		fmt.Fprintln(prompt.out, "已取消删除。")
		return nil
	}
	if err := client.DeleteInstance(ctx, selected.ID); err != nil {
		return fmt.Errorf("删除实例 %s (ID: %d) 失败: %w", selected.Label, selected.ID, err)
	}
	fmt.Fprintf(prompt.out, "已删除: %s (ID: %d)\n", selected.Label, selected.ID)
	return nil
}

func instanceIPv4(addresses []*net.IP) string {
	for _, address := range addresses {
		if address != nil && address.To4() != nil && !address.IsPrivate() {
			return address.String()
		}
	}
	for _, address := range addresses {
		if address != nil && address.To4() != nil {
			return address.String()
		}
	}
	return "-"
}
