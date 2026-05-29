package secrets

func RedactValue(secret string) string {
	if len(secret) <= 4 {
		return "[REDACTED]"
	}

	prefixLen := 4
	if len(secret) < prefixLen {
		prefixLen = len(secret)
	}

	return secret[:prefixLen] + "…"
}
