package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// setupDirs points the global config dir at a temp dir and returns it
// alongside a temp project dir, writing the given TOML contents (skipped when
// empty) as the respective config.toml.
func setupDirs(t *testing.T, globalTOML, projectTOML string) (globalDir, projectDir string) {
	t.Helper()
	xdg := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", xdg)
	globalDir = filepath.Join(xdg, "ttanic")
	projectDir = t.TempDir()
	writeConfig(t, globalDir, globalTOML)
	writeConfig(t, filepath.Join(projectDir, ".ttanic"), projectTOML)
	return globalDir, projectDir
}

func writeConfig(t *testing.T, dir, content string) {
	t.Helper()
	if content == "" {
		return
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.toml"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func ptr[T any](v T) *T { return &v }

func TestLoad(t *testing.T) {
	tests := []struct {
		name        string
		global      string
		project     string
		standalone  bool // load with projectDir == ""
		ov          Overrides
		want        func(*Config) // mutations on top of Default()
		wantErr     error
		wantErrText string
	}{
		{
			name: "no files, defaults",
		},
		{
			name:   "global only",
			global: "[compression]\nlevel = \"best\"\nworkers = 4\n",
			want: func(c *Config) {
				c.Compression.Level = LevelBest
				c.Compression.Workers = 4
			},
		},
		{
			name:    "project overrides global",
			global:  "[compression]\nlevel = \"best\"\n",
			project: "[compression]\nlevel = \"fastest\"\n",
			want:    func(c *Config) { c.Compression.Level = LevelFastest },
		},
		{
			name:    "project sets only level, UI prefs and leader survive",
			global:  "[ui]\ntheme = \"dark\"\nshow_hidden = true\nsort = \"size\"\n\n[keys]\nleader = \",\"\n",
			project: "[compression]\nlevel = \"better\"\n",
			want: func(c *Config) {
				c.Compression.Level = LevelBetter
				c.UI.Theme = "dark"
				c.UI.ShowHidden = true
				c.UI.Sort = SortSize
				c.Keys.Leader = ","
			},
		},
		{
			name:    "defined zero values still override",
			global:  "[compression]\nworkers = 4\n\n[ui]\nshow_hidden = true\n",
			project: "[compression]\nworkers = 0\n\n[ui]\nshow_hidden = false\n",
			want:    func(_ *Config) {}, // back to the default zero values, explicitly
		},
		{
			name:       "standalone mode ignores project file",
			global:     "[compression]\nlevel = \"best\"\n",
			project:    "[compression]\nlevel = \"fastest\"\n",
			standalone: true,
			want:       func(c *Config) { c.Compression.Level = LevelBest },
		},
		{
			name:    "overrides beat both files",
			global:  "[compression]\nlevel = \"best\"\n",
			project: "[compression]\nlevel = \"fastest\"\n\n[ui]\nshow_hidden = true\n",
			ov: Overrides{
				Level:      ptr(LevelBetter),
				Workers:    ptr(8),
				OnSymlink:  ptr(SymlinkSkip),
				Sort:       ptr(SortMtime),
				ShowHidden: ptr(false),
			},
			want: func(c *Config) {
				c.Compression.Level = LevelBetter
				c.Compression.Workers = 8
				c.Archive.OnSymlink = SymlinkSkip
				c.UI.Sort = SortMtime
				c.UI.ShowHidden = false
			},
		},
		{
			name:   "every key merges from file",
			global: "[compression]\nlevel = \"fastest\"\nworkers = 2\n\n[archive]\non_symlink = \"skip\"\n\n[ui]\ntheme = \"dark\"\nshow_hidden = true\nsort = \"mtime\"\neditor = \"hx\"\nicons = \"ascii\"\n\n[keys]\nleader = \",\"\n",
			want: func(c *Config) {
				c.Compression.Level = LevelFastest
				c.Compression.Workers = 2
				c.Archive.OnSymlink = SymlinkSkip
				c.UI.Theme = "dark"
				c.UI.ShowHidden = true
				c.UI.Sort = SortMtime
				c.UI.Editor = "hx"
				c.UI.Icons = IconsASCII
				c.Keys.Leader = ","
			},
		},
		{
			name:    "bad level in file",
			project: "[compression]\nlevel = \"turbo\"\n",
			wantErr: ErrUnknownLevel,
		},
		{
			name:    "negative workers in file",
			project: "[compression]\nworkers = -1\n",
			wantErr: ErrInvalidWorkers,
		},
		{
			name:    "bad override value",
			ov:      Overrides{Level: ptr(Level("turbo"))},
			wantErr: ErrUnknownLevel,
		},
		{
			name:        "unknown key errors with the key name",
			project:     "[compression]\nworker = 4\n",
			wantErrText: "worker",
		},
		{
			name:        "malformed TOML errors with the file path",
			project:     "[compression\nlevel = \"best\"\n",
			wantErrText: "config.toml",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, projectDir := setupDirs(t, tt.global, tt.project)
			if tt.standalone {
				projectDir = ""
			}
			got, err := Load(projectDir, tt.ov)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("Load() error = %v, want errors.Is(..., %v)", err, tt.wantErr)
				}
				return
			}
			if tt.wantErrText != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErrText) {
					t.Fatalf("Load() error = %v, want it to contain %q", err, tt.wantErrText)
				}
				return
			}
			if err != nil {
				t.Fatalf("Load() unexpected error: %v", err)
			}
			want := Default()
			if tt.want != nil {
				tt.want(&want)
			}
			if got != want {
				t.Errorf("Load() = %+v, want %+v", got, want)
			}
		})
	}
}

func TestGlobalDir(t *testing.T) {
	t.Run("XDG_CONFIG_HOME set", func(t *testing.T) {
		t.Setenv("XDG_CONFIG_HOME", "/xdg/cfg")
		got, err := GlobalDir()
		if err != nil {
			t.Fatal(err)
		}
		if want := filepath.Join("/xdg/cfg", "ttanic"); got != want {
			t.Errorf("GlobalDir() = %q, want %q", got, want)
		}
	})
	t.Run("XDG_CONFIG_HOME unset falls back to ~/.config", func(t *testing.T) {
		t.Setenv("XDG_CONFIG_HOME", "")
		t.Setenv("HOME", "/home/crew")
		got, err := GlobalDir()
		if err != nil {
			t.Fatal(err)
		}
		if want := filepath.Join("/home/crew", ".config", "ttanic"); got != want {
			t.Errorf("GlobalDir() = %q, want %q", got, want)
		}
	})
}
