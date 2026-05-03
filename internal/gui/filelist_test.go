// internal/gui/filelist_test.go
package gui

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/SilentMalachite/img2png/internal/job"
)

func TestAddItems_DedupesByPath(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a.jpg")
	_ = os.WriteFile(a, []byte("x"), 0o644)

	got := AddItems(nil, []string{a, a})
	if len(got) != 1 {
		t.Errorf("expected 1 item after dedup, got %d", len(got))
	}
}

func TestAddItems_RejectsUnsupportedExtension(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "doc.txt")
	_ = os.WriteFile(a, []byte("x"), 0o644)

	got := AddItems(nil, []string{a})
	if len(got) != 0 {
		t.Errorf("expected unsupported file to be rejected, got %d items", len(got))
	}
}

func TestAddItems_RejectsMissingPath(t *testing.T) {
	got := AddItems(nil, []string{"/nope/does/not/exist.jpg"})
	if len(got) != 0 {
		t.Errorf("expected missing path to be rejected, got %d items", len(got))
	}
}

func TestAddItems_AcceptsDirectory(t *testing.T) {
	dir := t.TempDir()
	got := AddItems(nil, []string{dir})
	if len(got) != 1 || !got[0].IsDir {
		t.Errorf("expected one directory item, got %+v", got)
	}
}

func TestRemoveAt(t *testing.T) {
	items := []job.FileItem{{Path: "a"}, {Path: "b"}, {Path: "c"}}
	got := RemoveAt(items, 1)
	if len(got) != 2 || got[0].Path != "a" || got[1].Path != "c" {
		t.Errorf("RemoveAt(1) = %+v", got)
	}
}

func TestRemoveAt_OutOfRange_NoOp(t *testing.T) {
	items := []job.FileItem{{Path: "a"}}
	got := RemoveAt(items, 5)
	if len(got) != 1 {
		t.Errorf("out-of-range remove should be no-op, got %+v", got)
	}
}
