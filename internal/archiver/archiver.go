package archiver

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Archive creates a ZIP file at zipPath containing the given files.
// Each file is stored with only its base name.
// Duplicate base names are resolved by appending _2, _3, etc.
// If zipPath already exists, it is overwritten.
func Archive(files []string, zipPath string) error {
	f, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer f.Close()

	w := zip.NewWriter(f)

	seen := make(map[string]int)
	for _, src := range files {
		if err := addFile(w, src, seen); err != nil {
			w.Close()
			return err
		}
	}

	return w.Close()
}

func addFile(w *zip.Writer, src string, seen map[string]int) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()

	base := filepath.Base(src)
	name := DedupeFilename(base, seen)
	seen[base]++

	dst, err := w.Create(name)
	if err != nil {
		return err
	}

	_, err = io.Copy(dst, f)
	return err
}

// DedupeFilename returns a unique filename given a seen count map.
// First occurrence keeps the original name; subsequent ones get _2, _3, etc.
func DedupeFilename(base string, seen map[string]int) string {
	count := seen[base]
	if count == 0 {
		return base
	}
	ext := filepath.Ext(base)
	stem := strings.TrimSuffix(base, ext)
	return fmt.Sprintf("%s_%d%s", stem, count+1, ext)
}
