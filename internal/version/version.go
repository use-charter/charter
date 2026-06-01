package version

import "runtime/debug"

// injected is overridden at release time via
// -ldflags "-X go.use-charter.dev/charter/internal/version.injected=v1.2.3".
var injected string

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
