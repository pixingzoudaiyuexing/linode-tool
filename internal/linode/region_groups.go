package linode

// RegionGroups is used only for display grouping.
// Actual available regions are still fetched from Linode API.
var RegionGroups = map[string][]string{
	"亚洲 Asia": {
		"jp-tyo",
		"jp-osa",
		"sg",
		"ap-south",
	},
	"欧洲 Europe": {
		"eu",
	},
	"美洲 America": {
		"us",
		"ca",
	},
}

func RegionGroup(regionID string) string {
	for group, prefixes := range RegionGroups {
		for _, prefix := range prefixes {
			if len(regionID) >= len(prefix) && regionID[:len(prefix)] == prefix {
				return group
			}
		}
	}
	return "其他 Other"
}
