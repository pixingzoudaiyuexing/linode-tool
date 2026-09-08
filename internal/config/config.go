package config

import (
	"os"
)

func Token() string {
	return os.Getenv("LINODE_TOKEN")
}
