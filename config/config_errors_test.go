package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/BurntSushi/toml"
)

func TestLoadConfigFileErrors(t *testing.T) {
	for _, scenario := range []struct {
		name, fixture string
		explicit      bool
		envDSN        string
	}{
		{name: "missing default", fixture: "missing"},
		{name: "missing explicit", fixture: "missing", explicit: true},
		{name: "missing explicit with DSN override", fixture: "missing", explicit: true, envDSN: ":memory:"},
		{name: "valid default", fixture: "valid"},
		{name: "valid explicit", fixture: "valid", explicit: true},
		{name: "malformed default", fixture: "malformed"},
		{name: "malformed explicit", fixture: "malformed", explicit: true},
		{name: "directory default", fixture: "directory"},
		{name: "directory explicit", fixture: "directory", explicit: true},
		{name: "invalid parent default", fixture: "invalid parent"},
		{name: "invalid parent explicit", fixture: "invalid parent", explicit: true},
		{name: "unreadable default", fixture: "unreadable"},
		{name: "unreadable explicit", fixture: "unreadable", explicit: true},
		{name: "inaccessible parent default", fixture: "inaccessible parent"},
		{name: "inaccessible parent explicit", fixture: "inaccessible parent", explicit: true},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			root := t.TempDir()
			configHome := filepath.Join(root, "config")
			parent := filepath.Join(configHome, "troveler")
			path := filepath.Join(parent, "config.toml")
			t.Setenv("HOME", filepath.Join(root, "home"))
			t.Setenv("XDG_CONFIG_HOME", configHome)
			t.Setenv("XDG_DATA_HOME", filepath.Join(root, "data"))
			t.Setenv("TROVELER_DSN", scenario.envDSN)
			switch scenario.fixture {
			case "valid", "unreadable", "inaccessible parent":
				writeXDGConfigForTest(t, path, "dsn = ':memory:'\n[search]\ntagline_width = 79\n")
				if scenario.fixture == "unreadable" || scenario.fixture == "inaccessible parent" {
					if os.Geteuid() == 0 {
						t.Skip("root bypasses fixture permission restrictions")
					}
					blocked := path
					if scenario.fixture == "inaccessible parent" {
						blocked = parent
					}
					if err := os.Chmod(blocked, 0000); err != nil {
						t.Fatal(err)
					}
					t.Cleanup(func() { _ = os.Chmod(blocked, 0700) })
				}
			case "malformed":
				writeXDGConfigForTest(t, path, "dsn = ':memory:'\n[search]\ntagline_width = [\n")
			case "directory":
				if err := os.MkdirAll(path, 0700); err != nil {
					t.Fatal(err)
				}
			case "invalid parent":
				if err := os.MkdirAll(configHome, 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(parent, []byte("not a directory"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			argument := ""
			if scenario.explicit {
				argument = path
			}
			cfg, err := Load(argument)
			if scenario.fixture == "valid" || scenario.fixture == "missing" && !scenario.explicit {
				if err != nil || cfg == nil {
					t.Fatalf("expected successful load: config = %v, error = %v", cfg, err)
				}
				wantDSN, wantWidth := ":memory:", 79
				if scenario.fixture == "missing" {
					wantDSN = "file:" + filepath.Join(root, "data", "troveler", "troveler.db") + "?cache=shared&mode=rwc"
					wantWidth = 50
				}
				if cfg.DSN != wantDSN || cfg.Search.TaglineWidth != wantWidth {
					t.Fatalf("dsn = %q, width = %d; want %q, %d", cfg.DSN, cfg.Search.TaglineWidth, wantDSN, wantWidth)
				}
				return
			}
			if err == nil || cfg != nil || !strings.Contains(err.Error(), path) {
				t.Fatalf("expected error identifying %q and no partial config: config = %v, error = %v", path, cfg, err)
			}
			if scenario.fixture == "malformed" {
				var parseErr toml.ParseError
				if !errors.As(err, &parseErr) {
					t.Fatalf("TOML parse error was not preserved: %v", err)
				}
			} else {
				var pathErr *os.PathError
				if !errors.As(err, &pathErr) || pathErr.Path != path {
					t.Fatalf("filesystem path error was not preserved: %v", err)
				}
				if scenario.fixture == "missing" && !errors.Is(err, os.ErrNotExist) {
					t.Fatalf("missing-file error was not preserved: %v", err)
				}
				if (scenario.fixture == "unreadable" || scenario.fixture == "inaccessible parent") && !errors.Is(err, os.ErrPermission) {
					t.Fatalf("permission error was not preserved: %v", err)
				}
			}
		})
	}
}
