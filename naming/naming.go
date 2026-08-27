// Package naming defines the session-name contract shared by every service
// that assembles or validates org/repo/bookmark triples.
//
// Session names are derived from the workspace triple as
// `org:repo:bookmark` (e.g. `acme:my.repo:feature-x`). Components must never
// contain ':' themselves, which keeps the derivation a strict bijection.
// Components MAY contain dots (`my.repo`, `v1.2`) — jj/git ref rules only
// forbid `..`, trailing `.`, and `.lock`, all checked below.
//
// The separator is chosen for readability only: the session name is never
// embedded in NATS subjects or KV keys (the agent hashes it into a safe
// token there and always carries the real name in payloads), and ':' is
// legal in URL path segments unencoded.
package naming

import (
	"regexp"
	"strings"
)

var componentRe = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,127}$`)

// ValidComponent reports whether s is acceptable as an org/repo/bookmark name.
func ValidComponent(s string) bool {
	if !componentRe.MatchString(s) {
		return false
	}
	if strings.Contains(s, ":") {
		return false // reserved as the separator (defense in depth)
	}
	if strings.Contains(s, "..") {
		return false // path traversal / jj rule
	}
	if strings.HasSuffix(s, ".") || strings.HasSuffix(s, ".lock") {
		return false // jj/git ref rules
	}
	return true
}

// Session derives the session name for a workspace triple.
func Session(org, repo, bookmark string) string {
	return org + ":" + repo + ":" + bookmark
}

// Parse splits a derived session name back into its triple. ok is false when
// the name does not have exactly three ':'-separated components.
func Parse(name string) (org, repo, bookmark string, ok bool) {
	parts := strings.Split(name, ":")
	if len(parts) != 3 {
		return "", "", "", false
	}
	return parts[0], parts[1], parts[2], true
}

// Require validates all named components or returns the first offender.
func Require(pairs ...[2]string) (string, bool) {
	for _, p := range pairs {
		if !ValidComponent(p[1]) {
			return p[0], false
		}
	}
	return "", true
}
