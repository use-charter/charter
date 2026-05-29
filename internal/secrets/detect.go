package secrets

import (
	"regexp"
	"strings"
)

var (
	envRefPattern      = regexp.MustCompile(`\$\{[A-Z0-9_]+\}|\$[A-Z0-9_]+`)
	placeholderPattern = regexp.MustCompile(`(?i)(^|[^A-Za-z0-9_-])your-api-key-here($|[^A-Za-z0-9_-])`)
	tokenPatterns      = []struct {
		reason string
		re     *regexp.Regexp
	}{
		{
			reason: "openai-style token",
			re:     regexp.MustCompile(`(^|[^A-Za-z0-9_-])(sk-[A-Za-z0-9_-]{20,})($|[^A-Za-z0-9_-])`),
		},
		{
			reason: "github personal access token",
			re:     regexp.MustCompile(`(^|[^A-Za-z0-9_])(ghp_[A-Za-z0-9]{30,})($|[^A-Za-z0-9_])`),
		},
		{
			reason: "aws access key id",
			re:     regexp.MustCompile(`(^|[^A-Za-z0-9])(AKIA[A-Z0-9]{16})($|[^A-Za-z0-9])`),
		},
		{
			reason: "slack bot token",
			re:     regexp.MustCompile(`(^|[^A-Za-z0-9_-])(xoxb-[A-Za-z0-9-]{20,})($|[^A-Za-z0-9_-])`),
		},
	}
)

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
	sanitized = placeholderPattern.ReplaceAllString(sanitized, "$1 $2")

	for _, pattern := range tokenPatterns {
		if groups := pattern.re.FindStringSubmatch(sanitized); len(groups) == 4 {
			return Match{Found: true, Reason: pattern.reason, Secret: groups[2], Prefix: prefixFrom(groups[2])}
		}
	}

	if strings.Contains(sanitized, "BEGIN RSA PRIVATE KEY") || strings.Contains(sanitized, "BEGIN EC PRIVATE KEY") || strings.Contains(sanitized, "BEGIN PRIVATE KEY") {
		return Match{Found: true, Reason: "private key header", Secret: "PRIVATE KEY", Prefix: "BEGIN"}
	}

	return Match{}
}

func prefixFrom(secret string) string {
	switch {
	case strings.HasPrefix(secret, "sk-"):
		return "sk-"
	case strings.HasPrefix(secret, "ghp_"):
		return "ghp_"
	case strings.HasPrefix(secret, "AKIA"):
		return "AKIA"
	case strings.HasPrefix(secret, "xoxb-"):
		return "xoxb-"
	default:
		return ""
	}
}
