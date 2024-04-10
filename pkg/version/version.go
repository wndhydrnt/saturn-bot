package version

import (
	"fmt"
	"strings"
)

var (
	Version  = "dev"
	Hash     = ""
	DateTime = ""
)

// String returns the full version string.
func String() string {
	return strings.TrimSpace(fmt.Sprintf("%s %s %s", Version, DateTime, Hash))
}
