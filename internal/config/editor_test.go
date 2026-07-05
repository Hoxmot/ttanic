package config

import "testing"

func TestResolveEditor(t *testing.T) {
	tests := []struct {
		name          string
		editor        string
		visual, envEd string
		want          string
	}{
		{name: "config wins", editor: "hx", visual: "code -w", envEd: "vim", want: "hx"},
		{name: "VISUAL over EDITOR", editor: "", visual: "code -w", envEd: "vim", want: "code -w"},
		{name: "EDITOR as fallback", editor: "", visual: "", envEd: "vim", want: "vim"},
		{name: "nothing set", editor: "", visual: "", envEd: "", want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("VISUAL", tt.visual)
			t.Setenv("EDITOR", tt.envEd)
			cfg := Default()
			cfg.UI.Editor = tt.editor
			if got := cfg.ResolveEditor(); got != tt.want {
				t.Errorf("ResolveEditor() = %q, want %q", got, tt.want)
			}
		})
	}
}
