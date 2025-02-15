package version

import (
	"fmt"
	"runtime"
	"strings"
)

// VersionInfo exposes information about saturn-bot and the Go runtime.
type VersionInfo struct {
	BuildDate string `json:"buildDate"`
	Commit    string `json:"commit"`
	GoArch    string `json:"goArch"`
	GoOS      string `json:"goOS"`
	GoVersion string `json:"goVersion"`
	Version   string `json:"version"`
}

// String implements [fmt.Stringer].
func (vi VersionInfo) String() string {
	return strings.TrimSpace(fmt.Sprintf(
		"%s %s %s/%s %s %s",
		vi.Version,
		vi.GoVersion,
		vi.GoOS,
		vi.GoArch,
		vi.BuildDate,
		vi.Commit,
	))
}

var (
	Version  = "v0.0.0-dev"
	Hash     = ""
	DateTime = ""
	Info     VersionInfo
)

func init() {
	Info = VersionInfo{
		BuildDate: DateTime,
		Commit:    Hash,
		GoArch:    runtime.GOARCH,
		GoOS:      runtime.GOOS,
		GoVersion: runtime.Version(),
		Version:   Version,
	}
}

// String returns the full version string.
func String() string {
	return Info.String()
}
