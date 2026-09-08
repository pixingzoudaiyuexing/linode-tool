package linode

import (
	"fmt"

	"github.com/linode/linodego"
)

// InstanceLabel returns predictable labels for batch-created nodes.
func InstanceLabel(prefix string, index int) string {
	if prefix == "" {
		prefix = "cg-node"
	}
	return fmt.Sprintf("%s-%03d", prefix, index)
}

func NextInstanceLabels(instances []linodego.Instance, count int) []string {
	used := make(map[string]struct{}, len(instances))
	for _, instance := range instances {
		used[instance.Label] = struct{}{}
	}

	labels := make([]string, 0, count)
	for index := 1; len(labels) < count; index++ {
		label := InstanceLabel("", index)
		if _, exists := used[label]; exists {
			continue
		}
		labels = append(labels, label)
	}
	return labels
}
