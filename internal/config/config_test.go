package config

import (
	"errors"
	"strings"
	"testing"
)

func TestDefaultIsValid(t *testing.T) {
	if err := Default().Validate(); err != nil {
		t.Fatalf("Default() must validate, got: %v", err)
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*Config)
		wantErr error
		wantIn  string // substring the error message must contain
	}{
		{
			name:    "unknown level",
			mutate:  func(c *Config) { c.Compression.Level = "turbo" },
			wantErr: ErrUnknownLevel,
			wantIn:  "turbo",
		},
		{
			name:    "negative workers",
			mutate:  func(c *Config) { c.Compression.Workers = -1 },
			wantErr: ErrInvalidWorkers,
			wantIn:  "-1",
		},
		{
			name:    "unknown symlink policy",
			mutate:  func(c *Config) { c.Archive.OnSymlink = "follow" },
			wantErr: ErrUnknownSymlinkPolicy,
			wantIn:  "follow",
		},
		{
			name:    "unknown sort",
			mutate:  func(c *Config) { c.UI.Sort = "kind" },
			wantErr: ErrUnknownSort,
			wantIn:  "kind",
		},
		{
			name:    "unknown icons",
			mutate:  func(c *Config) { c.UI.Icons = "emoji" },
			wantErr: ErrUnknownIcons,
			wantIn:  "emoji",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Default()
			tt.mutate(&cfg)
			err := cfg.Validate()
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Validate() = %v, want errors.Is(..., %v)", err, tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantIn) {
				t.Errorf("error %q does not name the bad value %q", err, tt.wantIn)
			}
		})
	}
}
