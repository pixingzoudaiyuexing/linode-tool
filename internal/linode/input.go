package linode

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// ReadInput provides the common CLI input flow used by create commands.
func ReadInput(prompt string) string {
	fmt.Print(prompt)
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		return strings.TrimSpace(scanner.Text())
	}
	return ""
}

func ReadCount() int {
	for {
		value := ReadInput("开机数量: ")
		count, err := strconv.Atoi(value)
		if err == nil && count > 0 {
			return count
		}
		fmt.Println("请输入正确数量")
	}
}
