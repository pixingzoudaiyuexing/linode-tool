package linode

import (
	"fmt"

	"github.com/linode/linodego"
)

func PrintRegions(client *linodego.Client) {
	regions, err := client.ListRegions(nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Available Regions:")
	for _, r := range regions {
		fmt.Printf("%-15s %s\n", r.ID, r.Label)
	}
}
