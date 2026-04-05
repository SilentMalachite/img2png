package archiver

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func createTempFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func readZipEntries(t *testing.T, zipPath string) map[string]string {
	t.Helper()
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	entries := make(map[string]string)
	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			t.Fatal(err)
		}
		data, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			t.Fatal(err)
		}
		entries[f.Name] = string(data)
	}
	return entries
}

func TestArchive_Basic(t *testing.T) {
	srcDir := t.TempDir()
	files := []string{
		createTempFile(t, srcDir, "a.png", "data-a"),
		createTempFile(t, srcDir, "b.png", "data-b"),
		createTempFile(t, srcDir, "c.png", "data-c"),
	}

	zipPath := filepath.Join(t.TempDir(), "out.zip")
	if err := Archive(files, zipPath); err != nil {
		t.Fatal(err)
	}

	entries := readZipEntries(t, zipPath)
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}
	for _, name := range []string{"a.png", "b.png", "c.png"} {
		if _, ok := entries[name]; !ok {
			t.Errorf("missing entry: %s", name)
		}
	}
}

func TestArchive_Overwrite(t *testing.T) {
	srcDir := t.TempDir()
	zipPath := filepath.Join(t.TempDir(), "out.zip")

	files1 := []string{
		createTempFile(t, srcDir, "old.png", "old-data"),
	}
	if err := Archive(files1, zipPath); err != nil {
		t.Fatal(err)
	}

	files2 := []string{
		createTempFile(t, srcDir, "new.png", "new-data"),
	}
	if err := Archive(files2, zipPath); err != nil {
		t.Fatal(err)
	}

	entries := readZipEntries(t, zipPath)
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry after overwrite, got %d", len(entries))
	}
	if _, ok := entries["new.png"]; !ok {
		t.Error("expected new.png after overwrite")
	}
}

func TestArchive_DuplicateNames(t *testing.T) {
	// 3つの異なるディレクトリに同名ファイルを配置
	dir1 := t.TempDir()
	dir2 := t.TempDir()
	dir3 := t.TempDir()

	files := []string{
		createTempFile(t, dir1, "photo.png", "first"),
		createTempFile(t, dir2, "photo.png", "second"),
		createTempFile(t, dir3, "photo.png", "third"),
	}

	zipPath := filepath.Join(t.TempDir(), "out.zip")
	if err := Archive(files, zipPath); err != nil {
		t.Fatal(err)
	}

	entries := readZipEntries(t, zipPath)
	expected := map[string]string{
		"photo.png":   "first",
		"photo_2.png": "second",
		"photo_3.png": "third",
	}
	for name, wantContent := range expected {
		got, ok := entries[name]
		if !ok {
			t.Errorf("missing entry: %s", name)
			continue
		}
		if got != wantContent {
			t.Errorf("%s: got %q, want %q", name, got, wantContent)
		}
	}
}

func TestArchive_MixedDuplicates(t *testing.T) {
	dir1 := t.TempDir()
	dir2 := t.TempDir()
	dir3 := t.TempDir()

	files := []string{
		createTempFile(t, dir1, "a.png", "a1"),
		createTempFile(t, dir2, "b.png", "b1"),
		createTempFile(t, dir3, "a.png", "a2"),
	}

	zipPath := filepath.Join(t.TempDir(), "out.zip")
	if err := Archive(files, zipPath); err != nil {
		t.Fatal(err)
	}

	entries := readZipEntries(t, zipPath)
	expected := map[string]string{
		"a.png":   "a1",
		"b.png":   "b1",
		"a_2.png": "a2",
	}
	for name, wantContent := range expected {
		got, ok := entries[name]
		if !ok {
			t.Errorf("missing entry: %s", name)
			continue
		}
		if got != wantContent {
			t.Errorf("%s: got %q, want %q", name, got, wantContent)
		}
	}
}

func TestDedupeFilename(t *testing.T) {
	seen := make(map[string]int)

	// 初回はそのまま
	got := DedupeFilename("photo.png", seen)
	if got != "photo.png" {
		t.Errorf("first: got %q, want %q", got, "photo.png")
	}
	seen["photo.png"]++

	// 2回目は _2
	got = DedupeFilename("photo.png", seen)
	if got != "photo_2.png" {
		t.Errorf("second: got %q, want %q", got, "photo_2.png")
	}
	seen["photo.png"]++

	// 3回目は _3
	got = DedupeFilename("photo.png", seen)
	if got != "photo_3.png" {
		t.Errorf("third: got %q, want %q", got, "photo_3.png")
	}
}
