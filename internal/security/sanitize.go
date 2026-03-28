package security

import (
	"fmt"
	"regexp"
)

var (
	reSafeName   = regexp.MustCompile(`^[a-zA-Z0-9_\-]+$`)
	reSafeDomain = regexp.MustCompile(`^[a-zA-Z0-9._\-]+$`)
	reSafePath   = regexp.MustCompile(`^[a-zA-Z0-9/_\-\.]+$`)
)

// SafeName validates a project/DB/user name (alphanumeric, dash, underscore only).
func SafeName(s string) error {
	if !reSafeName.MatchString(s) {
		return fmt.Errorf("unsafe name %q: only alphanumeric, dash, and underscore allowed", s)
	}
	return nil
}

// SafeDomain validates a domain name.
func SafeDomain(s string) error {
	if !reSafeDomain.MatchString(s) {
		return fmt.Errorf("unsafe domain %q", s)
	}
	return nil
}

// SafePath validates an absolute unix path used in shell commands.
func SafePath(s string) error {
	if len(s) == 0 || s[0] != '/' {
		return fmt.Errorf("path must be absolute: %q", s)
	}
	if !reSafePath.MatchString(s) {
		return fmt.Errorf("unsafe path %q: unexpected characters", s)
	}
	return nil
}
