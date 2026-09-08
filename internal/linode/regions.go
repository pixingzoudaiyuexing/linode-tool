package linode

import (
	"fmt"

	"github.com/linode/linodego"
)

// RegionDisplay contains human-readable region information.
type RegionDisplay struct {
	ID        string
	Continent string
	Country   string
	City      string
	Name      string
}

// RegionLabels keeps display names while available regions still come from Linode API.
var RegionLabels = map[string]RegionDisplay{
	"jp-tyo-3": {ID: "jp-tyo-3", Continent: "亚洲", Country: "日本", City: "东京", Name: "日本东京3"},
	"jp-osa":   {ID: "jp-osa", Continent: "亚洲", Country: "日本", City: "大阪", Name: "日本大阪"},
	"sg-sin-2": {ID: "sg-sin-2", Continent: "亚洲", Country: "新加坡", City: "新加坡", Name: "新加坡2"},
}

func PrintRegions(client *linodego.Client) {
	regions, err := client.ListRegions(nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("可用区域:")
	for _, r := range regions {
		if item, ok := RegionLabels[r.ID]; ok {
			fmt.Printf("%-15s %s (%s)\n", item.ID, item.Name, item.Continent)
			continue
		}
		fmt.Printf("%-15s %s\n", r.ID, r.Label)
	}
}
