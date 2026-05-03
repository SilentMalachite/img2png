// internal/settings/settings_test.go
package settings

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveLoad_Roundtrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")

	in := Settings{
		OutputDir:    "/tmp/out",
		OutputMode:   "zip",
		Overwrite:    "increment",
		Language:     "ja",
		WindowWidth:  900,
		WindowHeight: 600,
	}
	if err := SaveTo(path, in); err != nil {
		t.Fatal(err)
	}
	got, err := LoadFrom(path)
	if err != nil {
		t.Fatal(err)
	}
	if got != in {
		t.Errorf("roundtrip mismatch:\n got: %+v\nwant: %+v", got, in)
	}
}

func TestLoad_MissingFile_ReturnsDefaults(t *testing.T) {
	got, err := LoadFrom(filepath.Join(t.TempDir(), "does-not-exist.json"))
	if err != nil {
		t.Fatalf("expected no error for missing file, got %v", err)
	}
	def := Defaults()
	if got != def {
		t.Errorf("expected defaults, got %+v", got)
	}
}

func TestLoad_CorruptFile_ReturnsDefaults(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")
	if err := os.WriteFile(path, []byte("not json {{{"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := LoadFrom(path)
	if err != nil {
		t.Fatalf("expected no error for corrupt file, got %v", err)
	}
	if got != Defaults() {
		t.Errorf("expected defaults on corrupt file, got %+v", got)
	}
}
