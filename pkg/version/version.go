package version

import (
	"fmt"
	"runtime"
	"runtime/debug"
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
	Version = "v0.0.0-dev"
	Info    VersionInfo
)

func init() {
	Info = VersionInfo{
		BuildDate: readBuildInfo("vcs.time"),
		Commit:    readBuildInfo("vcs.revision"),
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

func readBuildInfo(key string) string {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == key {
				return setting.Value
			}
		}
	}

	return ""
}
