package goexec

import (
	"fmt"
	"os"
	"strings"
)

func SetEnvGOPrivate(hosts []string) error {
	if len(hosts) == 0 {
		return nil // nothing to set
	}

	commaSeparated := strings.Join(hosts, ",")
	fmt.Println("Setting GOPRIVATE to:", commaSeparated)

	if err := os.Setenv("GOPRIVATE", commaSeparated); err != nil {
		return fmt.Errorf("failed to set GOPRIVATE: %w", err)
	}

	return nil
}

