// +build vita

package filepath

import "strings"

// IsAbs reports whether the path is absolute.
func IsAbs(path string) bool {
	l := volumeNameLen(path)
	if l == 0 {
		return false
	}
	path = path[l:]
	if path == "" {
		return false
	}
	// the spash is optional after the partition name
	return true
}

// volumeNameLen returns length of the leading volume name on vita, e.g.: "ux0:" return 4
// It returns 0 elsewhere.
func volumeNameLen(path string) int {
	i := strings.Index(path, ":")
	if i < 1 {
		return 0
	}

	return i + 1
}

// HasPrefix exists for historical compatibility and should not be used.
//
// Deprecated: HasPrefix does not respect path boundaries and
// does not ignore case when required.
func HasPrefix(p, prefix string) bool {
	return strings.HasPrefix(p, prefix)
}

func splitList(path string) []string {
	if path == "" {
		return []string{}
	}
	return strings.Split(path, string(ListSeparator))
}

func abs(path string) (string, error) {
	return unixAbs(path)
}

func join(elem []string) string {
	// If there's a bug here, fix the logic in ./path_plan9.go too.
	for i, e := range elem {
		if e != "" {
			return Clean(strings.Join(elem[i:], string(Separator)))
		}
	}
	return ""
}

func sameWord(a, b string) bool {
	return a == b
}
