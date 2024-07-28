package version

import (
	"fmt"
	"runtime"
	"strings"
)

var (
	Version  = "dev"
	Hash     = ""
	DateTime = ""
)

// String returns the full version string.
func String() string {
	return strings.TrimSpace(fmt.Sprintf(
		"%s %s %s/%s %s %s",
		Version,
		runtime.Version(),
		runtime.GOOS,
		runtime.GOARCH,
		DateTime,
		Hash,
	))
}
