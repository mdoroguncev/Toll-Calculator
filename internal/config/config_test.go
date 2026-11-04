package config

import (
	"os"
	"path/filepath"
	"testing"
)

// unsetenv removes key for the duration of the test and restores it afterwards.
func unsetenv(t *testing.T, key string) {
	t.Helper()
	t.Setenv(key, "")
	os.Unsetenv(key)
}

func TestGet(t *testing.T) {
	const key = "TOLL_CONFIG_TEST_KEY"

	tests := []struct {
		name  string
		set   bool
		value string
		want  string
	}{
		{name: "not set returns default", set: false, want: "default"},
		{name: "set returns value", set: true, value: "kafka:29092", want: "kafka:29092"},
		{name: "empty returns default", set: true, value: "", want: "default"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.set {
				t.Setenv(key, tt.value)
			} else {
				unsetenv(t, key)
			}

			if got := Get(key, "default"); got != tt.want {
				t.Errorf("Get(%q) = %q, want %q", key, got, tt.want)
			}
		})
	}
}

func TestLoadDotEnv(t *testing.T) {
	t.Run("missing file is not an error", func(t *testing.T) {
		t.Chdir(t.TempDir())

		if err := LoadDotEnv(); err != nil {
			t.Fatalf("LoadDotEnv() without .env: %v", err)
		}
	})

	t.Run("loads file but keeps existing variables", func(t *testing.T) {
		dir := t.TempDir()
		content := "TOLL_FROM_FILE=file\nTOLL_ALREADY_SET=file\n"
		if err := os.WriteFile(filepath.Join(dir, ".env"), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
		t.Chdir(dir)
		unsetenv(t, "TOLL_FROM_FILE")
		t.Setenv("TOLL_ALREADY_SET", "env")

		if err := LoadDotEnv(); err != nil {
			t.Fatalf("LoadDotEnv(): %v", err)
		}
		if got := os.Getenv("TOLL_FROM_FILE"); got != "file" {
			t.Errorf("TOLL_FROM_FILE = %q, want %q", got, "file")
		}
		if got := os.Getenv("TOLL_ALREADY_SET"); got != "env" {
			t.Errorf("TOLL_ALREADY_SET = %q, want %q (environment must win over .env)", got, "env")
		}
	})
}
