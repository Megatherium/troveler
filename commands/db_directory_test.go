package commands

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"troveler/config"
	"troveler/db"
)

func TestWithDBCreatesDefaultDataDirectory(t *testing.T) {
	for _, name := range []string{"xdg", "home fallback", "relative xdg", "without home", "existing directory"} {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			home := filepath.Join(root, "home")
			data := filepath.Join(root, "custom data", "nested")
			t.Setenv("HOME", home)
			t.Setenv("XDG_CONFIG_HOME", filepath.Join(root, "config"))
			t.Setenv("XDG_DATA_HOME", data)
			t.Setenv("TROVELER_DSN", "")
			switch name {
			case "home fallback", "relative xdg":
				t.Setenv("XDG_DATA_HOME", "")
				if name == "relative xdg" {
					t.Setenv("XDG_DATA_HOME", "relative/data")
				}
				data = filepath.Join(home, ".local", "share")
			case "without home":
				t.Setenv("HOME", "")
			}
			dir := filepath.Join(data, "troveler")
			wantMode := os.FileMode(0700)
			if name == "existing directory" {
				if err := os.MkdirAll(dir, 0700); err != nil {
					t.Fatal(err)
				}
				wantMode = 0750
				if err := os.Chmod(dir, wantMode); err != nil {
					t.Fatal(err)
				}
			}
			cfg, err := config.Load("")
			if err != nil {
				t.Fatal(err)
			}
			if name != "existing directory" {
				if _, err := os.Stat(dir); !errors.Is(err, os.ErrNotExist) {
					t.Fatalf("Load must not create the data directory: %v", err)
				}
			}
			// Initialization must prepare the location captured during loading.
			t.Setenv("XDG_DATA_HOME", filepath.Join(root, "changed data"))
			cmd := &cobra.Command{}
			cmd.SetContext(WithConfig(context.Background(), cfg))
			if err := WithDB(cmd, func(ctx context.Context, database *db.SQLiteDB) error {
				return database.SaveToolSnapshot(ctx, &db.Tool{ID: "first", Slug: "first", Name: "First tool"}, nil)
			}); err != nil {
				t.Fatal(err)
			}
			info, err := os.Stat(dir)
			if err != nil || !info.IsDir() || info.Mode().Perm() != wantMode {
				t.Fatalf("directory info = %v, error = %v; want mode %o", info, err, wantMode)
			}
			if _, err := os.Stat(filepath.Join(dir, "troveler.db")); err != nil {
				t.Fatal(err)
			}
			if err := WithDB(cmd, func(ctx context.Context, database *db.SQLiteDB) error {
				tools, err := database.GetAllTools(ctx)
				if err != nil {
					return err
				}
				if len(tools) != 1 || tools[0].Name != "First tool" {
					return fmt.Errorf("persisted tools = %v", tools)
				}
				return nil
			}); err != nil {
				t.Fatal(err)
			}
			if _, err := os.Stat(filepath.Join(root, "changed data")); !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("initialization used changed environment: %v", err)
			}
		})
	}
}

func TestWithDBPreservesExplicitDSNs(t *testing.T) {
	for _, name := range []string{"config memory", "environment memory", "environment over default", "environment default path", "memory uri", "explicit file", "missing explicit parent", "changed config dsn", "manual config"} {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			blockedData := filepath.Join(root, "blocked data")
			if err := os.WriteFile(blockedData, []byte("not a directory"), 0600); err != nil {
				t.Fatal(err)
			}
			t.Setenv("HOME", root)
			t.Setenv("XDG_CONFIG_HOME", filepath.Join(root, "config"))
			t.Setenv("XDG_DATA_HOME", blockedData)
			t.Setenv("TROVELER_DSN", "")
			dsn := ":memory:"
			switch name {
			case "memory uri":
				dsn = "file:test-default-directory?mode=memory&cache=shared"
			case "explicit file":
				dsn = "file:" + filepath.Join(root, "explicit.db") + "?mode=rwc"
			case "missing explicit parent":
				dsn = "file:" + filepath.Join(root, "missing", "explicit.db") + "?mode=rwc"
			case "environment default path":
				dsn = "file:" + filepath.Join(blockedData, "troveler", "troveler.db") + "?cache=shared&mode=rwc"
			}
			path := ""
			if name == "environment over default" || name == "environment default path" {
				t.Setenv("TROVELER_DSN", dsn)
			}
			if name != "changed config dsn" && name != "manual config" && name != "environment over default" && name != "environment default path" {
				path = filepath.Join(root, "config.toml")
				fileDSN := dsn
				if name == "environment memory" {
					fileDSN = "file:/invalid/parent/overridden.db"
					t.Setenv("TROVELER_DSN", dsn)
				}
				if err := os.WriteFile(path, []byte(fmt.Sprintf("dsn = %q\n", fileDSN)), 0600); err != nil {
					t.Fatal(err)
				}
			}
			cfg, err := config.Load(path)
			if err != nil {
				t.Fatal(err)
			}
			if name == "changed config dsn" {
				cfg.DSN = dsn
			} else if name == "manual config" {
				cfg = &config.Config{DSN: dsn}
			}
			cmd := &cobra.Command{}
			cmd.SetContext(WithConfig(context.Background(), cfg))
			called := false
			err = WithDB(cmd, func(ctx context.Context, database *db.SQLiteDB) error {
				called = true
				count, err := database.ToolCount(ctx)
				if err == nil && count != 0 {
					return fmt.Errorf("fresh database has %d tools", count)
				}
				return err
			})
			if name == "environment default path" {
				var pathErr *os.PathError
				if err == nil || called || errors.As(err, &pathErr) || !strings.Contains(err.Error(), "failed to ping db") {
					t.Fatalf("explicit environment DSN matching default must bypass preparation: error = %v, called = %v", err, called)
				}
			} else if name == "missing explicit parent" {
				if err == nil || called {
					t.Fatalf("explicit missing parent must still fail: error = %v, called = %v", err, called)
				}
				if _, err := os.Stat(filepath.Join(root, "missing")); !errors.Is(err, os.ErrNotExist) {
					t.Fatalf("created an explicit DSN's parent: %v", err)
				}
			} else if err != nil || !called {
				t.Fatalf("explicit DSN failed: error = %v, called = %v", err, called)
			}
		})
	}
}

func TestWithDBReportsDefaultDirectoryFailure(t *testing.T) {
	root := t.TempDir()
	data := filepath.Join(root, "data")
	if err := os.Mkdir(data, 0700); err != nil {
		t.Fatal(err)
	}
	blockedDir := filepath.Join(data, "troveler")
	if err := os.WriteFile(blockedDir, []byte("keep this file"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", root)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(root, "config"))
	t.Setenv("XDG_DATA_HOME", data)
	t.Setenv("TROVELER_DSN", "")
	cfg, err := config.Load("")
	if err != nil {
		t.Fatal(err)
	}
	cmd := &cobra.Command{}
	cmd.SetContext(WithConfig(context.Background(), cfg))
	called := false
	err = WithDB(cmd, func(context.Context, *db.SQLiteDB) error {
		called = true
		return nil
	})
	var pathErr *os.PathError
	if err == nil || called || !errors.As(err, &pathErr) || !strings.Contains(err.Error(), "create default database directory") || !strings.Contains(err.Error(), blockedDir) {
		t.Fatalf("expected contextual directory error before callback: error = %v, called = %v", err, called)
	}
	contents, readErr := os.ReadFile(blockedDir)
	if readErr != nil || string(contents) != "keep this file" {
		t.Fatalf("blocking file changed: %q, error = %v", contents, readErr)
	}
}
