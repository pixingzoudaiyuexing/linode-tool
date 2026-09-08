package linode

import "fmt"

// InstanceLabel returns predictable labels for batch-created nodes.
func InstanceLabel(prefix string, index int) string {
	if prefix == "" {
		prefix = "cg-node"
	}
	return fmt.Sprintf("%s-%03d", prefix, index)
}
