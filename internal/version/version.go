package version

import "runtime/debug"

// injected, commit, and date are overridden at release time via
// -ldflags "-X go.use-charter.dev/charter/internal/version.<name>=<value>"
// (ADR-0016 GoReleaser injects Version/FullCommit/CommitDate).
var (
	injected string
	commit   string
	date     string
)

// Version returns the Charter build version: the release value injected via
// ldflags when set, else the module version embedded by the Go toolchain (e.g.
// for `go install go.use-charter.dev/charter/cmd/charter@v1.2.3`), else a
// development placeholder.
func Version() string {
	if injected != "" {
		return injected
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		if v := info.Main.Version; v != "" && v != "(devel)" {
			return v
		}
	}
	return "0.0.0-dev"
}

// Commit returns the VCS revision: the ldflags-injected value, else the
// toolchain-embedded vcs.revision, else "unknown".
func Commit() string {
	if commit != "" {
		return commit
	}
	return buildSetting("vcs.revision")
}

// Date returns the build/commit timestamp: the ldflags-injected value, else the
// toolchain-embedded vcs.time, else "unknown".
func Date() string {
	if date != "" {
		return date
	}
	return buildSetting("vcs.time")
}

func buildSetting(key string) string {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, s := range info.Settings {
			if s.Key == key && s.Value != "" {
				return s.Value
			}
		}
	}
	return "unknown"
}
