package linode

// CreateConfig contains the fixed deployment settings used by linode-tool.
// The tool intentionally keeps the first version simple: Nanode + Debian 12.
type CreateConfig struct {
	Region       string
	RootPassword string
	Count        int
}

const (
	DefaultType  = "g6-nanode-1"
	DefaultImage = "linode/debian12"
)

func NormalizeCount(count int) int {
	if count < 1 {
		return 1
	}
	return count
}
