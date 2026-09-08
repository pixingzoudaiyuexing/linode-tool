package linode

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// CreateOptions contains the fixed deployment profile used by linode-tool.
// The tool intentionally keeps the first version simple: Nanode + Debian 12.
type CreateOptions struct {
	Region string
	RootPassword string
	Quantity int
}

func ReadQuantity() int {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("开机数量: ")
	text, _ := reader.ReadString('\n')
	text = strings.TrimSpace(text)

	count, err := strconv.Atoi(text)
	if err != nil || count < 1 {
		return 1
	}
	return count
}

func PrintCreatePlan(options CreateOptions) {
	fmt.Println("创建计划:")
	fmt.Println("套餐:", DefaultType)
	fmt.Println("系统:", DefaultImage)
	fmt.Println("区域:", options.Region)
	fmt.Println("数量:", options.Quantity)
}
