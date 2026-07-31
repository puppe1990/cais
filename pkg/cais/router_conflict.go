package cais

import (
	"fmt"
	"strings"
)

// patternsConflict reports whether two path patterns would panic ServeMux registration
// under the same method (overlapping matches with neither more specific).
//
// Example conflict: /webhooks/{id}/delete vs /webhooks/kiwify/{token}
// Non-conflict:     /users/{id} vs /users/me (literal is more specific)
func patternsConflict(a, b string) bool {
	sa := pathSegments(a)
	sb := pathSegments(b)
	if len(sa) != len(sb) {
		return false
	}
	if len(sa) == 0 {
		return a == b
	}

	aMoreSpecific := true
	bMoreSpecific := true
	for i := range sa {
		wa, wb := isWildcardSeg(sa[i]), isWildcardSeg(sb[i])
		switch {
		case !wa && !wb:
			if sa[i] != sb[i] {
				return false
			}
		case wa && !wb:
			// b has a literal where a is wild — b is more specific on this segment
			aMoreSpecific = false
		case !wa && wb:
			bMoreSpecific = false
		default:
			// both wildcards — neither is more specific on this segment
			aMoreSpecific = false
			bMoreSpecific = false
		}
	}
	// ServeMux allows registration when one pattern is strictly more specific.
	if aMoreSpecific != bMoreSpecific {
		return false
	}
	return true
}

func pathSegments(pattern string) []string {
	pattern = strings.Trim(pattern, "/")
	if pattern == "" {
		return nil
	}
	return strings.Split(pattern, "/")
}

func isWildcardSeg(seg string) bool {
	return strings.HasPrefix(seg, "{") && strings.HasSuffix(seg, "}")
}

// formatRouteConflict rewrites a ServeMux registration panic with a short cais hint (#142).
func formatRouteConflict(registering string, rec any) string {
	return fmt.Sprintf(
		"cais: conflicting route %q: %v\n\n"+
			"Go 1.22+ ServeMux rejects overlapping patterns under the same method when neither is more specific\n"+
			"(e.g. POST /webhooks/{id}/delete vs POST /webhooks/kiwify/{token} both match /webhooks/kiwify/delete).\n"+
			"Prefer:\n"+
			"  - DELETE /webhooks/{id} instead of POST …/{id}/delete\n"+
			"  - a distinct prefix: /webhooks/incoming/{token} vs /webhooks/{id}\n"+
			"See AGENTS.md “Router path params and groups”.",
		registering, rec,
	)
}
