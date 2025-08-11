package commit

import (
	"errors"
	"regexp"
	"strings"

	"github.com/selimacerbas/flow/internal/common"
)

func CleanMessage(s string) string {
	// mimic: tr -d '\n' (and drop \r just in case)
	s = strings.ReplaceAll(s, "\r", "")
	s = strings.ReplaceAll(s, "\n", "")
	return strings.TrimSpace(s)
}

func ValidateMessage(msg string) error {
	re := regexp.MustCompile(common.Pattern)

	msg = CleanMessage(msg)
	if re.MatchString(msg) {
		return nil
	}
	return errors.New("commit message does not match required format")
}
