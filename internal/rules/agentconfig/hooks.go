// Package agentconfig implements the AE-CC agent-config rules: it scans JSON
// hook configuration files for dangerous shell commands (AE-CC-001) and checks
// the agent context for an explicit edit-scope declaration (AE-CC-002).
package agentconfig

import (
	"encoding/json"
	"fmt"
	"strings"
)

// HookCommand is a single command (or joined args) extracted from a hook config.
type HookCommand struct {
	Value string
	Line  int // 1-based best-effort source line; 0 if unknown
}

// ConfigFile is one parsed JSON hook config file.
type ConfigFile struct {
	Path     string
	Commands []HookCommand
}

func parseHookConfig(path string, data []byte) (ConfigFile, error) {
	var root map[string]any
	if err := json.Unmarshal(data, &root); err != nil {
		return ConfigFile{}, fmt.Errorf("parse %s: %w", path, err)
	}

	cf := ConfigFile{Path: path}
	hooks, ok := root["hooks"]
	if !ok {
		return cf, nil
	}

	var raw []string
	collectCommands(hooks, &raw)
	text := string(data)
	for _, v := range raw {
		cf.Commands = append(cf.Commands, HookCommand{Value: v, Line: bestEffortLine(text, v)})
	}
	return cf, nil
}

// collectCommands walks an arbitrary JSON subtree and collects every "command"
// string and joined "args" string slice, robust to Claude (nested) and Cursor
// (flat) hook shapes. The trailing range over the map re-visits the already-read
// "command"/"args" values, but those are strings or string slices that fall
// through the type switch as no-ops, so no command is double-counted. Commands
// live inside JSON arrays (order-preserving); only sibling map keys are visited
// in Go's randomized map order, which does not affect the collected order for
// the array-wrapped Claude/Cursor schemas.
func collectCommands(node any, out *[]string) {
	switch n := node.(type) {
	case map[string]any:
		if c, ok := n["command"].(string); ok && strings.TrimSpace(c) != "" {
			*out = append(*out, c)
		}
		if a, ok := n["args"].([]any); ok {
			parts := make([]string, 0, len(a))
			for _, e := range a {
				if s, ok := e.(string); ok {
					parts = append(parts, s)
				}
			}
			if len(parts) > 0 {
				*out = append(*out, strings.Join(parts, " "))
			}
		}
		for _, v := range n {
			collectCommands(v, out)
		}
	case []any:
		for _, e := range n {
			collectCommands(e, out)
		}
	}
}

// bestEffortLine returns the 1-based line of the first occurrence of value in
// text; 0 if not found. Diagnostic only.
func bestEffortLine(text, value string) int {
	if value == "" {
		return 0
	}
	for i, line := range strings.Split(text, "\n") {
		if strings.Contains(line, value) {
			return i + 1
		}
	}
	return 0
}
