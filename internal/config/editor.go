package config

import "os"

// ResolveEditor returns the editor command to use: the config editor key,
// else $VISUAL, else $EDITOR. Empty means none is configured; the caller
// decides how to report that.
func (c Config) ResolveEditor() string {
	if c.UI.Editor != "" {
		return c.UI.Editor
	}
	if v := os.Getenv("VISUAL"); v != "" {
		return v
	}
	return os.Getenv("EDITOR")
}
