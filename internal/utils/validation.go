package utils

import (
	"regexp"
	"strings"
)

// Username validation regex: only lowercase a-z, digits 0-9, dots, hyphens, and at signs
var usernameRegex = regexp.MustCompile(`^[a-z0-9.@-]+$`)

// IsValidUsername checks if username contains only lowercase letters, digits, dots, and hyphens
func IsValidUsername(username string) bool {
	return usernameRegex.MatchString(username)
}

var usernameDisallowedChars = regexp.MustCompile(`[^a-z0-9.@-]+`)
var usernameRepeatedHyphens = regexp.MustCompile(`-{2,}`)

// NormalizeUsername maps an external identity's username (which may contain
// uppercase letters, spaces, or other characters IsValidUsername rejects --
// Home Assistant's own usernames aren't restricted the way donetick's are)
// into donetick's username format: lowercased, disallowed characters
// collapsed to a single hyphen, and leading/trailing hyphens trimmed. The
// result still isn't guaranteed valid (e.g. an all-symbol input normalizes
// to ""), so callers must still check it with IsValidUsername.
func NormalizeUsername(username string) string {
	lowered := strings.ToLower(strings.TrimSpace(username))
	collapsed := usernameDisallowedChars.ReplaceAllString(lowered, "-")
	collapsed = usernameRepeatedHyphens.ReplaceAllString(collapsed, "-")
	return strings.Trim(collapsed, "-")
}
