package config

import "strings"

// IsPlaceholder reports whether a config value is an unset cloud/production placeholder.
// Local values such as localhost or 127.0.0.1 are never placeholders.
func IsPlaceholder(v string) bool {
	s := strings.ToLower(strings.TrimSpace(v))
	if s == "" {
		return true
	}
	for _, m := range []string{
		"placeholder",
		"rm-xxx",
		"redis-placeholder",
		"oss-placeholder",
		"mq-placeholder",
		"example.invalid",
	} {
		if strings.Contains(s, m) {
			return true
		}
	}
	return false
}
