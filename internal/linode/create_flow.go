package linode

import (
	"context"
	"fmt"
	"io"

	"github.com/linode/linodego"
)

func CreateInteractive(ctx context.Context, client Client, prompt *Prompter) error {
	region, err := SelectRegion(ctx, client, prompt)
	if err != nil {
		return err
	}
	rootPassword, err := prompt.ReadPassword("Root Password: ")
	if err != nil {
		return fmt.Errorf("读取 Root Password 失败: %w", err)
	}
	count, err := prompt.ReadPositiveInt("创建数量: ")
	if err != nil {
		return fmt.Errorf("读取创建数量失败: %w", err)
	}

	config := CreateConfig{Region: region, RootPassword: rootPassword, Count: count}
	PrintCreatePlan(prompt.out, config)
	return CreateInstances(ctx, client, config, prompt.out)
}

func PrintCreatePlan(out io.Writer, config CreateConfig) {
	fmt.Fprintln(out, "\n创建计划:")
	fmt.Fprintln(out, "套餐:", DefaultType)
	fmt.Fprintln(out, "系统:", DefaultImage)
	fmt.Fprintln(out, "区域:", config.Region)
	fmt.Fprintln(out, "数量:", config.Count)
}

func CreateInstances(ctx context.Context, client Client, config CreateConfig, out io.Writer) error {
	instances, err := client.ListInstances(ctx, nil)
	if err != nil {
		return fmt.Errorf("获取已有实例失败: %w", err)
	}
	labels := NextInstanceLabels(instances, NormalizeCount(config.Count))

	failed := 0
	for index, label := range labels {
		fmt.Fprintf(out, "[%d/%d] 创建中...\n", index+1, len(labels))
		instance, err := client.CreateInstance(ctx, linodego.InstanceCreateOptions{
			Region:   config.Region,
			Type:     DefaultType,
			Image:    DefaultImage,
			Label:    label,
			RootPass: config.RootPassword,
		})
		if err != nil {
			failed++
			fmt.Fprintf(out, "[%d/%d] 创建失败: %v\n", index+1, len(labels), err)
			continue
		}

		firewall, err := CreateFirewallForInstance(ctx, client, *instance)
		if err != nil {
			failed++
			cleanupErr := client.DeleteInstance(ctx, instance.ID)
			if cleanupErr != nil {
				fmt.Fprintf(out, "[%d/%d] Firewall 创建失败: %v；实例清理失败，请手动删除实例 %d: %v\n", index+1, len(labels), err, instance.ID, cleanupErr)
			} else {
				fmt.Fprintf(out, "[%d/%d] Firewall 创建失败: %v；已删除刚创建的实例 %d\n", index+1, len(labels), err, instance.ID)
			}
			continue
		}

		fmt.Fprintf(out, "[%d/%d] 创建成功: %s (ID: %d, Firewall ID: %d)\n", index+1, len(labels), instance.Label, instance.ID, firewall.ID)
	}

	if failed > 0 {
		return fmt.Errorf("批量创建完成，但 %d/%d 台失败", failed, len(labels))
	}
	return nil
}
