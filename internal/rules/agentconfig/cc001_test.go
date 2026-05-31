package agentconfig

import (
	"strings"
	"testing"

	"go.use-charter.dev/charter/internal/findings"
)

func TestIsDangerousCommand(t *testing.T) {
	cases := []struct {
		cmd  string
		want bool
	}{
		{"rm -rf ./build", true},
		{"sudo chmod 777 ./bin", true},
		{"git reset --hard origin/main", true},
		{"chown -R root /app", true},
		{"dd if=/dev/zero of=/dev/sda", true},
		{`"$CLAUDE_PROJECT_DIR"/.claude/hooks/format.sh`, false},
		{"cd app && npm test", false},
		{"jq -r '.tool_input.file_path'", false},
		{"prettier --write", false},
		{"git add .", false},
		{"git add -A", false},
		{"echo untruncated output", false},
		{"printf truncated_result", false},
		{"truncate -s 0 /var/log/app.log", true},
	}
	for _, c := range cases {
		if got := isDangerousCommand(c.cmd); got != c.want {
			t.Errorf("isDangerousCommand(%q) = %v, want %v", c.cmd, got, c.want)
		}
	}
}

func TestCheckDangerousCommandsFinding(t *testing.T) {
	files := []ConfigFile{{Path: ".claude/settings.json", Commands: []HookCommand{
		{Value: "rm -rf ./build", Line: 7},
		{Value: "prettier --write", Line: 9},
	}}}
	fs := checkDangerousCommands(files)
	if len(fs) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(fs))
	}
	f := fs[0]
	if f.RuleID != "AE-CC-001" || f.Severity != findings.SeverityBlocker {
		t.Fatalf("wrong finding: %+v", f)
	}
	if len(f.Locations) != 1 || f.Locations[0].Path != ".claude/settings.json" || f.Locations[0].Line != 7 {
		t.Fatalf("wrong location: %+v", f.Locations)
	}
	if !strings.Contains(strings.Join(f.Evidence, " "), "rm -rf") {
		t.Fatalf("evidence missing offending command: %v", f.Evidence)
	}
}
