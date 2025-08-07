package goexec

import (
	"fmt"
	"os"
)

func SetEnvGOPrivate(hosts string) error {
	if err := os.Setenv("GOPRIVATE", hosts); err != nil {
		return fmt.Errorf("failed to set GOPRIVATE: %w", err)
	}
	return nil
}
