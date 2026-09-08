package main

import (
	"context"
	"fmt"
	"os"

	"github.com/pixingzoudaiyuexing/linode-tool/internal/linode"
)

func main() {
	if err := run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, "错误:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) < 2 {
		usage()
		return nil
	}

	command := args[1]
	switch command {
	case "regions", "list", "create", "delete":
	default:
		usage()
		return fmt.Errorf("未知命令 %q", command)
	}

	client, err := linode.NewClient()
	if err != nil {
		return err
	}
	ctx := context.Background()
	prompter := linode.NewPrompter(os.Stdin, os.Stdout)

	switch command {
	case "regions":
		return linode.PrintRegions(ctx, client, os.Stdout)
	case "list":
		return linode.PrintInstances(ctx, client, os.Stdout)
	case "create":
		return linode.CreateInteractive(ctx, client, prompter)
	case "delete":
		return linode.DeleteInteractive(ctx, client, prompter)
	}
	return nil
}

func usage() {
	fmt.Println("linode-tool")
	fmt.Println()
	fmt.Println("命令:")
	fmt.Println("  create    创建 Debian 12 Nanode 实例")
	fmt.Println("  list      查看实例")
	fmt.Println("  delete    删除实例")
	fmt.Println("  regions   查看可用地区")
}
