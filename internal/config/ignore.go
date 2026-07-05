package config

import "strings"

// Rule is one parsed ignore pattern. Rules come from gitignore-syntax files
// (ParseIgnore) and are fused into a Matcher (NewMatcher). Pattern and Negate
// are exported for `ttanic ignore list`, which reports effective patterns and
// their source.
type Rule struct {
	Pattern string // the line as written in the ignore file
	Negate  bool   // "!" prefix: matching paths are re-included

	dirOnly bool     // trailing "/": matches directories only
	segs    []string // slash-split pattern; leading "**" when unanchored
}

// ParseIgnore parses one ignore file's content: gitignore syntax, one pattern
// per line. Blank lines and "#" comments are skipped, trailing unescaped
// spaces are trimmed. There is nothing to reject — every remaining line is a
// pattern — so no error is returned.
func ParseIgnore(src string) []Rule {
	var rules []Rule
	for _, line := range strings.Split(src, "\n") {
		if r, ok := parseLine(strings.TrimSuffix(line, "\r")); ok {
			rules = append(rules, r)
		}
	}
	return rules
}

func parseLine(line string) (Rule, bool) {
	if line == "" || strings.HasPrefix(line, "#") {
		return Rule{}, false
	}
	r := Rule{Pattern: line}
	pat := trimTrailingSpace(line)
	if strings.HasPrefix(pat, "!") {
		r.Negate = true
		pat = pat[1:]
	}
	if strings.HasSuffix(pat, "/") {
		r.dirOnly = true
		pat = strings.TrimRight(pat, "/")
	}
	if pat == "" {
		return Rule{}, false
	}
	// A separator at the start or middle anchors the pattern to the project
	// root; otherwise it matches at any depth, expressed as a leading "**".
	anchored := strings.Contains(pat, "/")
	pat = strings.TrimPrefix(pat, "/")
	r.segs = strings.Split(pat, "/")
	if !anchored {
		r.segs = append([]string{"**"}, r.segs...)
	}
	return r, true
}

// trimTrailingSpace strips trailing spaces unless they are backslash-escaped
// (a space preceded by an odd number of backslashes stays).
func trimTrailingSpace(s string) string {
	for strings.HasSuffix(s, " ") {
		backslashes := 0
		for i := len(s) - 2; i >= 0 && s[i] == '\\'; i-- {
			backslashes++
		}
		if backslashes%2 == 1 {
			break
		}
		s = s[:len(s)-1]
	}
	return s
}
