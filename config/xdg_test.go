package config

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadXDGDirectories(t *testing.T) {
	for _, scenario := range []struct {
		name                         string
		configLocation, dataLocation string
		unset, noHome                bool
	}{
		{name: "both custom", configLocation: "absolute", dataLocation: "absolute"},
		{name: "config only", configLocation: "absolute"},
		{name: "data only", dataLocation: "absolute"},
		{name: "empty"},
		{name: "unset", unset: true},
		{name: "relative values", configLocation: "relative", dataLocation: "relative"},
		{name: "absolute config and relative data", configLocation: "absolute", dataLocation: "relative"},
		{name: "absolute XDG without HOME", configLocation: "absolute", dataLocation: "absolute", noHome: true},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			root := t.TempDir()
			home := filepath.Join(root, "home")
			configHome := filepath.Join(root, "custom config")
			dataHome := filepath.Join(root, "custom data")
			t.Setenv("HOME", home)
			t.Setenv("TROVELER_DSN", "")
			writeXDGConfigForTest(t, filepath.Join(home, ".config", "troveler", "config.toml"), "[search]\ntagline_width = 11\n")
			writeXDGConfigForTest(t, filepath.Join(configHome, "troveler", "config.toml"), "[search]\ntagline_width = 73\n")
			locations := map[string]string{"XDG_CONFIG_HOME": scenario.configLocation, "XDG_DATA_HOME": scenario.dataLocation}
			for variable, location := range locations {
				value := ""
				if location == "absolute" {
					value = configHome
					if variable == "XDG_DATA_HOME" {
						value = dataHome
					}
				} else if location == "relative" {
					value = "relative/xdg"
				}
				t.Setenv(variable, value)
				if scenario.unset {
					if err := os.Unsetenv(variable); err != nil {
						t.Fatal(err)
					}
				}
			}
			if scenario.noHome {
				t.Setenv("HOME", "")
			}
			cfg, err := Load("")
			if err != nil {
				t.Fatal(err)
			}
			wantWidth := 11
			if scenario.configLocation == "absolute" {
				wantWidth = 73
			}
			if cfg.Search.TaglineWidth != wantWidth {
				t.Errorf("loaded tagline_width=%d; want %d from the selected config file", cfg.Search.TaglineWidth, wantWidth)
			}
			dataBase := filepath.Join(home, ".local", "share")
			if scenario.dataLocation == "absolute" {
				dataBase = dataHome
			}
			wantDSN := "file:" + filepath.Join(dataBase, "troveler", "troveler.db") + "?cache=shared&mode=rwc"
			if cfg.DSN != wantDSN {
				t.Errorf("dsn=%q; want %q", cfg.DSN, wantDSN)
			}
		})
	}
}

func TestMissingXDGConfigDoesNotLoadHomeConfig(t *testing.T) {
	root := t.TempDir()
	t.Setenv("HOME", filepath.Join(root, "home"))
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(root, "missing config"))
	t.Setenv("XDG_DATA_HOME", filepath.Join(root, "data"))
	t.Setenv("TROVELER_DSN", "")
	writeXDGConfigForTest(t, filepath.Join(root, "home", ".config", "troveler", "config.toml"), "default_to_tui = true\n")
	cfg, err := Load("")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.DefaultToTUI || cfg.Search.TaglineWidth != 50 {
		t.Fatal("missing custom config must use defaults rather than a HOME config")
	}
}

func TestXDGConfigAndDSNOverrides(t *testing.T) {
	for _, scenario := range []struct {
		name, fileDSN, envDSN string
		explicit              bool
	}{
		{name: "configured DSN", fileDSN: ":memory:"},
		{name: "environment overrides configured DSN", fileDSN: ":memory:", envDSN: "file:override.db?mode=memory&cache=shared"},
		{name: "environment overrides default DSN", envDSN: ":memory:"},
		{name: "explicit config overrides XDG lookup", fileDSN: "file:explicit.db?mode=memory&cache=shared", explicit: true},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			root := t.TempDir()
			t.Setenv("HOME", filepath.Join(root, "home"))
			t.Setenv("XDG_CONFIG_HOME", filepath.Join(root, "config"))
			t.Setenv("XDG_DATA_HOME", filepath.Join(root, "data"))
			t.Setenv("TROVELER_DSN", scenario.envDSN)
			writeXDGConfigForTest(t, filepath.Join(root, "config", "troveler", "config.toml"), fmt.Sprintf("dsn = %q\n[search]\ntagline_width = 73\n", scenario.fileDSN))
			configPath, wantWidth := "", 73
			if scenario.explicit {
				configPath, wantWidth = filepath.Join(root, "explicit.toml"), 19
				writeXDGConfigForTest(t, configPath, fmt.Sprintf("dsn = %q\n[search]\ntagline_width = 19\n", scenario.fileDSN))
			}
			cfg, err := Load(configPath)
			if err != nil {
				t.Fatal(err)
			}
			wantDSN := scenario.fileDSN
			if scenario.envDSN != "" {
				wantDSN = scenario.envDSN
			}
			if cfg.DSN != wantDSN || cfg.Search.TaglineWidth != wantWidth {
				t.Fatalf("dsn=%q, width=%d; want dsn=%q, width=%d", cfg.DSN, cfg.Search.TaglineWidth, wantDSN, wantWidth)
			}
		})
	}
}

func writeXDGConfigForTest(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
}
