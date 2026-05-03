// Package gui contains the Fyne-based GUI for img2png.
package gui

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/SilentMalachite/img2png/internal/job"
)

// supportedExt mirrors job.supportedExt. Kept local to avoid coupling the GUI
// to internal job constants beyond what's needed.
func supportedExt(ext string) bool {
	switch strings.ToLower(ext) {
	case ".tif", ".tiff", ".jpg", ".jpeg", ".webp", ".bmp", ".gif", ".png":
		return true
	}
	return false
}

// AddItems validates and appends paths to the existing list, dropping
// duplicates (by Path) and unsupported / missing entries. Directories are
// always accepted (they may contain supported files).
func AddItems(existing []job.FileItem, paths []string) []job.FileItem {
	seen := make(map[string]struct{}, len(existing))
	for _, it := range existing {
		seen[it.Path] = struct{}{}
	}
	out := append([]job.FileItem(nil), existing...)
	for _, p := range paths {
		p = filepath.Clean(p)
		if _, dup := seen[p]; dup {
			continue
		}
		info, err := os.Stat(p)
		if err != nil {
			continue
		}
		if !info.IsDir() && !supportedExt(filepath.Ext(p)) {
			continue
		}
		seen[p] = struct{}{}
		out = append(out, job.FileItem{Path: p, IsDir: info.IsDir()})
	}
	return out
}

// RemoveAt returns a copy of items with index i removed. Out-of-range is a no-op.
func RemoveAt(items []job.FileItem, i int) []job.FileItem {
	if i < 0 || i >= len(items) {
		return items
	}
	out := make([]job.FileItem, 0, len(items)-1)
	out = append(out, items[:i]...)
	out = append(out, items[i+1:]...)
	return out
}
