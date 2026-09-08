package linode

import (
	"strings"

	"github.com/linode/linodego"
)

type RegionGroup struct {
	Name      string
	Countries []string
	Prefixes  []string
}

// RegionGroups controls only the three requested menu groups. Region entries
// themselves always come from the Linode API.
var RegionGroups = []RegionGroup{
	{
		Name:      "亚洲 Asia",
		Countries: []string{"au", "id", "in", "jp", "sg"},
		Prefixes:  []string{"ap-", "au-", "id-", "in-", "jp-", "sg-"},
	},
	{
		Name:      "欧洲 Europe",
		Countries: []string{"de", "es", "fr", "gb", "it", "nl", "se"},
		Prefixes:  []string{"eu-", "de-", "es-", "fr-", "gb-", "it-", "nl-", "se-"},
	},
	{
		Name:      "美洲 America",
		Countries: []string{"br", "ca", "us"},
		Prefixes:  []string{"br-", "ca-", "us-"},
	},
}

func regionGroupIndex(region linodego.Region) int {
	country := strings.ToLower(region.Country)
	regionID := strings.ToLower(region.ID)
	for index, group := range RegionGroups {
		for _, item := range group.Countries {
			if country == item {
				return index
			}
		}
		for _, prefix := range group.Prefixes {
			if strings.HasPrefix(regionID, prefix) {
				return index
			}
		}
	}
	return -1
}
