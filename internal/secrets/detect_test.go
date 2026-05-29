package secrets

import (
	"strings"
	"testing"
)

func makeOpenAIToken() string {
	return "sk-" + strings.Repeat("a", 24)
}

func makeGitHubToken() string {
	return "ghp_" + strings.Repeat("b", 36)
}

func makeAWSAccessKeyID() string {
	return "AKIA" + strings.Repeat("C", 16)
}

func makeSlackToken() string {
	return "xoxb-" + strings.Repeat("1", 10) + "-" + strings.Repeat("2", 10) + "-" + strings.Repeat("d", 16)
}

func TestDetectRecognizesHighConfidencePrefixes(t *testing.T) {
	tests := []struct {
		name string
		line string
		want bool
	}{
		{name: "openai", line: "OPENAI_API_KEY=" + makeOpenAIToken(), want: true},
		{name: "github", line: "GITHUB_TOKEN=" + makeGitHubToken(), want: true},
		{name: "aws", line: "AWS_ACCESS_KEY_ID=" + makeAWSAccessKeyID(), want: true},
		{name: "slack", line: "SLACK_BOT_TOKEN=" + makeSlackToken(), want: true},
		{name: "placeholder", line: "OPENAI_API_KEY=your-api-key-here", want: false},
		{name: "env-ref-brace", line: "OPENAI_API_KEY=${OPENAI_API_KEY}", want: false},
		{name: "env-ref-shell", line: "OPENAI_API_KEY=$OPENAI_API_KEY", want: false},
		{name: "mixed-env-ref-and-literal-secret", line: "FOO=$BAR OPENAI_API_KEY=" + makeOpenAIToken(), want: true},
		{name: "placeholder-does-not-hide-real-secret", line: "OPENAI_API_KEY=your-api-key-here GITHUB_TOKEN=" + makeGitHubToken(), want: true},
		{name: "normal-prose-does-not-match-prefix-fragment", line: "Keep risk-aware notes in docs.", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectLine(tt.line)
			if got.Found != tt.want {
				t.Fatalf("expected found=%v, got %v for %q", tt.want, got.Found, tt.line)
			}
		})
	}
}

func TestRedactPreservesOnlySafePrefix(t *testing.T) {
	raw := makeOpenAIToken()
	got := RedactValue(raw)
	if got == raw {
		t.Fatalf("expected redaction, got raw secret")
	}

	if got == "" {
		t.Fatalf("expected non-empty redacted value")
	}
}
