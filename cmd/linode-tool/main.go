package main

import (
	"context"
	"fmt"
	"os"

	"github.com/pixingzoudaiyuexing/linode-tool/internal/config"
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
		return runMenu()
	}

	command := args[1]
	switch command {
	case "help", "regions", "list", "create", "delete":
	default:
		usage()
		return fmt.Errorf("未知命令 %q", command)
	}
	if command == "help" {
		usage()
		return nil
	}

	prompter := linode.NewPrompter(os.Stdin, os.Stdout)
	client, err := clientFromEnvironmentOrPrompt(config.Token(), prompter)
	if err != nil {
		return err
	}
	ctx := context.Background()

	return runCommand(ctx, command, client, prompter)
}

func runMenu() error {
	prompter := linode.NewPrompter(os.Stdin, os.Stdout)
	client, err := clientFromEnvironmentOrPrompt(config.Token(), prompter)
	if err != nil {
		return err
	}
	ctx := context.Background()
	for {
		fmt.Fprintln(os.Stdout, "\nlinode-tool 主菜单")
		fmt.Fprintln(os.Stdout, "1. 创建实例")
		fmt.Fprintln(os.Stdout, "2. 查看实例")
		fmt.Fprintln(os.Stdout, "3. 删除实例")
		fmt.Fprintln(os.Stdout, "4. 查看地区")
		fmt.Fprintln(os.Stdout, "0. 退出")
		choice, err := prompter.ReadChoiceIncludingZero("请输入序号: ", 4)
		if err != nil {
			return err
		}
		if choice == 0 {
			fmt.Fprintln(os.Stdout, "已退出。")
			return nil
		}
		commands := []string{"", "create", "list", "delete", "regions"}
		if err := runCommand(ctx, commands[choice], client, prompter); err != nil {
			return err
		}
	}
}

func clientFromEnvironmentOrPrompt(environmentToken string, prompt *linode.Prompter) (linode.Client, error) {
	token, err := tokenFromEnvironmentOrPrompt(environmentToken, prompt)
	if err != nil {
		return nil, fmt.Errorf("读取 Linode API Token 失败: %w", err)
	}
	client, err := linode.NewClient(token)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func runCommand(ctx context.Context, command string, client linode.Client, prompter *linode.Prompter) error {
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

func tokenFromEnvironmentOrPrompt(environmentToken string, prompt *linode.Prompter) (string, error) {
	if environmentToken != "" {
		return environmentToken, nil
	}
	return prompt.ReadPassword("Linode API Token: ")
}

func usage() {
	fmt.Println("linode-tool")
	fmt.Println()
	fmt.Println("不带参数运行将进入交互菜单。")
	fmt.Println("命令:")
	fmt.Println("  help      显示帮助")
	fmt.Println("  create    创建 Debian 12 Nanode 实例")
	fmt.Println("  list      查看实例")
	fmt.Println("  delete    删除实例")
	fmt.Println("  regions   查看可用地区")
}
