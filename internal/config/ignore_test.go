package config

import (
	"strings"
	"testing"
)

func TestParseIgnore(t *testing.T) {
	rules := ParseIgnore(`# comment

foo
!bar
build/

\#literal
`)
	var patterns []string
	for _, r := range rules {
		patterns = append(patterns, r.Pattern)
	}
	want := []string{"foo", "!bar", "build/", `\#literal`}
	if strings.Join(patterns, ",") != strings.Join(want, ",") {
		t.Errorf("ParseIgnore() patterns = %v, want %v", patterns, want)
	}
	if rules[0].Negate || !rules[1].Negate {
		t.Errorf("Negate flags wrong: foo=%v !bar=%v", rules[0].Negate, rules[1].Negate)
	}
}

func TestTrimTrailingSpace(t *testing.T) {
	tests := []struct{ in, want string }{
		{"foo", "foo"},
		{"foo   ", "foo"},
		{`foo\ `, `foo\ `},     // odd backslashes: the space is escaped
		{`foo\\ `, `foo\\`},    // even: the backslashes escape each other
		{`foo\\\ `, `foo\\\ `}, // odd again: last backslash escapes the space
		{`foo\  `, `foo\ `},    // trims up to the escaped space, keeps it
	}
	for _, tt := range tests {
		if got := trimTrailingSpace(tt.in); got != tt.want {
			t.Errorf("trimTrailingSpace(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
