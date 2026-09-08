package linode

import (
	"fmt"

	"github.com/linode/linodego"
)

const (
	DefaultType  = "g6-nanode-1"
	DefaultImage = "linode/debian12"
)

func PrintInstances(client *linodego.Client) {
	instances, err := client.ListInstances(nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, item := range instances {
		fmt.Printf("%d %s %s\n", item.ID, item.Label, item.Status)
	}
}

func CreateInteractive(client *linodego.Client) {
	fmt.Println("create flow will create:")
	fmt.Println("Type:", DefaultType)
	fmt.Println("Image:", DefaultImage)
	fmt.Println("TODO: region/password input and firewall attach")
}

func DeleteInteractive(client *linodego.Client) {
	fmt.Println("TODO: delete instance")
}
