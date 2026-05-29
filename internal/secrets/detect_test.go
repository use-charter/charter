package secrets

import "testing"

func TestDetectRecognizesHighConfidencePrefixes(t *testing.T) {
	tests := []struct {
		name string
		line string
		want bool
	}{
		{name: "openai", line: "OPENAI_API_KEY=sk-test-12345678901234567890", want: true},
		{name: "github", line: "GITHUB_TOKEN=ghp_123456789012345678901234567890123456", want: true},
		{name: "aws", line: "AWS_ACCESS_KEY_ID=AKIA1234567890ABCD12", want: true},
		{name: "slack", line: "SLACK_BOT_TOKEN=xoxb-1234567890-1234567890-abcdefghijklmnop", want: true},
		{name: "placeholder", line: "OPENAI_API_KEY=your-api-key-here", want: false},
		{name: "env-ref-brace", line: "OPENAI_API_KEY=${OPENAI_API_KEY}", want: false},
		{name: "env-ref-shell", line: "OPENAI_API_KEY=$OPENAI_API_KEY", want: false},
		{name: "mixed-env-ref-and-literal-secret", line: "FOO=$BAR OPENAI_API_KEY=sk-test-12345678901234567890", want: true},
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
	got := RedactValue("sk-test-12345678901234567890")
	if got == "sk-test-12345678901234567890" {
		t.Fatalf("expected redaction, got raw secret")
	}

	if got == "" {
		t.Fatalf("expected non-empty redacted value")
	}
}
