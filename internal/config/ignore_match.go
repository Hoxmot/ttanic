package config

import (
	"path"
	"strings"
)

// Matcher reports whether paths are ignored. Rules are evaluated in order and
// the last match wins — that is what lets a later "!" rule re-include a path
// an earlier rule ignored, both within one file and across layers.
type Matcher struct {
	rules []Rule
}

// NewMatcher fuses rule layers into a matcher. Layers are concatenated in
// argument order, so a later layer (the project ignore file) can negate an
// earlier one (the global file). A matcher with no rules matches nothing.
func NewMatcher(layers ...[]Rule) *Matcher {
	m := &Matcher{}
	for _, l := range layers {
		m.rules = append(m.rules, l...)
	}
	return m
}

// Match reports whether relPath — slash-separated, relative to the project
// root — is ignored. isDir gates directory-only rules like "build/"; walkers
// get it from fs.DirEntry.IsDir.
//
// Match judges the path itself, not its ancestors: it does not check whether
// a parent directory is ignored. Walkers must prune ignored directories
// top-down, which also yields git's "cannot re-include below an excluded
// directory" behavior.
func (m *Matcher) Match(relPath string, isDir bool) bool {
	p := strings.TrimPrefix(path.Clean("/"+relPath), "/")
	if p == "" {
		return false
	}
	segs := strings.Split(p, "/")
	ignored := false
	for _, r := range m.rules {
		if r.dirOnly && !isDir {
			continue
		}
		if matchSegments(r.segs, segs) {
			ignored = !r.Negate
		}
	}
	return ignored
}

// matchSegments matches a slash-split pattern against path segments. A "**"
// segment matches zero or more directories, except in trailing position where
// it matches everything *inside* the prefix but not the prefix itself
// ("abc/**" matches "abc/x" and deeper, never "abc").
func matchSegments(pat, segs []string) bool {
	for len(pat) > 0 {
		if pat[0] == "**" {
			if len(pat) == 1 {
				return len(segs) > 0
			}
			for skip := 0; skip <= len(segs); skip++ {
				if matchSegments(pat[1:], segs[skip:]) {
					return true
				}
			}
			return false
		}
		if len(segs) == 0 || !matchSegment(pat[0], segs[0]) {
			return false
		}
		pat, segs = pat[1:], segs[1:]
	}
	return len(segs) == 0
}

// matchSegment matches one path segment against one pattern segment: "*" any
// run of bytes, "?" one byte, "[...]" a character class, "\x" a literal x.
// None of them can match "/" because segments contain none. Matching is
// byte-wise and case-sensitive.
func matchSegment(pat, name string) bool {
	px, nx := 0, 0
	starPx, starNx := -1, 0
	for nx < len(name) {
		if px < len(pat) {
			switch c := pat[px]; c {
			case '*':
				starPx, starNx = px, nx
				px++
				continue
			case '?':
				px++
				nx++
				continue
			case '[':
				matched, next, ok := matchClass(pat, px, name[nx])
				if ok && matched {
					px = next
					nx++
					continue
				}
				if !ok && name[nx] == '[' { // unterminated class: literal "["
					px++
					nx++
					continue
				}
			case '\\':
				if px+1 < len(pat) {
					if pat[px+1] == name[nx] {
						px += 2
						nx++
						continue
					}
				} else if name[nx] == '\\' { // trailing backslash: literal
					px++
					nx++
					continue
				}
			default:
				if c == name[nx] {
					px++
					nx++
					continue
				}
			}
		}
		// Mismatch: backtrack to the last "*", letting it eat one more byte.
		if starPx < 0 {
			return false
		}
		starNx++
		px, nx = starPx+1, starNx
	}
	for px < len(pat) && pat[px] == '*' {
		px++
	}
	return px == len(pat)
}

// matchClass matches c against the "[...]" class starting at pat[start].
// next is the index just past the closing "]". ok is false when the class is
// unterminated, in which case the caller treats "[" as a literal byte.
func matchClass(pat string, start int, c byte) (matched bool, next int, ok bool) {
	i := start + 1
	negate := false
	if i < len(pat) && (pat[i] == '!' || pat[i] == '^') {
		negate = true
		i++
	}
	first := true
	for {
		if i >= len(pat) {
			return false, 0, false
		}
		if pat[i] == ']' && !first {
			return matched != negate, i + 1, true
		}
		first = false
		if pat[i] == '\\' && i+1 < len(pat) {
			i++
		}
		lo := pat[i]
		i++
		hi := lo
		if i+1 < len(pat) && pat[i] == '-' && pat[i+1] != ']' {
			i++
			if pat[i] == '\\' && i+1 < len(pat) {
				i++
			}
			hi = pat[i]
			i++
		}
		if lo <= c && c <= hi {
			matched = true
		}
	}
}
