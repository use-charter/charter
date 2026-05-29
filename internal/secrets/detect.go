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

	sanitized := strings.TrimSpace(envRefPattern.ReplaceAllString(trimmed, " "))

	lower := strings.ToLower(sanitized)
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
		if idx := strings.Index(sanitized, pattern.prefix); idx >= 0 {
			secret := tokenFrom(sanitized[idx:])
			return Match{Found: true, Reason: pattern.reason, Secret: secret, Prefix: pattern.prefix}
		}
	}

	if strings.Contains(sanitized, "BEGIN RSA PRIVATE KEY") || strings.Contains(sanitized, "BEGIN EC PRIVATE KEY") || strings.Contains(sanitized, "BEGIN PRIVATE KEY") {
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
