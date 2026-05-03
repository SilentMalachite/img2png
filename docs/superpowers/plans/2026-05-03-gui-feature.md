# img2png GUI Feature Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a Fyne-based GUI to img2png such that running the binary with no arguments opens a window with file list, settings panel, drag-and-drop, progress, and cancel — while leaving the existing CLI behavior 100% intact when arguments are passed.

**Architecture:** Single binary that dispatches in `main()` based on `os.Args`. New packages `internal/gui` (Fyne UI), `internal/job` (orchestration), `internal/settings` (persistence), `internal/i18n` (JA/EN labels). Existing `internal/converter` and `internal/archiver` are reused unchanged.

**Tech Stack:** Go 1.26, Fyne v2 (`fyne.io/fyne/v2`), `golang.org/x/image` (already present). CGO=1 (newly required). Native packaging via `fyne package` on each OS runner in CI.

**Spec:** [`docs/superpowers/specs/2026-05-03-gui-feature-design.md`](../specs/2026-05-03-gui-feature-design.md)

---

## File Map

### Created

| File | Responsibility |
|---|---|
| `FyneApp.toml` | App metadata (name, ID, version) consumed by `fyne package` |
| `assets/icon.png` | 512×512 PNG used for `.app` / `.exe` icon |
| `cmd/genicon/main.go` | One-off generator for the placeholder icon |
| `internal/job/types.go` | Shared types: `FileItem`, `Status`, `Mode`, `OverwritePolicy`, `Event`, `Job` |
| `internal/job/runner.go` | `Run(ctx, job, eventCh)` — sequential conversion with cancel |
| `internal/job/runner_test.go` | TDD tests for runner |
| `internal/settings/settings.go` | `Load` / `Save` of JSON settings, OS-standard path |
| `internal/settings/settings_test.go` | TDD tests |
| `internal/i18n/i18n.go` | `Translator` with JA/EN tables, locale auto-detect |
| `internal/i18n/i18n_test.go` | TDD tests |
| `internal/gui/filelist.go` | File list state + pure helpers (validate, dedupe) + Fyne widget |
| `internal/gui/filelist_test.go` | TDD tests for pure helpers |
| `internal/gui/settings_panel.go` | Settings panel widget + `BuildJob` pure helper |
| `internal/gui/settings_panel_test.go` | TDD tests for `BuildJob` |
| `internal/gui/progress.go` | Progress widget |
| `internal/gui/menu.go` | Menu construction |
| `internal/gui/app.go` | `Run()` — initializes app and assembles main window |

### Modified

| File | Change |
|---|---|
| `go.mod`, `go.sum` | Add `fyne.io/fyne/v2` |
| `main.go` | Dispatch: no args → `gui.Run()`, args → existing CLI logic, `--help` flag |
| `.github/workflows/test.yml` | Switch to `CGO_ENABLED=1`, install Linux GUI deps |
| `.github/workflows/build.yml` | Native runners per OS, use `fyne package` |
| `README.md` | Remove "CGO 不使用" claim, add GUI section, update artifact names |

---

## Task 1: Add Fyne dependency and Go version sanity-check

**Files:**
- Modify: `go.mod`
- Modify: `go.sum`

- [ ] **Step 1: Add Fyne to module dependencies**

```bash
go get fyne.io/fyne/v2@v2.6.0
```

Expected: `go.mod` now has a `fyne.io/fyne/v2 v2.6.x` line under `require`. `go.sum` is updated with hashes.

- [ ] **Step 2: Verify build still works (existing CLI)**

```bash
go build ./...
```

Expected: Builds without error. (At this stage there's no GUI code yet, only the dependency download.)

- [ ] **Step 3: Run existing tests to confirm baseline**

```bash
go test -race ./...
```

Expected: All existing converter/archiver tests pass.

- [ ] **Step 4: Commit**

```bash
git add go.mod go.sum
git commit -m "deps: add fyne.io/fyne/v2 for upcoming GUI"
```

---

## Task 2: Create app metadata file

**Files:**
- Create: `FyneApp.toml`

- [ ] **Step 1: Write `FyneApp.toml`**

```toml
[Details]
Icon = "assets/icon.png"
Name = "img2png"
ID = "com.silentmalachite.img2png"
Version = "1.0.0"
Build = 1
```

- [ ] **Step 2: Commit**

```bash
git add FyneApp.toml
git commit -m "build: add FyneApp.toml for fyne package metadata"
```

---

## Task 3: Add placeholder icon

**Files:**
- Create: `cmd/genicon/main.go`
- Create: `assets/icon.png` (binary, generated)

The icon must exist before `fyne package` runs in CI. We generate a simple solid-color 512×512 placeholder; a designer can replace `assets/icon.png` later without code changes.

- [ ] **Step 1: Write the generator**

```go
// cmd/genicon/main.go
// Generates assets/icon.png — a 512×512 solid-color placeholder.
// Run once: go run ./cmd/genicon
package main

import (
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
)

func main() {
	const size = 512
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	bg := color.RGBA{R: 0x2a, G: 0x68, B: 0xa8, A: 0xff}
	fg := color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			img.Set(x, y, bg)
		}
	}
	// Simple "i2p" mark: three filled squares forming a vertical line / arrow hint.
	// Centered block.
	for y := 192; y < 320; y++ {
		for x := 192; x < 320; x++ {
			img.Set(x, y, fg)
		}
	}

	if err := os.MkdirAll("assets", 0o755); err != nil {
		log.Fatal(err)
	}
	f, err := os.Create("assets/icon.png")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		log.Fatal(err)
	}
	log.Println("wrote assets/icon.png")
}
```

- [ ] **Step 2: Run the generator**

```bash
go run ./cmd/genicon
```

Expected stdout: `wrote assets/icon.png`. File `assets/icon.png` exists.

- [ ] **Step 3: Verify the PNG is valid**

```bash
file assets/icon.png
```

Expected output: `assets/icon.png: PNG image data, 512 x 512, 8-bit/color RGBA, non-interlaced`.

- [ ] **Step 4: Commit**

```bash
git add cmd/genicon/main.go assets/icon.png
git commit -m "build: add placeholder app icon and generator"
```

---

## Task 4: Define `internal/job` types

**Files:**
- Create: `internal/job/types.go`

These types are shared between the runner and the GUI; defining them up front lets later TDD tests reference exact names.

- [ ] **Step 1: Write the types file**

```go
// Package job defines the conversion job types and a runner that converts
// images sequentially with cancellation support.
package job

// Status is the per-item lifecycle state.
type Status int

const (
	StatusPending Status = iota
	StatusRunning
	StatusDone
	StatusFailed
	StatusSkipped
)

// Mode is the output mode when the input is a directory.
type Mode int

const (
	ModeZip Mode = iota
	ModeIndividual
)

// OverwritePolicy controls how same-name PNGs are handled at the output dir.
type OverwritePolicy int

const (
	PolicyIncrement OverwritePolicy = iota // photo.png → photo_2.png
	PolicyOverwrite
	PolicySkip
)

// FileItem is one entry in the GUI file list. Path may point to a file or directory.
type FileItem struct {
	Path  string
	IsDir bool
}

// Job is the converted-side request: what to convert and how.
type Job struct {
	Items      []FileItem
	OutputDir  string // empty → derive from each Item per current CLI rules
	OutputMode Mode
	Overwrite  OverwritePolicy
}

// EventKind names a progress event.
type EventKind int

const (
	EventStart EventKind = iota // emitted once before the first item
	EventItem                   // per converted file (or skip)
	EventDone                   // emitted once after the last item
)

// Event carries one progress notification on the runner's output channel.
// SourcePath is the original file (not the produced PNG) for EventItem.
type Event struct {
	Kind       EventKind
	SourcePath string
	OutputPath string // populated on success
	Status     Status // StatusDone / StatusFailed / StatusSkipped for EventItem
	Err        error
	Total      int // total expected files (set on EventStart and EventDone)
	Completed  int // running count of completed (any status) at this event
}
```

- [ ] **Step 2: Verify it compiles**

```bash
go build ./internal/job
```

Expected: no output, exit 0.

- [ ] **Step 3: Commit**

```bash
git add internal/job/types.go
git commit -m "feat(job): define job/event/status types"
```

---

## Task 5: Implement `job.Run` — single file success

**Files:**
- Create: `internal/job/runner.go`
- Create: `internal/job/runner_test.go`

- [ ] **Step 1: Add a tiny PNG fixture helper to the test file**

Tests need a real source image. We write a 1×1 PNG via the standard library at test setup time.

```go
// internal/job/runner_test.go
package job

import (
	"context"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

// writePNG creates a 1×1 PNG at path. Used as a synthetic input for tests
// (ConvertFile decodes any registered format; PNG decodes via image/png which
// is registered transitively through the converter package).
func writePNG(t *testing.T, path string) {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.RGBA{R: 1, G: 2, B: 3, A: 255})
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
}

// drain reads all events from ch into a slice (channel must be closed).
func drain(ch <-chan Event) []Event {
	var out []Event
	for ev := range ch {
		out = append(out, ev)
	}
	return out
}
```

- [ ] **Step 2: Write the failing test for a single-file individual-PNG job**

Append to `internal/job/runner_test.go`:

```go
func TestRun_SingleFile_Individual(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "in.png")
	writePNG(t, src)

	out := filepath.Join(dir, "out")
	if err := os.MkdirAll(out, 0o755); err != nil {
		t.Fatal(err)
	}

	j := Job{
		Items:      []FileItem{{Path: src, IsDir: false}},
		OutputDir:  out,
		OutputMode: ModeIndividual,
		Overwrite:  PolicyIncrement,
	}

	ch := make(chan Event, 8)
	if err := Run(context.Background(), j, ch); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	close(ch)

	events := drain(ch)
	if len(events) < 3 {
		t.Fatalf("expected at least Start/Item/Done events, got %d", len(events))
	}
	if events[0].Kind != EventStart || events[0].Total != 1 {
		t.Errorf("first event should be EventStart with Total=1, got %+v", events[0])
	}
	if events[len(events)-1].Kind != EventDone {
		t.Errorf("last event should be EventDone, got %+v", events[len(events)-1])
	}

	// File created
	if _, err := os.Stat(filepath.Join(out, "in.png")); err != nil {
		t.Errorf("expected output PNG: %v", err)
	}
}
```

- [ ] **Step 3: Run the test, verify it fails to compile**

```bash
go test ./internal/job/...
```

Expected: build error `undefined: Run`.

- [ ] **Step 4: Implement minimal `runner.go`**

```go
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
```

- [ ] **Step 5: Run the test, verify it passes**

```bash
go test -race ./internal/job/...
```

Expected: PASS for `TestRun_SingleFile_Individual`.

- [ ] **Step 6: Commit**

```bash
git add internal/job/runner.go internal/job/runner_test.go
git commit -m "feat(job): implement sequential runner with start/item/done events"
```

---

## Task 6: `job.Run` — directory walk and ZIP mode

**Files:**
- Modify: `internal/job/runner_test.go`

- [ ] **Step 1: Add the failing test for directory + ZIP mode**

Append to `internal/job/runner_test.go`:

```go
func TestRun_Directory_ZipMode(t *testing.T) {
	root := t.TempDir()
	srcDir := filepath.Join(root, "photos")
	if err := os.MkdirAll(srcDir, 0o755); err != nil {
		t.Fatal(err)
	}
	writePNG(t, filepath.Join(srcDir, "a.png"))
	writePNG(t, filepath.Join(srcDir, "b.png"))

	j := Job{
		Items:      []FileItem{{Path: srcDir, IsDir: true}},
		OutputDir:  "", // default → parent of dir
		OutputMode: ModeZip,
		Overwrite:  PolicyIncrement,
	}

	ch := make(chan Event, 16)
	if err := Run(context.Background(), j, ch); err != nil {
		t.Fatalf("Run: %v", err)
	}
	close(ch)

	events := drain(ch)
	last := events[len(events)-1]
	if last.Kind != EventDone {
		t.Fatalf("last event Kind=%v want EventDone", last.Kind)
	}
	if last.OutputPath == "" {
		t.Fatalf("expected EventDone.OutputPath to point to the zip")
	}
	if filepath.Base(last.OutputPath) != "photos.zip" {
		t.Errorf("zip name = %q, want photos.zip", filepath.Base(last.OutputPath))
	}
	if _, err := os.Stat(last.OutputPath); err != nil {
		t.Errorf("zip not created: %v", err)
	}
}
```

- [ ] **Step 2: Run, expect pass without further code changes**

```bash
go test -race ./internal/job/...
```

Expected: PASS for both tests. (The runner already handles this path; the test confirms it.)

- [ ] **Step 3: Commit**

```bash
git add internal/job/runner_test.go
git commit -m "test(job): cover directory + ZIP mode"
```

---

## Task 7: `job.Run` — cancellation behavior

**Files:**
- Modify: `internal/job/runner_test.go`

- [ ] **Step 1: Add the failing test for cancellation**

Append to `internal/job/runner_test.go`:

```go
func TestRun_Cancel_StopsBeforeRemainingItems(t *testing.T) {
	dir := t.TempDir()
	src1 := filepath.Join(dir, "a.png")
	src2 := filepath.Join(dir, "b.png")
	src3 := filepath.Join(dir, "c.png")
	writePNG(t, src1)
	writePNG(t, src2)
	writePNG(t, src3)

	out := filepath.Join(dir, "out")
	_ = os.MkdirAll(out, 0o755)

	j := Job{
		Items: []FileItem{
			{Path: src1}, {Path: src2}, {Path: src3},
		},
		OutputDir:  out,
		OutputMode: ModeIndividual,
		Overwrite:  PolicyIncrement,
	}

	ctx, cancel := context.WithCancel(context.Background())
	ch := make(chan Event, 32)

	// Cancel after the first item event.
	go func() {
		for ev := range ch {
			if ev.Kind == EventItem {
				cancel()
				// Drain remaining without acting on them.
				for range ch {
				}
				return
			}
		}
	}()

	if err := Run(ctx, j, ch); err != nil {
		t.Fatalf("Run: %v", err)
	}
	close(ch)

	// At least one PNG should exist (the one before cancel).
	entries, _ := os.ReadDir(out)
	if len(entries) == 0 {
		t.Errorf("expected at least one produced PNG before cancel")
	}
	if len(entries) == 3 {
		t.Errorf("expected cancel to skip remaining items, but got all 3")
	}
}
```

- [ ] **Step 2: Run, expect pass**

```bash
go test -race ./internal/job/...
```

Expected: PASS.

- [ ] **Step 3: Commit**

```bash
git add internal/job/runner_test.go
git commit -m "test(job): cover cancellation stops remaining items"
```

---

## Task 8: `job.Run` — overwrite policies

**Files:**
- Modify: `internal/job/runner_test.go`

- [ ] **Step 1: Add tests for each overwrite policy**

Append to `internal/job/runner_test.go`:

```go
func TestRun_PolicyIncrement_AppendsSuffix(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "photo.png")
	writePNG(t, src)

	out := filepath.Join(dir, "out")
	_ = os.MkdirAll(out, 0o755)

	// Pre-existing photo.png at output location.
	if err := os.WriteFile(filepath.Join(out, "photo.png"), []byte("existing"), 0o644); err != nil {
		t.Fatal(err)
	}

	j := Job{
		Items:      []FileItem{{Path: src}},
		OutputDir:  out,
		OutputMode: ModeIndividual,
		Overwrite:  PolicyIncrement,
	}
	ch := make(chan Event, 8)
	if err := Run(context.Background(), j, ch); err != nil {
		t.Fatal(err)
	}
	close(ch)

	if _, err := os.Stat(filepath.Join(out, "photo_2.png")); err != nil {
		t.Errorf("expected photo_2.png to be created: %v", err)
	}
	// Pre-existing file untouched
	b, _ := os.ReadFile(filepath.Join(out, "photo.png"))
	if string(b) != "existing" {
		t.Errorf("original photo.png was overwritten")
	}
}

func TestRun_PolicyOverwrite_ReplacesExisting(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "photo.png")
	writePNG(t, src)

	out := filepath.Join(dir, "out")
	_ = os.MkdirAll(out, 0o755)
	_ = os.WriteFile(filepath.Join(out, "photo.png"), []byte("existing"), 0o644)

	j := Job{
		Items: []FileItem{{Path: src}}, OutputDir: out,
		OutputMode: ModeIndividual, Overwrite: PolicyOverwrite,
	}
	ch := make(chan Event, 8)
	_ = Run(context.Background(), j, ch)
	close(ch)

	b, _ := os.ReadFile(filepath.Join(out, "photo.png"))
	if string(b) == "existing" {
		t.Errorf("expected photo.png to be overwritten with PNG bytes")
	}
}

func TestRun_PolicySkip_LeavesExistingAlone(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "photo.png")
	writePNG(t, src)

	out := filepath.Join(dir, "out")
	_ = os.MkdirAll(out, 0o755)
	_ = os.WriteFile(filepath.Join(out, "photo.png"), []byte("existing"), 0o644)

	j := Job{
		Items: []FileItem{{Path: src}}, OutputDir: out,
		OutputMode: ModeIndividual, Overwrite: PolicySkip,
	}
	ch := make(chan Event, 8)
	_ = Run(context.Background(), j, ch)
	close(ch)

	b, _ := os.ReadFile(filepath.Join(out, "photo.png"))
	if string(b) != "existing" {
		t.Errorf("PolicySkip should leave existing file untouched")
	}
}
```

- [ ] **Step 2: Run all job tests**

```bash
go test -race ./internal/job/...
```

Expected: PASS.

- [ ] **Step 3: Commit**

```bash
git add internal/job/runner_test.go
git commit -m "test(job): cover increment/overwrite/skip policies"
```

---

## Task 9: `internal/settings` — Save and Load

**Files:**
- Create: `internal/settings/settings.go`
- Create: `internal/settings/settings_test.go`

- [ ] **Step 1: Write the failing test**

```go
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
```

- [ ] **Step 2: Run, expect compile failure**

```bash
go test ./internal/settings/...
```

Expected: build error `undefined: Settings`.

- [ ] **Step 3: Implement `settings.go`**

```go
// Package settings persists GUI preferences as JSON in the OS user config dir.
package settings

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

// Settings is the persisted shape. Values are kept as plain strings/ints so
// the JSON file is human-readable and forward-compatible.
type Settings struct {
	OutputDir    string `json:"output_dir"`
	OutputMode   string `json:"output_mode"`  // "zip" | "individual"
	Overwrite    string `json:"overwrite"`    // "increment" | "overwrite" | "skip"
	Language     string `json:"language"`     // "ja" | "en" | "" (auto)
	WindowWidth  int    `json:"window_width"`
	WindowHeight int    `json:"window_height"`
}

// Defaults returns the values used on first launch and as fallback when the
// settings file is missing or corrupt.
func Defaults() Settings {
	return Settings{
		OutputDir:    "",
		OutputMode:   "zip",
		Overwrite:    "increment",
		Language:     "",
		WindowWidth:  720,
		WindowHeight: 480,
	}
}

// Path returns the OS-standard settings file path.
// macOS:   ~/Library/Application Support/img2png/settings.json
// Windows: %APPDATA%\img2png\settings.json
// Linux:   ~/.config/img2png/settings.json
func Path() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "img2png", "settings.json"), nil
}

// Load reads the settings from the OS-standard path. Missing or corrupt files
// return Defaults() with a nil error. I/O errors other than "not exist" return
// the error.
func Load() (Settings, error) {
	p, err := Path()
	if err != nil {
		return Defaults(), err
	}
	return LoadFrom(p)
}

// LoadFrom is Load with an explicit path (testable).
func LoadFrom(path string) (Settings, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return Defaults(), nil
		}
		return Defaults(), err
	}
	var s Settings
	if err := json.Unmarshal(b, &s); err != nil {
		return Defaults(), nil // corrupt → defaults, swallow
	}
	return s, nil
}

// Save writes settings to the OS-standard path, creating directories as needed.
func Save(s Settings) error {
	p, err := Path()
	if err != nil {
		return err
	}
	return SaveTo(p, s)
}

// SaveTo is Save with an explicit path (testable).
func SaveTo(path string, s Settings) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}
```

- [ ] **Step 4: Run, expect pass**

```bash
go test -race ./internal/settings/...
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/settings/settings.go internal/settings/settings_test.go
git commit -m "feat(settings): JSON persistence with defaults on missing/corrupt"
```

---

## Task 10: `internal/i18n` — translator with locale auto-detect

**Files:**
- Create: `internal/i18n/i18n.go`
- Create: `internal/i18n/i18n_test.go`

- [ ] **Step 1: Write the failing tests**

```go
// internal/i18n/i18n_test.go
package i18n

import "testing"

func TestT_KnownKey_Japanese(t *testing.T) {
	tr := New("ja")
	if got := tr.T("button.convert"); got != "変換開始" {
		t.Errorf("ja button.convert = %q, want 変換開始", got)
	}
}

func TestT_KnownKey_English(t *testing.T) {
	tr := New("en")
	if got := tr.T("button.convert"); got != "Convert" {
		t.Errorf("en button.convert = %q, want Convert", got)
	}
}

func TestT_UnknownKey_ReturnsKey(t *testing.T) {
	tr := New("en")
	if got := tr.T("nope.nope"); got != "nope.nope" {
		t.Errorf("unknown key should round-trip, got %q", got)
	}
}

func TestNew_BlankLanguage_FallsBackToEnglish(t *testing.T) {
	// "" means "no preference, use detector" — tests force fallback by passing
	// an unknown locale.
	tr := New("klingon")
	if got := tr.T("button.convert"); got != "Convert" {
		t.Errorf("unknown lang should fall back to English, got %q", got)
	}
}

func TestDetectFromLocale(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"ja_JP.UTF-8", "ja"},
		{"ja-JP", "ja"},
		{"en_US.UTF-8", "en"},
		{"fr_FR", "en"}, // unsupported → english
		{"", "en"},
	}
	for _, tc := range cases {
		if got := DetectFromLocale(tc.in); got != tc.want {
			t.Errorf("DetectFromLocale(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
```

- [ ] **Step 2: Run, expect compile failure**

```bash
go test ./internal/i18n/...
```

- [ ] **Step 3: Implement `i18n.go`**

```go
// Package i18n provides a tiny key→label translator for the GUI.
// Supported languages: Japanese ("ja"), English ("en"). English is the
// fallback for any unknown key or unknown language.
package i18n

import "strings"

// Translator looks up labels by key.
type Translator struct {
	lang  string
	table map[string]string
}

// New returns a translator for the given language code ("ja" / "en"). Any
// other value falls back to English.
func New(lang string) *Translator {
	t := &Translator{lang: "en", table: tableEN}
	if lang == "ja" {
		t.lang = "ja"
		t.table = tableJA
	}
	return t
}

// T returns the localized label for key, or key itself if no entry exists.
func (t *Translator) T(key string) string {
	if v, ok := t.table[key]; ok {
		return v
	}
	if t.lang != "en" {
		if v, ok := tableEN[key]; ok {
			return v
		}
	}
	return key
}

// Lang returns the active language code.
func (t *Translator) Lang() string { return t.lang }

// DetectFromLocale parses a POSIX/Windows locale string and returns one of
// "ja" or "en" (the only supported values).
func DetectFromLocale(loc string) string {
	loc = strings.ToLower(strings.TrimSpace(loc))
	if loc == "" {
		return "en"
	}
	// Take the language portion before any '_', '-', or '.'.
	for _, sep := range []string{".", "_", "-"} {
		if i := strings.Index(loc, sep); i >= 0 {
			loc = loc[:i]
		}
	}
	if loc == "ja" {
		return "ja"
	}
	return "en"
}

// Tables. Add new keys here in alphabetic order.
var tableEN = map[string]string{
	"button.add_files":   "+ Add Files",
	"button.add_folder":  "+ Add Folder",
	"button.cancel":      "Cancel",
	"button.clear":       "Clear",
	"button.convert":     "Convert",
	"button.open_output": "Open output folder",
	"label.drop":         "Drop files or folders here",
	"label.output_dir":   "Output folder",
	"label.output_mode":  "Folder input",
	"label.overwrite":    "On duplicate name",
	"mode.individual":    "Individual PNG",
	"mode.zip":           "ZIP",
	"policy.increment":   "Append number",
	"policy.overwrite":   "Overwrite",
	"policy.skip":        "Skip",
	"status.done":        "Done",
	"status.failed":      "Failed",
	"status.pending":     "Pending",
	"status.running":     "Running",
	"status.skipped":     "Skipped",
	"summary.canceled":   "Canceled",
	"summary.results":    "Done: %d   Skipped: %d   Failed: %d",
	"window.title":       "img2png",
}

var tableJA = map[string]string{
	"button.add_files":   "＋ ファイル追加",
	"button.add_folder":  "＋ フォルダ追加",
	"button.cancel":      "キャンセル",
	"button.clear":       "クリア",
	"button.convert":     "変換開始",
	"button.open_output": "出力フォルダを開く",
	"label.drop":         "ファイル / フォルダをここにドロップ",
	"label.output_dir":   "出力先",
	"label.output_mode":  "フォルダ入力時",
	"label.overwrite":    "同名ファイル時",
	"mode.individual":    "個別PNG",
	"mode.zip":           "ZIPにまとめる",
	"policy.increment":   "連番付与",
	"policy.overwrite":   "上書き",
	"policy.skip":        "スキップ",
	"status.done":        "完了",
	"status.failed":      "失敗",
	"status.pending":     "待機中",
	"status.running":     "処理中",
	"status.skipped":     "スキップ",
	"summary.canceled":   "キャンセルされました",
	"summary.results":    "成功 %d / スキップ %d / 失敗 %d",
	"window.title":       "img2png",
}
```

- [ ] **Step 4: Run, expect pass**

```bash
go test -race ./internal/i18n/...
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/i18n/i18n.go internal/i18n/i18n_test.go
git commit -m "feat(i18n): minimal JA/EN translator with locale detect"
```

---

## Task 11: `internal/gui/filelist` — pure helpers (validate, dedupe, remove)

**Files:**
- Create: `internal/gui/filelist.go`
- Create: `internal/gui/filelist_test.go`

This task only adds the pure functions used by the widget. The Fyne widget itself is added in Task 14 once all helpers and panels exist.

- [ ] **Step 1: Write the failing tests**

```go
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
```

- [ ] **Step 2: Run, expect compile failure**

```bash
go test ./internal/gui/...
```

- [ ] **Step 3: Implement `filelist.go`**

```go
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
```

- [ ] **Step 4: Run, expect pass**

```bash
go test -race ./internal/gui/...
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/gui/filelist.go internal/gui/filelist_test.go
git commit -m "feat(gui): file list helpers (add with validation, remove)"
```

---

## Task 12: `internal/gui/settings_panel` — `BuildJob` pure helper

**Files:**
- Create: `internal/gui/settings_panel.go`
- Create: `internal/gui/settings_panel_test.go`

- [ ] **Step 1: Write the failing tests**

```go
// internal/gui/settings_panel_test.go
package gui

import (
	"testing"

	"github.com/SilentMalachite/img2png/internal/job"
)

func TestBuildJob_AllFieldsCopied(t *testing.T) {
	items := []job.FileItem{{Path: "/a", IsDir: false}}
	state := PanelState{
		OutputDir:  "/tmp/out",
		OutputMode: "zip",
		Overwrite:  "skip",
	}
	got := BuildJob(items, state)

	if len(got.Items) != 1 || got.Items[0].Path != "/a" {
		t.Errorf("Items = %+v", got.Items)
	}
	if got.OutputDir != "/tmp/out" {
		t.Errorf("OutputDir = %q", got.OutputDir)
	}
	if got.OutputMode != job.ModeZip {
		t.Errorf("OutputMode = %v, want ModeZip", got.OutputMode)
	}
	if got.Overwrite != job.PolicySkip {
		t.Errorf("Overwrite = %v, want PolicySkip", got.Overwrite)
	}
}

func TestBuildJob_UnknownStrings_FallToDefaults(t *testing.T) {
	got := BuildJob(nil, PanelState{
		OutputMode: "garbage",
		Overwrite:  "garbage",
	})
	if got.OutputMode != job.ModeZip {
		t.Errorf("default OutputMode should be ModeZip")
	}
	if got.Overwrite != job.PolicyIncrement {
		t.Errorf("default Overwrite should be PolicyIncrement")
	}
}
```

- [ ] **Step 2: Run, expect compile failure**

- [ ] **Step 3: Implement the helper**

```go
// internal/gui/settings_panel.go
package gui

import "github.com/SilentMalachite/img2png/internal/job"

// PanelState is the plain-data view of the settings panel, decoupled from
// Fyne widgets so it can be tested.
type PanelState struct {
	OutputDir  string
	OutputMode string // "zip" | "individual"
	Overwrite  string // "increment" | "overwrite" | "skip"
}

// BuildJob constructs a job.Job from the file list and panel state. Unknown
// or missing string values map to defaults (ModeZip, PolicyIncrement).
func BuildJob(items []job.FileItem, p PanelState) job.Job {
	mode := job.ModeZip
	if p.OutputMode == "individual" {
		mode = job.ModeIndividual
	}
	pol := job.PolicyIncrement
	switch p.Overwrite {
	case "overwrite":
		pol = job.PolicyOverwrite
	case "skip":
		pol = job.PolicySkip
	}
	return job.Job{
		Items:      items,
		OutputDir:  p.OutputDir,
		OutputMode: mode,
		Overwrite:  pol,
	}
}
```

- [ ] **Step 4: Run, expect pass**

```bash
go test -race ./internal/gui/...
```

- [ ] **Step 5: Commit**

```bash
git add internal/gui/settings_panel.go internal/gui/settings_panel_test.go
git commit -m "feat(gui): BuildJob — panel state to job.Job"
```

---

## Task 13: GUI widgets — file list, settings panel, progress, menu

**Files:**
- Modify: `internal/gui/filelist.go` (append widget code)
- Modify: `internal/gui/settings_panel.go` (append widget code)
- Create: `internal/gui/progress.go`
- Create: `internal/gui/menu.go`

These are integration code that calls Fyne. They are not unit-tested (per the spec's testing strategy: GUI logic is tested via the pure helpers above).

- [ ] **Step 1: Replace `internal/gui/filelist.go` with the widget-aware version**

Add the widget code below the existing pure helpers. Here is the **complete new content** of `internal/gui/filelist.go` (the helpers `supportedExt`, `AddItems`, `RemoveAt` are unchanged from Task 11; the new code is the `FileListWidget` type):

```go
package gui

import (
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/SilentMalachite/img2png/internal/i18n"
	"github.com/SilentMalachite/img2png/internal/job"
)

func supportedExt(ext string) bool {
	switch strings.ToLower(ext) {
	case ".tif", ".tiff", ".jpg", ".jpeg", ".webp", ".bmp", ".gif", ".png":
		return true
	}
	return false
}

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

func RemoveAt(items []job.FileItem, i int) []job.FileItem {
	if i < 0 || i >= len(items) {
		return items
	}
	out := make([]job.FileItem, 0, len(items)-1)
	out = append(out, items[:i]...)
	out = append(out, items[i+1:]...)
	return out
}

// FileListWidget binds a slice of FileItem to a fyne widget.List, with a
// drop hint label above and add/clear buttons below. The owner reads/writes
// items via Items()/SetItems().
type FileListWidget struct {
	tr        *i18n.Translator
	items     []job.FileItem
	list      *widget.List
	dropHint  *widget.Label
	container *fyne.Container
	onChange  func()
}

// NewFileListWidget creates a new file list widget bound to tr (for labels)
// and onChange (called whenever Items mutates).
func NewFileListWidget(tr *i18n.Translator, onChange func()) *FileListWidget {
	w := &FileListWidget{tr: tr, onChange: onChange}
	w.dropHint = widget.NewLabel(tr.T("label.drop"))
	w.list = widget.NewList(
		func() int { return len(w.items) },
		func() fyne.CanvasObject {
			return container.NewBorder(nil, nil, nil,
				widget.NewButton("×", nil),
				widget.NewLabel(""),
			)
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			if i < 0 || i >= len(w.items) {
				return
			}
			row := o.(*fyne.Container)
			lbl := row.Objects[0].(*widget.Label)
			btn := row.Objects[1].(*widget.Button)
			it := w.items[i]
			marker := "📷 "
			if it.IsDir {
				marker = "📁 "
			}
			lbl.SetText(marker + filepath.Base(it.Path))
			idx := i
			btn.OnTapped = func() {
				w.items = RemoveAt(w.items, idx)
				w.list.Refresh()
				if w.onChange != nil {
					w.onChange()
				}
			}
		},
	)
	w.container = container.NewBorder(w.dropHint, nil, nil, nil, w.list)
	return w
}

// CanvasObject returns the root container for embedding into a parent layout.
func (w *FileListWidget) CanvasObject() fyne.CanvasObject { return w.container }

// Items returns the current items (read-only — do not mutate).
func (w *FileListWidget) Items() []job.FileItem { return w.items }

// AddPaths appends paths via AddItems and refreshes the list.
func (w *FileListWidget) AddPaths(paths []string) {
	w.items = AddItems(w.items, paths)
	w.list.Refresh()
	if w.onChange != nil {
		w.onChange()
	}
}

// Clear removes all items.
func (w *FileListWidget) Clear() {
	w.items = nil
	w.list.Refresh()
	if w.onChange != nil {
		w.onChange()
	}
}

// SetTranslator updates the language; call list.Refresh after to redraw labels.
func (w *FileListWidget) SetTranslator(tr *i18n.Translator) {
	w.tr = tr
	w.dropHint.SetText(tr.T("label.drop"))
}
```

- [ ] **Step 2: Replace `internal/gui/settings_panel.go` with widget-aware version**

```go
package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/SilentMalachite/img2png/internal/i18n"
	"github.com/SilentMalachite/img2png/internal/job"
)

type PanelState struct {
	OutputDir  string
	OutputMode string
	Overwrite  string
}

func BuildJob(items []job.FileItem, p PanelState) job.Job {
	mode := job.ModeZip
	if p.OutputMode == "individual" {
		mode = job.ModeIndividual
	}
	pol := job.PolicyIncrement
	switch p.Overwrite {
	case "overwrite":
		pol = job.PolicyOverwrite
	case "skip":
		pol = job.PolicySkip
	}
	return job.Job{
		Items:      items,
		OutputDir:  p.OutputDir,
		OutputMode: mode,
		Overwrite:  pol,
	}
}

// SettingsPanel is the right-pane widget exposing OutputDir / OutputMode /
// Overwrite as PanelState.
type SettingsPanel struct {
	tr        *i18n.Translator
	outputDir *widget.Entry
	mode      *widget.RadioGroup
	overwrite *widget.Select
	container *fyne.Container
}

func NewSettingsPanel(tr *i18n.Translator, init PanelState) *SettingsPanel {
	p := &SettingsPanel{tr: tr}
	p.outputDir = widget.NewEntry()
	p.outputDir.SetText(init.OutputDir)

	modeOptions := []string{tr.T("mode.zip"), tr.T("mode.individual")}
	p.mode = widget.NewRadioGroup(modeOptions, nil)
	if init.OutputMode == "individual" {
		p.mode.SetSelected(modeOptions[1])
	} else {
		p.mode.SetSelected(modeOptions[0])
	}

	policyOptions := []string{tr.T("policy.increment"), tr.T("policy.overwrite"), tr.T("policy.skip")}
	p.overwrite = widget.NewSelect(policyOptions, nil)
	switch init.Overwrite {
	case "overwrite":
		p.overwrite.SetSelected(policyOptions[1])
	case "skip":
		p.overwrite.SetSelected(policyOptions[2])
	default:
		p.overwrite.SetSelected(policyOptions[0])
	}

	p.container = container.NewVBox(
		widget.NewLabel(tr.T("label.output_dir")),
		p.outputDir,
		widget.NewLabel(tr.T("label.output_mode")),
		p.mode,
		widget.NewLabel(tr.T("label.overwrite")),
		p.overwrite,
	)
	return p
}

func (p *SettingsPanel) CanvasObject() fyne.CanvasObject { return p.container }

// State returns the current panel values as a PanelState.
func (p *SettingsPanel) State() PanelState {
	mode := "zip"
	if p.mode.Selected == p.tr.T("mode.individual") {
		mode = "individual"
	}
	overwrite := "increment"
	switch p.overwrite.Selected {
	case p.tr.T("policy.overwrite"):
		overwrite = "overwrite"
	case p.tr.T("policy.skip"):
		overwrite = "skip"
	}
	return PanelState{
		OutputDir:  p.outputDir.Text,
		OutputMode: mode,
		Overwrite:  overwrite,
	}
}
```

- [ ] **Step 3: Create `internal/gui/progress.go`**

```go
package gui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/SilentMalachite/img2png/internal/i18n"
	"github.com/SilentMalachite/img2png/internal/job"
)

// ProgressArea shows a progress bar, current-file label, cancel button, and
// final summary. Hidden until ShowFor(total) is called.
type ProgressArea struct {
	tr      *i18n.Translator
	bar     *widget.ProgressBar
	current *widget.Label
	summary *widget.Label
	cancel  *widget.Button
	open    *widget.Button

	container *fyne.Container

	onCancel  func()
	openPath  string
}

func NewProgressArea(tr *i18n.Translator, onCancel func()) *ProgressArea {
	p := &ProgressArea{tr: tr, onCancel: onCancel}
	p.bar = widget.NewProgressBar()
	p.current = widget.NewLabel("")
	p.summary = widget.NewLabel("")
	p.cancel = widget.NewButton(tr.T("button.cancel"), func() {
		if p.onCancel != nil {
			p.onCancel()
		}
	})
	p.open = widget.NewButton(tr.T("button.open_output"), func() {
		if p.openPath != "" {
			openInFileBrowser(p.openPath)
		}
	})
	p.open.Hide()
	p.container = container.NewVBox(p.bar, p.current, container.NewHBox(p.cancel, p.open), p.summary)
	p.container.Hide()
	return p
}

func (p *ProgressArea) CanvasObject() fyne.CanvasObject { return p.container }

// ShowFor resets state and shows the area for a job of size total.
func (p *ProgressArea) ShowFor(total int) {
	p.bar.Min = 0
	p.bar.Max = float64(total)
	p.bar.SetValue(0)
	p.current.SetText("")
	p.summary.SetText("")
	p.cancel.Enable()
	p.open.Hide()
	p.openPath = ""
	p.container.Show()
}

// Update handles a progress event from the runner. Must be called on the Fyne
// main thread.
func (p *ProgressArea) Update(ev job.Event) {
	switch ev.Kind {
	case job.EventStart:
		p.bar.Max = float64(ev.Total)
	case job.EventItem:
		p.bar.SetValue(float64(ev.Completed))
		p.current.SetText(ev.SourcePath)
	case job.EventDone:
		p.bar.SetValue(p.bar.Max)
		p.current.SetText("")
	}
}

// Finish renders the final summary line and exposes the "open output" button.
func (p *ProgressArea) Finish(done, skipped, failed int, openPath string, canceled bool) {
	p.cancel.Disable()
	if canceled {
		p.summary.SetText(p.tr.T("summary.canceled"))
	} else {
		p.summary.SetText(fmt.Sprintf(p.tr.T("summary.results"), done, skipped, failed))
	}
	if openPath != "" {
		p.openPath = openPath
		p.open.Show()
	}
}

// Hide collapses the area (e.g., when starting a fresh job).
func (p *ProgressArea) Hide() { p.container.Hide() }
```

- [ ] **Step 4: Create `internal/gui/menu.go` and `open_path.go` for the cross-platform "open folder"**

```go
// internal/gui/menu.go
package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/SilentMalachite/img2png/internal/i18n"
)

// BuildMainMenu returns a *fyne.MainMenu with Language and Help submenus.
// onLangChange is called with "ja", "en", or "" (auto/system). onAbout is
// called when the user selects About.
func BuildMainMenu(tr *i18n.Translator, onLangChange func(code string), onAbout func()) *fyne.MainMenu {
	lang := fyne.NewMenu("Language",
		fyne.NewMenuItem("Auto", func() { onLangChange("") }),
		fyne.NewMenuItem("English", func() { onLangChange("en") }),
		fyne.NewMenuItem("日本語", func() { onLangChange("ja") }),
	)
	help := fyne.NewMenu("Help",
		fyne.NewMenuItem("About img2png", onAbout),
	)
	return fyne.NewMainMenu(lang, help)
}

// MenuBarFallback is for Windows where users may expect an in-window control.
// Returns a small dropdown button that mirrors the main menu language items.
func MenuBarFallback(tr *i18n.Translator, onLangChange func(code string)) fyne.CanvasObject {
	btn := widget.NewButton("⋮", nil)
	btn.OnTapped = func() {
		// Simplest UX: cycle Auto → en → ja → Auto.
		// Real submenu handling is left to the native menu in BuildMainMenu.
		onLangChange("ja")
	}
	return container.NewHBox(btn)
}
```

```go
// internal/gui/open_path.go
package gui

import (
	"os/exec"
	"runtime"
)

// openInFileBrowser reveals the given file or folder in the OS file manager.
// macOS: open -R (file) / open (dir). Windows: explorer /select. Linux: xdg-open.
// Errors are intentionally swallowed — this is a convenience action.
func openInFileBrowser(path string) {
	switch runtime.GOOS {
	case "darwin":
		_ = exec.Command("open", "-R", path).Start()
	case "windows":
		_ = exec.Command("explorer", "/select,", path).Start()
	default:
		_ = exec.Command("xdg-open", path).Start()
	}
}
```

- [ ] **Step 5: Build to verify everything compiles together**

```bash
go build ./...
```

Expected: no output, exit 0.

- [ ] **Step 6: Re-run all tests (pure helpers must still pass)**

```bash
go test -race ./...
```

Expected: PASS for converter, archiver, job, settings, i18n, gui. New widget code has no tests, but `go vet ./...` and `go build` confirm shape.

- [ ] **Step 7: Commit**

```bash
git add internal/gui/filelist.go internal/gui/settings_panel.go internal/gui/progress.go internal/gui/menu.go internal/gui/open_path.go
git commit -m "feat(gui): add file list, settings, progress, and menu widgets"
```

---

## Task 14: GUI app entry — assemble main window in `internal/gui/app.go`

**Files:**
- Create: `internal/gui/app.go`

- [ ] **Step 1: Write `app.go`**

```go
package gui

import (
	"context"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/SilentMalachite/img2png/internal/i18n"
	"github.com/SilentMalachite/img2png/internal/job"
	"github.com/SilentMalachite/img2png/internal/settings"
)

// Run starts the GUI event loop. Blocks until the window closes. The function
// loads persisted settings, builds the window, and saves settings on exit.
func Run() error {
	cfg, _ := settings.Load()
	lang := cfg.Language
	if lang == "" {
		lang = i18n.DetectFromLocale(os.Getenv("LANG"))
	}
	tr := i18n.New(lang)

	a := app.NewWithID("com.silentmalachite.img2png")
	w := a.NewWindow(tr.T("window.title"))
	w.Resize(fyne.NewSize(float32(cfg.WindowWidth), float32(cfg.WindowHeight)))

	// Left pane.
	fileList := NewFileListWidget(tr, nil)
	addFilesBtn := widget.NewButton(tr.T("button.add_files"), func() {
		dialog.ShowFileOpen(func(rc fyne.URIReadCloser, err error) {
			if err != nil || rc == nil {
				return
			}
			defer rc.Close()
			fileList.AddPaths([]string{rc.URI().Path()})
		}, w)
	})
	addFolderBtn := widget.NewButton(tr.T("button.add_folder"), func() {
		dialog.ShowFolderOpen(func(u fyne.ListableURI, err error) {
			if err != nil || u == nil {
				return
			}
			fileList.AddPaths([]string{u.Path()})
		}, w)
	})
	clearBtn := widget.NewButton(tr.T("button.clear"), fileList.Clear)
	leftBottom := container.NewHBox(addFilesBtn, addFolderBtn, clearBtn)
	left := container.NewBorder(nil, leftBottom, nil, nil, fileList.CanvasObject())

	// Right pane.
	panel := NewSettingsPanel(tr, PanelState{
		OutputDir:  cfg.OutputDir,
		OutputMode: cfg.OutputMode,
		Overwrite:  cfg.Overwrite,
	})
	convertBtn := widget.NewButton(tr.T("button.convert"), nil)
	right := container.NewBorder(panel.CanvasObject(), convertBtn, nil, nil)

	// Progress area (bottom of left pane).
	var cancelFn context.CancelFunc
	progress := NewProgressArea(tr, func() {
		if cancelFn != nil {
			cancelFn()
		}
	})

	leftWithProgress := container.NewBorder(nil, progress.CanvasObject(), nil, nil, left)
	split := container.NewHSplit(leftWithProgress, right)
	split.SetOffset(0.6)
	w.SetContent(split)

	// Convert button wiring — defined after widgets exist.
	convertBtn.OnTapped = func() {
		items := fileList.Items()
		if len(items) == 0 {
			return
		}
		j := BuildJob(items, panel.State())

		// Pre-flight: confirm we can write to the output directory before
		// starting the job. This matches the spec's "scenario D" — show a
		// modal and do not start when the destination is unwritable.
		if dest := preflightOutputDir(j); dest != "" {
			if err := checkWritable(dest); err != nil {
				dialog.ShowError(err, w)
				return
			}
		}

		ctx, cancel := context.WithCancel(context.Background())
		cancelFn = cancel
		progress.ShowFor(estimateTotal(items))

		ch := make(chan job.Event, 32)
		go func() {
			_ = job.Run(ctx, j, ch)
			close(ch)
		}()
		go func() {
			done, skipped, failed := 0, 0, 0
			var openPath string
			for ev := range ch {
				ev := ev
				fyne.Do(func() { progress.Update(ev) })
				switch ev.Kind {
				case job.EventItem:
					switch ev.Status {
					case job.StatusDone:
						done++
					case job.StatusSkipped:
						skipped++
					case job.StatusFailed:
						failed++
					}
				case job.EventDone:
					if ev.OutputPath != "" {
						openPath = ev.OutputPath
					}
				}
			}
			canceled := ctx.Err() != nil
			fyne.Do(func() {
				progress.Finish(done, skipped, failed, openPath, canceled)
			})
		}()
	}

	// Drag-and-drop into the window.
	w.SetOnDropped(func(_ fyne.Position, uris []fyne.URI) {
		paths := make([]string, 0, len(uris))
		for _, u := range uris {
			paths = append(paths, u.Path())
		}
		fileList.AddPaths(paths)
	})

	// Menu (language switch + About).
	w.SetMainMenu(BuildMainMenu(tr,
		func(code string) {
			cfg.Language = code
			_ = settings.Save(cfg)
			dialog.ShowInformation("Restart required", "Restart the app to apply the language change.", w)
		},
		func() {
			dialog.ShowInformation("About img2png",
				"img2png — convert JPEG/TIFF/WebP/BMP/GIF to PNG.\nhttps://github.com/SilentMalachite/img2png",
				w)
		},
	))

	// Persist window size and panel state on close.
	w.SetCloseIntercept(func() {
		size := w.Canvas().Size()
		cfg.WindowWidth = int(size.Width)
		cfg.WindowHeight = int(size.Height)
		st := panel.State()
		cfg.OutputDir = st.OutputDir
		cfg.OutputMode = st.OutputMode
		cfg.Overwrite = st.Overwrite
		_ = settings.Save(cfg)
		w.Close()
	})

	w.ShowAndRun()
	return nil
}

// estimateTotal returns the number of converted files we expect for `items`.
// For directories, we walk and count supported entries up front so the
// progress bar has a real maximum.
func estimateTotal(items []job.FileItem) int {
	total := 0
	for _, it := range items {
		if !it.IsDir {
			total++
			continue
		}
		_ = filepath.Walk(it.Path, func(p string, info os.FileInfo, err error) error {
			if err != nil || info == nil || info.IsDir() {
				return nil
			}
			if supportedExt(filepath.Ext(p)) {
				total++
			}
			return nil
		})
	}
	if total == 0 {
		total = 1
	}
	return total
}

// preflightOutputDir picks the directory the job will write its first output
// to: explicit OutputDir if set, otherwise the parent dir of the first item.
// Returns "" when nothing is writable-checkable up front (e.g., empty list).
func preflightOutputDir(j job.Job) string {
	if j.OutputDir != "" {
		return j.OutputDir
	}
	for _, it := range j.Items {
		if it.IsDir {
			return filepath.Dir(it.Path)
		}
		return filepath.Dir(it.Path)
	}
	return ""
}

// checkWritable verifies write permission by creating and removing a temp file.
// Mirrors the helper in main.go so the GUI can refuse to start a doomed job.
func checkWritable(dir string) error {
	tmp, err := os.CreateTemp(dir, ".img2png_check_*")
	if err != nil {
		return err
	}
	tmp.Close()
	os.Remove(tmp.Name())
	return nil
}
```

- [ ] **Step 2: Build**

```bash
go build ./...
```

Expected: success.

- [ ] **Step 3: Run all tests**

```bash
go test -race ./...
```

Expected: all PASS.

- [ ] **Step 4: Commit**

```bash
git add internal/gui/app.go
git commit -m "feat(gui): assemble main window with D&D, progress, and menu"
```

---

## Task 15: Wire main.go dispatch and `--help` flag

**Files:**
- Modify: `main.go`

- [ ] **Step 1: Replace `main.go` with the dispatching version**

```go
package main

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/SilentMalachite/img2png/internal/archiver"
	"github.com/SilentMalachite/img2png/internal/converter"
	"github.com/SilentMalachite/img2png/internal/gui"
)

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		// No args → GUI mode.
		if err := gui.Run(); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
		return
	}

	if len(args) == 1 {
		switch args[0] {
		case "-h", "--help":
			printHelp()
			return
		}
		if err := run(args[0]); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			pause()
			os.Exit(1)
		}
		pause()
		return
	}

	fmt.Fprintln(os.Stderr, "Usage: img2png <file or folder>")
	fmt.Fprintln(os.Stderr, "Run img2png without arguments to open the GUI.")
	pause()
	os.Exit(1)
}

func printHelp() {
	fmt.Println(`img2png — convert images to PNG.

Usage:
  img2png                    Open the GUI.
  img2png <file>             Convert a single file; PNG is written next to it.
  img2png <folder>           Convert every supported image in <folder>;
                             a <folder>.zip is written next to it.
  img2png -h | --help        Show this help.

Supported inputs: .tif .tiff .jpg .jpeg .webp .bmp .gif`)
}

func run(path string) error {
	path = filepath.Clean(path)

	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("cannot access: %s", path)
	}

	if info.IsDir() {
		return runDir(path)
	}
	return runFile(path)
}

func runFile(path string) error {
	if !supportedExt(filepath.Ext(path)) {
		return fmt.Errorf("unsupported format: %s", filepath.Base(path))
	}

	outDir := filepath.Dir(path)
	if err := checkWritable(outDir); err != nil {
		return err
	}

	base := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	outName := base + ".png"

	outPath, err := converter.ConvertFile(path, outDir, outName)
	if err != nil {
		return err
	}

	fmt.Println("Done:", filepath.Base(outPath))
	return nil
}

func runDir(dirPath string) error {
	parentDir := filepath.Dir(dirPath)
	if err := checkWritable(parentDir); err != nil {
		return err
	}

	var files []string
	_ = filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() && supportedExt(filepath.Ext(path)) {
			files = append(files, path)
		}
		return nil
	})

	if len(files) == 0 {
		return fmt.Errorf("no supported image files found")
	}

	tmpDir, err := os.MkdirTemp("", "img2png_*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	seen := make(map[string]int)
	var pngFiles []string
	for _, src := range files {
		base := strings.TrimSuffix(filepath.Base(src), filepath.Ext(src)) + ".png"
		outName := archiver.DedupeFilename(base, seen)
		seen[base]++

		pngPath, err := converter.ConvertFile(src, tmpDir, outName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "skipped: %s (unsupported or broken)\n", filepath.Base(src))
			continue
		}
		pngFiles = append(pngFiles, pngPath)
	}

	if len(pngFiles) == 0 {
		return fmt.Errorf("all files failed to convert")
	}

	dirName := filepath.Base(dirPath)
	zipPath := filepath.Join(parentDir, dirName+".zip")

	if err := archiver.Archive(pngFiles, zipPath); err != nil {
		return fmt.Errorf("failed to create ZIP: %w", err)
	}

	fmt.Println("Done:", filepath.Base(zipPath))
	return nil
}

func supportedExt(ext string) bool {
	switch strings.ToLower(ext) {
	case ".tif", ".tiff", ".jpg", ".jpeg", ".webp", ".bmp", ".gif":
		return true
	}
	return false
}

func checkWritable(dir string) error {
	tmp, err := os.CreateTemp(dir, ".img2png_check_*")
	if err != nil {
		return fmt.Errorf("cannot write to directory: %s", dir)
	}
	tmp.Close()
	os.Remove(tmp.Name())
	return nil
}

func pause() {
	fmt.Fprint(os.Stderr, "Press Enter to exit...")
	bufio.NewReader(os.Stdin).ReadString('\n')
}
```

- [ ] **Step 2: Build, verify CLI still works for the existing case**

```bash
go build -o /tmp/img2png-cli .
/tmp/img2png-cli -h
```

Expected: help text printed.

- [ ] **Step 3: Run all tests**

```bash
go test -race ./...
```

Expected: PASS.

- [ ] **Step 4: Commit**

```bash
git add main.go
git commit -m "feat(main): dispatch to GUI when no args, add --help"
```

---

## Task 16: Update test workflow for CGO + Linux deps

**Files:**
- Modify: `.github/workflows/test.yml`

- [ ] **Step 1: Replace the file contents**

```yaml
name: Test

on:
  push:
    branches:
      - main
  pull_request:
    branches:
      - main

jobs:
  test:
    strategy:
      fail-fast: false
      matrix:
        os: [ubuntu-latest, macos-latest, windows-latest]

    runs-on: ${{ matrix.os }}

    env:
      CGO_ENABLED: "1"

    steps:
      - uses: actions/checkout@v5

      - uses: actions/setup-go@v6
        with:
          go-version-file: go.mod
          cache: true

      - name: Install Linux GUI deps
        if: runner.os == 'Linux'
        run: |
          sudo apt-get update
          sudo apt-get install -y libgl1-mesa-dev xorg-dev

      - name: Test
        run: go test -race ./...
```

- [ ] **Step 2: Commit**

```bash
git add .github/workflows/test.yml
git commit -m "ci: enable CGO and install Linux GUI deps for tests"
```

---

## Task 17: Rewrite build workflow to use native runners

**Files:**
- Modify: `.github/workflows/build.yml`

- [ ] **Step 1: Replace the file contents**

```yaml
name: Build

on:
  push:
    tags:
      - "v*"
  workflow_dispatch:

jobs:
  build:
    strategy:
      fail-fast: false
      matrix:
        include:
          - runner: macos-14
            goos: darwin
            artifact_dir: img2png.app
            artifact_zip: img2png-darwin-arm64.zip
          - runner: macos-13
            goos: darwin
            artifact_dir: img2png.app
            artifact_zip: img2png-darwin-amd64.zip
          - runner: windows-latest
            goos: windows
            artifact_exe: img2png.exe
            artifact_zip: img2png-windows-amd64.zip
          - runner: ubuntu-latest
            goos: linux
            artifact_tar: img2png.tar.xz
            artifact_zip: img2png-linux-amd64.tar.xz

    runs-on: ${{ matrix.runner }}

    permissions:
      contents: write

    env:
      CGO_ENABLED: "1"

    steps:
      - uses: actions/checkout@v5

      - uses: actions/setup-go@v6
        with:
          go-version-file: go.mod
          cache: true

      - name: Install Linux GUI deps
        if: runner.os == 'Linux'
        run: |
          sudo apt-get update
          sudo apt-get install -y libgl1-mesa-dev xorg-dev

      - name: Test
        run: go test -race ./...

      - name: Install fyne CLI
        run: go install fyne.io/fyne/v2/cmd/fyne@latest

      - name: Package
        run: fyne package -os ${{ matrix.goos }} -name img2png -icon assets/icon.png

      - name: Zip artifact (macOS)
        if: matrix.goos == 'darwin'
        run: zip -r ${{ matrix.artifact_zip }} ${{ matrix.artifact_dir }}

      - name: Zip artifact (Windows)
        if: matrix.goos == 'windows'
        run: 7z a ${{ matrix.artifact_zip }} ${{ matrix.artifact_exe }}

      - name: Rename artifact (Linux)
        if: matrix.goos == 'linux'
        run: mv ${{ matrix.artifact_tar }} ${{ matrix.artifact_zip }}

      - name: Upload to Release
        uses: softprops/action-gh-release@v2
        with:
          files: ${{ matrix.artifact_zip }}
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

- [ ] **Step 2: Commit**

```bash
git add .github/workflows/build.yml
git commit -m "ci: native-runner build matrix using fyne package"
```

---

## Task 18: Update README

**Files:**
- Modify: `README.md`

- [ ] **Step 1: Open the file and apply these targeted edits**

Use the Edit tool with `replace_all: false` for each.

**Edit 1 — feature list (Japanese): remove the "CGO 不使用" bullet**

Find:
```
- **CGO 不使用**（macOS / Windows / Linux 対応の単一バイナリ）
- ファイルを **ドラッグ&ドロップ** するだけで使える
```

Replace with:
```
- **GUI モード対応**（macOS の `.app` バンドル / Windows の `.exe` をダブルクリック）
- ファイルを **ドラッグ&ドロップ** するだけで使える（ウィンドウ・バイナリ両対応）
```

**Edit 2 — feature list (English): same change**

Find:
```
- **No CGO** — single binary for macOS, Windows, and Linux
- Works with **drag-and-drop** — no terminal required
```

Replace with:
```
- **GUI mode** — double-click the `.app` (macOS) or `.exe` (Windows)
- Works with **drag-and-drop** — into the window or onto the binary
```

**Edit 3 — Japanese download table**

Find:
```
| OS | ファイル |
|----|---------|
| macOS (Apple Silicon) | `img2png-darwin-arm64` |
| macOS (Intel) | `img2png-darwin-amd64` |
| Windows | `img2png-windows-amd64.exe` |
| Linux | `img2png-linux-amd64` |
```

Replace with:
```
| OS | ファイル |
|----|---------|
| macOS (Apple Silicon) | `img2png-darwin-arm64.zip` |
| macOS (Intel) | `img2png-darwin-amd64.zip` |
| Windows | `img2png-windows-amd64.zip` |
| Linux | `img2png-linux-amd64.tar.xz` |
```

**Edit 4 — English download table**

Find:
```
| OS | File |
|----|------|
| macOS (Apple Silicon) | `img2png-darwin-arm64` |
| macOS (Intel) | `img2png-darwin-amd64` |
| Windows | `img2png-windows-amd64.exe` |
| Linux | `img2png-linux-amd64` |
```

Replace with:
```
| OS | File |
|----|------|
| macOS (Apple Silicon) | `img2png-darwin-arm64.zip` |
| macOS (Intel) | `img2png-darwin-amd64.zip` |
| Windows | `img2png-windows-amd64.zip` |
| Linux | `img2png-linux-amd64.tar.xz` |
```

**Edit 5 — add GUI section in Japanese (after "### 使い方" header but before "#### ドラッグ&ドロップ")**

Find:
```
### 使い方

#### ドラッグ&ドロップ
```

Replace with:
```
### 使い方

#### GUI モード（推奨）

`img2png.app`（macOS）または `img2png.exe`（Windows）をダブルクリックするとウィンドウが開きます。
ファイル/フォルダをウィンドウにドラッグ&ドロップするか、「＋ファイル追加」「＋フォルダ追加」ボタンで追加し、出力先・出力モード・上書きポリシーを設定して「変換開始」を押してください。

設定（言語・ウィンドウサイズ・出力先など）は次回起動時に復元されます。

#### ドラッグ&ドロップ（バイナリ直接）
```

**Edit 6 — add GUI section in English (mirror of Edit 5)**

Find:
```
### Usage

#### Drag and Drop

Drag an image file or folder onto the binary to start conversion.
```

Replace with:
```
### Usage

#### GUI mode (recommended)

Double-click `img2png.app` (macOS) or `img2png.exe` (Windows) to open a window. Drag files/folders into the window or use the **+ Add Files / + Add Folder** buttons, configure the output folder, mode, and duplicate-name policy, then press **Convert**.

Settings (language, window size, output folder, etc.) are restored on next launch.

#### Drag and drop (onto the binary)

Drag an image file or folder onto the binary to start conversion in CLI mode.
```

- [ ] **Step 2: Verify the file renders cleanly**

Open the file in your editor (or `cat README.md`) and skim — every section header must still be present and tables aligned.

- [ ] **Step 3: Commit**

```bash
git add README.md
git commit -m "docs: update README for GUI mode and new release artifacts"
```

---

## Task 19: Manual smoke test on the host machine

These steps cannot be automated and must be performed by a human (or driven via your IDE). They are the v1 acceptance gate.

- [ ] **Step 1: Build for the current platform**

```bash
go install fyne.io/fyne/v2/cmd/fyne@latest
fyne package -os $(go env GOOS) -name img2png -icon assets/icon.png
```

Expected on macOS: `img2png.app` directory in the working tree. On Windows: `img2png.exe`. On Linux: `img2png.tar.xz`.

- [ ] **Step 2: Launch the GUI**

- macOS: `open img2png.app`
- Windows: double-click `img2png.exe`
- Linux: `tar xf img2png.tar.xz && ./usr/local/bin/img2png`

Expected: window opens with title "img2png", left list empty, right panel showing output dir / mode / overwrite controls, Convert button.

- [ ] **Step 3: Drag and drop a folder containing images into the window**

Expected: each image's parent (or the dropped folder itself) appears in the list.

- [ ] **Step 4: Press Convert (with default settings — ZIP / Increment)**

Expected: progress bar advances, current-file label updates, summary shows `Done: N / Skipped: 0 / Failed: 0` and an "Open output folder" button.

- [ ] **Step 5: Verify the output**

Expected: a `<folder>.zip` next to the source folder, containing PNGs.

- [ ] **Step 6: Run a job, press Cancel partway through**

Expected: progress halts, summary shows "Canceled". Some PNGs are present in the staging temp dir's eventual ZIP (or none if cancel was very fast).

- [ ] **Step 7: Switch language via the menu, restart**

Expected: dialog asks to restart. After relaunch, all labels are in the chosen language.

- [ ] **Step 8: Resize the window, close, relaunch**

Expected: the window opens at the previous size.

- [ ] **Step 9: Verify CLI mode still works**

```bash
./img2png path/to/photo.jpg
./img2png path/to/folder
./img2png --help
```

Expected: each behaves exactly as before this feature, except `--help` now exists. Output filenames identical to pre-GUI builds.

- [ ] **Step 10: Tag the smoke-test pass**

If everything checks out:

```bash
git tag -a smoke-pass-$(date +%Y%m%d) -m "GUI smoke test passed on $(uname -s)"
```

(This is a local marker, not a release tag.)
