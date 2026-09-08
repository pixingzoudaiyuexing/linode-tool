package linode

import (
	"context"
	"fmt"
	"io"
	"sort"

	"github.com/linode/linodego"
)

func PrintRegions(ctx context.Context, client Client, out io.Writer) error {
	grouped, err := listGroupedRegions(ctx, client)
	if err != nil {
		return err
	}

	for index, group := range RegionGroups {
		fmt.Fprintln(out, group.Name)
		for _, region := range grouped[index] {
			fmt.Fprintf(out, "  %-16s %s\n", region.ID, regionDisplayName(region))
		}
		fmt.Fprintln(out)
	}
	return nil
}

func SelectRegion(ctx context.Context, client Client, prompt *Prompter) (string, error) {
	grouped, err := listGroupedRegions(ctx, client)
	if err != nil {
		return "", err
	}

	fmt.Fprintln(prompt.out, "请选择地区:")
	for index, group := range RegionGroups {
		fmt.Fprintf(prompt.out, "%d. %s\n", index+1, group.Name)
	}
	groupChoice, err := prompt.ReadChoice("请输入序号: ", len(RegionGroups))
	if err != nil {
		return "", err
	}

	regions := grouped[groupChoice-1]
	if len(regions) == 0 {
		return "", fmt.Errorf("%s 当前没有可用地区", RegionGroups[groupChoice-1].Name)
	}

	fmt.Fprintf(prompt.out, "\n%s\n\n", RegionGroups[groupChoice-1].Name)
	for index, region := range regions {
		fmt.Fprintf(prompt.out, "%d. %s\n   %s\n\n", index+1, regionDisplayName(region), region.ID)
	}
	regionChoice, err := prompt.ReadChoice("请输入地区序号: ", len(regions))
	if err != nil {
		return "", err
	}
	return regions[regionChoice-1].ID, nil
}

func listGroupedRegions(ctx context.Context, client Client) (map[int][]linodego.Region, error) {
	regions, err := client.ListRegions(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("获取地区失败: %w", err)
	}

	grouped := make(map[int][]linodego.Region, len(RegionGroups))
	for _, region := range regions {
		if !regionSupportsLinodes(region) {
			continue
		}
		groupIndex := regionGroupIndex(region)
		if groupIndex >= 0 {
			grouped[groupIndex] = append(grouped[groupIndex], region)
		}
	}
	for index := range RegionGroups {
		sort.Slice(grouped[index], func(i, j int) bool {
			return grouped[index][i].ID < grouped[index][j].ID
		})
	}
	return grouped, nil
}

func regionSupportsLinodes(region linodego.Region) bool {
	if region.Status != "" && region.Status != "ok" {
		return false
	}
	if len(region.Capabilities) == 0 {
		return true
	}
	for _, capability := range region.Capabilities {
		if capability == linodego.CapabilityLinodes {
			return true
		}
	}
	return false
}

func regionDisplayName(region linodego.Region) string {
	if name, ok := RegionLabels[region.ID]; ok {
		return name
	}
	if region.Label != "" {
		return region.Label
	}
	return region.ID
}
