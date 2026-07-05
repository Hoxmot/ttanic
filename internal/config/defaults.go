package config

// Default returns the built-in configuration, the base layer of every merge.
func Default() Config {
	return Config{
		Compression: Compression{
			Level:   LevelDefault,
			Workers: 0,
		},
		Archive: Archive{
			OnSymlink: SymlinkError,
		},
		UI: UI{
			Theme:      "default",
			ShowHidden: false,
			Sort:       SortName,
			Editor:     "",
			Icons:      IconsUnicode,
		},
		Keys: Keys{
			Leader: "space",
		},
	}
}
