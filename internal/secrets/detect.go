package secrets

import (
	"regexp"
	"strings"
)

var envRefPattern = regexp.MustCompile(`\$\{[A-Z0-9_]+\}|\$[A-Z0-9_]+`)

type Match struct {
	Found  bool
	Reason string
	Secret string
	Prefix string
}

func DetectLine(line string) Match {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return Match{}
	}

	if envRefPattern.MatchString(trimmed) {
		return Match{}
	}

	lower := strings.ToLower(trimmed)
	if strings.Contains(lower, "your-api-key-here") {
		return Match{}
	}

	patterns := []struct {
		prefix string
		reason string
	}{
		{prefix: "sk-", reason: "openai-style token"},
		{prefix: "ghp_", reason: "github personal access token"},
		{prefix: "AKIA", reason: "aws access key id"},
		{prefix: "xoxb-", reason: "slack bot token"},
	}

	for _, pattern := range patterns {
		if idx := strings.Index(trimmed, pattern.prefix); idx >= 0 {
			secret := tokenFrom(trimmed[idx:])
			return Match{Found: true, Reason: pattern.reason, Secret: secret, Prefix: pattern.prefix}
		}
	}

	if strings.Contains(trimmed, "BEGIN RSA PRIVATE KEY") || strings.Contains(trimmed, "BEGIN EC PRIVATE KEY") || strings.Contains(trimmed, "BEGIN PRIVATE KEY") {
		return Match{Found: true, Reason: "private key header", Secret: "PRIVATE KEY", Prefix: "BEGIN"}
	}

	return Match{}
}

func tokenFrom(s string) string {
	fields := strings.FieldsFunc(s, func(r rune) bool {
		switch r {
		case ' ', '\t', '\r', '\n', '\'', '"', ',', ')', ']', '}':
			return true
		default:
			return false
		}
	})
	if len(fields) == 0 {
		return s
	}
	return fields[0]
}
