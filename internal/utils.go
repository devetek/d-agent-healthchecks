package internal

import (
	"os"
	"strings"
)

func GetHostname() string {
	host, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return strings.ToLower(host)
}
