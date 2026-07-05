package config

import "testing"

func TestMatch(t *testing.T) {
	tests := []struct {
		name     string
		patterns string // one ignore file's content
		path     string
		isDir    bool
		want     bool
	}{
		// plain names: unanchored, any depth
		{"name at root", "foo", "foo", false, true},
		{"name nested", "foo", "a/b/foo", false, true},
		{"name is not a substring", "foo", "foobar", false, false},
		{"name matches dirs too", "foo", "a/foo", true, true},

		// * ? and character classes
		{"star suffix", "*.log", "a.log", false, true},
		{"star suffix nested", "*.log", "sub/b.log", false, true},
		{"star matches a run of chars", "a*e", "abcde", false, true},
		{"star matches zero chars", "a*e", "ae", false, true},
		{"star does not cross slash", "a*e", "a/e", false, false},
		{"question mark", "?at", "cat", false, true},
		{"question mark wrong length", "?at", "flat", false, false},
		{"class", "*.[oa]", "f.o", false, true},
		{"class second member", "*.[oa]", "f.a", false, true},
		{"class miss", "*.[oa]", "f.c", false, false},
		{"class range", "v[0-9]", "v5", false, true},
		{"negated class", "[!a]bc", "xbc", false, true},
		{"negated class miss", "[!a]bc", "abc", false, false},
		{"unterminated class is literal", "a[b", "a[b", false, true},

		// dir-only patterns
		{"dir-only matches dir", "build/", "build", true, true},
		{"dir-only skips file", "build/", "build", false, false},
		{"dir-only nested", "build/", "sub/build", true, true},

		// anchoring
		{"leading slash anchors", "/dist", "dist", false, true},
		{"leading slash not nested", "/dist", "sub/dist", false, false},
		{"middle slash anchors", "doc/frotz", "doc/frotz", true, true},
		{"middle slash not nested", "doc/frotz", "a/doc/frotz", true, false},
		{"trailing slash does not anchor", "frotz/", "a/frotz", true, true},

		// ** forms
		{"leading doublestar at root", "**/foo", "foo", false, true},
		{"leading doublestar nested", "**/foo", "a/b/foo", false, true},
		{"leading doublestar with tail", "**/foo/bar", "x/foo/bar", false, true},
		{"trailing doublestar contents", "abc/**", "abc/x", false, true},
		{"trailing doublestar deep", "abc/**", "abc/x/y", false, true},
		{"trailing doublestar not prefix itself", "abc/**", "abc", true, false},
		{"middle doublestar zero dirs", "a/**/b", "a/b", false, true},
		{"middle doublestar one dir", "a/**/b", "a/x/b", false, true},
		{"middle doublestar many dirs", "a/**/b", "a/x/y/b", false, true},
		{"middle doublestar needs tail", "a/**/b", "a/x/c", false, false},
		{"bare doublestar matches everything", "**", "any/thing", false, true},
		{"doublestar inside segment acts as star", "a**b", "aXXb", false, true},
		{"consecutive doublestars all match zero", "**/**/z", "z", false, true},
		{"consecutive doublestars nested", "**/**/z", "a/b/z", false, true},

		// escapes and trailing spaces
		{"escaped hash", `\#special`, "#special", false, true},
		{"escaped bang", `\!important`, "!important", false, true},
		{"trailing spaces trimmed", "foo   ", "foo", false, true},
		{"trimmed pattern misses name with space", "foo   ", "foo ", false, false},
		{"escaped trailing space kept", `foo\ `, "foo ", false, true},
		{"even backslashes trim the space", `foo\\ `, `foo\`, false, true},

		// negation, last match wins
		{"negation re-includes", "*.log\n!keep.log", "keep.log", false, false},
		{"negation leaves others", "*.log\n!keep.log", "other.log", false, true},
		{"negation before ignore loses", "!keep.log\n*.log", "keep.log", false, true},
		{"dir-only negation skips files", "keep*\n!keep/", "keep", false, true},

		// input normalization
		{"dot-slash prefix", "foo", "./foo", false, true},
		{"root never matches", "**", ".", true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMatcher(ParseIgnore(tt.patterns))
			if got := m.Match(tt.path, tt.isDir); got != tt.want {
				t.Errorf("Match(%q, %v) with %q = %v, want %v",
					tt.path, tt.isDir, tt.patterns, got, tt.want)
			}
		})
	}
}

func TestMatchLayers(t *testing.T) {
	global := ParseIgnore(`*.log
tmp/
`)
	project := ParseIgnore(`!keep.log
*.bak
`)
	m := NewMatcher(global, project)

	tests := []struct {
		path  string
		isDir bool
		want  bool
	}{
		{"debug.log", false, true},  // global rule
		{"keep.log", false, false},  // project negates global
		{"tmp", true, true},         // global dir-only survives layering
		{"old.bak", false, true},    // project appends
		{"README.md", false, false}, // matched by nothing
	}
	for _, tt := range tests {
		if got := m.Match(tt.path, tt.isDir); got != tt.want {
			t.Errorf("Match(%q, %v) = %v, want %v", tt.path, tt.isDir, got, tt.want)
		}
	}
}

func TestMatcherNoRules(t *testing.T) {
	m := NewMatcher()
	for _, p := range []string{"foo", "a/b/c", ".ttanic", "x.log"} {
		if m.Match(p, false) || m.Match(p, true) {
			t.Errorf("empty matcher ignored %q", p)
		}
	}
}
