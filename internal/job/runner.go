// internal/job/runner.go
package job

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/SilentMalachite/img2png/internal/archiver"
	"github.com/SilentMalachite/img2png/internal/converter"
)

// supportedExt mirrors main.supportedExt — keep in sync with that and the spec.
func supportedExt(ext string) bool {
	switch strings.ToLower(ext) {
	case ".tif", ".tiff", ".jpg", ".jpeg", ".webp", ".bmp", ".gif", ".png":
		// .png is included so the GUI accepts already-PNG inputs (passthrough copy).
		// Strictly the original CLI excludes .png; here we keep it because GUI users
		// may drop a PNG by mistake and expect it to "just work". Re-encoding a PNG
		// via converter.ConvertFile is harmless.
		return true
	}
	return false
}

// Run converts every file referenced by Items, sending an Event per item on eventCh.
// It does not close eventCh — the caller owns that channel.
// Cancellation: when ctx is canceled, the in-flight ConvertFile finishes and no
// further items start. Already-produced output files are kept.
func Run(ctx context.Context, j Job, eventCh chan<- Event) error {
	// Expand directories into a flat list of source files.
	srcs, err := expandItems(j.Items)
	if err != nil {
		return err
	}

	total := len(srcs)
	send(ctx, eventCh, Event{Kind: EventStart, Total: total})

	// Decide output destination per source.
	// In ZIP mode we route every PNG into a temp dir then archive.
	var tmpDir string
	if j.OutputMode == ModeZip && anyDir(j.Items) {
		td, err := os.MkdirTemp("", "img2png_*")
		if err != nil {
			return fmt.Errorf("temp dir: %w", err)
		}
		tmpDir = td
		defer os.RemoveAll(tmpDir)
	}

	seen := map[string]int{}
	var producedPNGs []string
	completed := 0

	for _, src := range srcs {
		if ctx.Err() != nil {
			break
		}

		outDir, fromZipBatch := destFor(src, j, tmpDir)
		base := strings.TrimSuffix(filepath.Base(src), filepath.Ext(src)) + ".png"
		name, skip := resolveName(base, outDir, j.Overwrite, seen, fromZipBatch)
		if skip {
			completed++
			send(ctx, eventCh, Event{Kind: EventItem, SourcePath: src, Status: StatusSkipped, Total: total, Completed: completed})
			continue
		}

		outPath, convErr := converter.ConvertFile(src, outDir, name)
		completed++
		if convErr != nil {
			send(ctx, eventCh, Event{Kind: EventItem, SourcePath: src, Status: StatusFailed, Err: convErr, Total: total, Completed: completed})
			continue
		}
		seen[base]++
		producedPNGs = append(producedPNGs, outPath)
		send(ctx, eventCh, Event{Kind: EventItem, SourcePath: src, OutputPath: outPath, Status: StatusDone, Total: total, Completed: completed})
	}

	// Finalize ZIP if needed.
	if tmpDir != "" && len(producedPNGs) > 0 {
		zipPath, err := finalizeZip(j, producedPNGs)
		if err != nil {
			send(ctx, eventCh, Event{Kind: EventDone, Err: err, Total: total, Completed: completed})
			return err
		}
		send(ctx, eventCh, Event{Kind: EventDone, OutputPath: zipPath, Total: total, Completed: completed})
		return nil
	}

	send(ctx, eventCh, Event{Kind: EventDone, Total: total, Completed: completed})
	return nil
}

// send is a non-blocking-friendly emit that respects ctx cancellation.
func send(ctx context.Context, ch chan<- Event, ev Event) {
	select {
	case ch <- ev:
	case <-ctx.Done():
	}
}

// expandItems walks any directory items and returns a flat list of supported files.
func expandItems(items []FileItem) ([]string, error) {
	var out []string
	for _, it := range items {
		if !it.IsDir {
			if supportedExt(filepath.Ext(it.Path)) {
				out = append(out, it.Path)
			}
			continue
		}
		_ = filepath.WalkDir(it.Path, func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if !d.IsDir() && supportedExt(filepath.Ext(p)) {
				out = append(out, p)
			}
			return nil
		})
	}
	return out, nil
}

func anyDir(items []FileItem) bool {
	for _, it := range items {
		if it.IsDir {
			return true
		}
	}
	return false
}

// destFor returns the output directory for src and whether it is the staging
// dir for a ZIP batch (in which case overwrite policy uses the seen map only,
// not on-disk inspection).
func destFor(src string, j Job, tmpDir string) (string, bool) {
	if tmpDir != "" {
		return tmpDir, true
	}
	if j.OutputDir != "" {
		return j.OutputDir, false
	}
	return filepath.Dir(src), false
}

// resolveName applies the overwrite policy to produce a final filename inside outDir.
// Returns (name, skip). When skip is true, the caller should record StatusSkipped
// and not call ConvertFile.
func resolveName(base, outDir string, policy OverwritePolicy, seen map[string]int, batchOnly bool) (string, bool) {
	switch policy {
	case PolicyIncrement:
		// Use seen-based dedupe (existing archiver helper); for non-batch dirs
		// also account for files that already exist on disk.
		name := archiver.DedupeFilename(base, seen)
		if batchOnly {
			return name, false
		}
		// Bump until the chosen name does not collide on disk.
		for {
			if _, err := os.Stat(filepath.Join(outDir, name)); os.IsNotExist(err) {
				return name, false
			}
			seen[base]++
			name = archiver.DedupeFilename(base, seen)
		}
	case PolicyOverwrite:
		return base, false
	case PolicySkip:
		if batchOnly {
			if seen[base] > 0 {
				return "", true
			}
			return base, false
		}
		if _, err := os.Stat(filepath.Join(outDir, base)); err == nil {
			return "", true
		}
		return base, false
	}
	return base, false
}

func finalizeZip(j Job, pngs []string) (string, error) {
	// ZIP path: prefer the explicit OutputDir, otherwise the parent of the
	// first directory item (matching current CLI behavior in main.runDir).
	var parentDir, dirName string
	for _, it := range j.Items {
		if it.IsDir {
			parentDir = filepath.Dir(it.Path)
			dirName = filepath.Base(it.Path)
			break
		}
	}
	if j.OutputDir != "" {
		parentDir = j.OutputDir
	}
	if dirName == "" {
		dirName = "img2png_output"
	}
	zipPath := filepath.Join(parentDir, dirName+".zip")
	if err := archiver.Archive(pngs, zipPath); err != nil {
		return "", err
	}
	return zipPath, nil
}
