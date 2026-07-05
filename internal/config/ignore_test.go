package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseIgnore(t *testing.T) {
	rules := ParseIgnore("# comment\n\nfoo\n!bar\nbuild/\n   \n\\#literal\n")
	var patterns []string
	for _, r := range rules {
		patterns = append(patterns, r.Pattern)
	}
	want := []string{"foo", "!bar", "build/", "\\#literal"}
	if strings.Join(patterns, ",") != strings.Join(want, ",") {
		t.Errorf("ParseIgnore() patterns = %v, want %v", patterns, want)
	}
	if rules[0].Negate || !rules[1].Negate {
		t.Errorf("Negate flags wrong: foo=%v !bar=%v", rules[0].Negate, rules[1].Negate)
	}
}

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

		// escapes and trailing spaces
		{"escaped hash", `\#special`, "#special", false, true},
		{"escaped bang", `\!important`, "!important", false, true},
		{"trailing spaces trimmed", "foo   ", "foo", false, true},
		{"escaped trailing space kept", `foo\ `, "foo ", false, true},

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
	global := ParseIgnore("*.log\ntmp/\n")
	project := ParseIgnore("!keep.log\n*.bak\n")
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

// writeIgnore writes content as dir/ignore, creating dir. Empty content is
// skipped, mirroring writeConfig.
func writeIgnore(t *testing.T, dir, content string) {
	t.Helper()
	if content == "" {
		return
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "ignore"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestLoadIgnore(t *testing.T) {
	tests := []struct {
		name       string
		global     string
		project    string
		standalone bool // load with projectDir == ""
		path       string
		isDir      bool
		want       bool
	}{
		{
			name: "both files absent, matches nothing",
			path: "anything.log",
		},
		{
			name:   "global only",
			global: "*.log\n",
			path:   "a.log",
			want:   true,
		},
		{
			name:    "project only",
			project: "build/\n",
			path:    "build",
			isDir:   true,
			want:    true,
		},
		{
			name:    "project negates global",
			global:  "*.log\n",
			project: "!keep.log\n",
			path:    "keep.log",
			want:    false,
		},
		{
			name:       "standalone mode skips project file",
			project:    "*.log\n",
			standalone: true,
			path:       "a.log",
			want:       false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			xdg := t.TempDir()
			t.Setenv("XDG_CONFIG_HOME", xdg)
			projectDir := t.TempDir()
			writeIgnore(t, filepath.Join(xdg, "ttanic"), tt.global)
			writeIgnore(t, filepath.Join(projectDir, ".ttanic"), tt.project)
			if tt.standalone {
				projectDir = ""
			}
			m, err := LoadIgnore(projectDir)
			if err != nil {
				t.Fatalf("LoadIgnore() unexpected error: %v", err)
			}
			if got := m.Match(tt.path, tt.isDir); got != tt.want {
				t.Errorf("Match(%q, %v) = %v, want %v", tt.path, tt.isDir, got, tt.want)
			}
		})
	}

	t.Run("unreadable file errors with its path", func(t *testing.T) {
		xdg := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", xdg)
		// A directory named "ignore" makes ReadFile fail with something
		// other than fs.ErrNotExist.
		if err := os.MkdirAll(filepath.Join(xdg, "ttanic", "ignore"), 0o755); err != nil {
			t.Fatal(err)
		}
		_, err := LoadIgnore("")
		if err == nil || !strings.Contains(err.Error(), "ignore") {
			t.Fatalf("LoadIgnore() error = %v, want it to name the ignore file", err)
		}
	})
}
