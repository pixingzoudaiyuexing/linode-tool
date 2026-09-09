package linode

import (
	"context"
	"fmt"
	"io"
	"net"
	"sort"
	"text/tabwriter"

	"github.com/linode/linodego"
)

func PrintInstances(ctx context.Context, client Client, out io.Writer) error {
	instances, err := client.ListInstances(ctx, nil)
	if err != nil {
		return fmt.Errorf("获取实例失败: %w", err)
	}

	writer := tabwriter.NewWriter(out, 0, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\t名称\t区域\tIP\t状态")
	for _, item := range instances {
		fmt.Fprintf(writer, "%d\t%s\t%s\t%s\t%s\n", item.ID, item.Label, displayRegion(item.Region), instanceIPv4(item.IPv4), item.Status)
	}
	return writer.Flush()
}

func displayRegion(region string) string {
	if label, ok := RegionLabels[region]; ok && label != "" {
		return fmt.Sprintf("%s（%s）", region, label)
	}
	return region
}

func displayInstanceDetails(instance linodego.Instance) string {
	return fmt.Sprintf("ID: %d, 区域: %s, IP: %s", instance.ID, displayRegion(instance.Region), instanceIPv4(instance.IPv4))
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
		fmt.Fprintf(prompt.out, "%d. %s (%s)\n", index+1, instance.Label, displayInstanceDetails(instance))
	}
	allChoice := len(instances) + 1
	fmt.Fprintf(prompt.out, "%d. 全部删除\n", allChoice)
	choice, err := prompt.ReadChoice("请输入序号: ", allChoice)
	if err != nil {
		return err
	}
	if choice == allChoice {
		return deleteAllInstances(ctx, client, instances, prompt)
	}
	selected := instances[choice-1]
	confirmation, err := prompt.Read(fmt.Sprintf("确认删除 %s (%s)? [Y/n]: ", selected.Label, displayInstanceDetails(selected)))
	if err != nil {
		return err
	}
	if confirmation != "" && confirmation != "y" && confirmation != "Y" && confirmation != "yes" && confirmation != "YES" {
		fmt.Fprintln(prompt.out, "已取消删除。")
		return nil
	}
	if err := client.DeleteInstance(ctx, selected.ID); err != nil {
		return fmt.Errorf("删除实例 %s (%s) 失败: %w", selected.Label, displayInstanceDetails(selected), err)
	}
	fmt.Fprintf(prompt.out, "已删除: %s (%s)\n", selected.Label, displayInstanceDetails(selected))
	return nil
}

func deleteAllInstances(ctx context.Context, client Client, instances []linodego.Instance, prompt *Prompter) error {
	confirmation, err := prompt.Read("输入 yes 确认删除全部实例: ")
	if err != nil {
		return err
	}
	if confirmation != "yes" {
		fmt.Fprintln(prompt.out, "已取消全部删除。")
		return nil
	}

	failed := 0
	for _, instance := range instances {
		if err := client.DeleteInstance(ctx, instance.ID); err != nil {
			failed++
			fmt.Fprintf(prompt.out, "删除失败: %s (%s): %v\n", instance.Label, displayInstanceDetails(instance), err)
			continue
		}
		fmt.Fprintf(prompt.out, "已删除: %s (%s)\n", instance.Label, displayInstanceDetails(instance))
	}
	if failed > 0 {
		return fmt.Errorf("全部删除完成，但 %d/%d 台失败", failed, len(instances))
	}
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
