package main

import (
	"fmt"
	"os"

	"github.com/pixingzoudaiyuexing/linode-tool/internal/linode"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		return
	}

	client, err := linode.NewClient()
	if err != nil {
		fmt.Println(err)
		return
	}

	switch os.Args[1] {
	case "regions":
		linode.PrintRegions(client)
	case "list":
		linode.PrintInstances(client)
	case "create":
		linode.CreateInteractive(client)
	case "delete":
		linode.DeleteInteractive(client)
	default:
		usage()
	}
}

func usage() {
	fmt.Println("linode-tool")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  regions   show available regions")
	fmt.Println("  create    create Debian 12 Nanode instance")
	fmt.Println("  list      list instances")
	fmt.Println("  delete    delete instance")
}
