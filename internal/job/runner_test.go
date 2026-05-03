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

func TestRun_Cancel_StopsBeforeRemainingItems(t *testing.T) {
	dir := t.TempDir()
	srcs := []string{
		filepath.Join(dir, "a.png"),
		filepath.Join(dir, "b.png"),
		filepath.Join(dir, "c.png"),
	}
	for _, p := range srcs {
		writePNG(t, p)
	}

	out := filepath.Join(dir, "out")
	if err := os.MkdirAll(out, 0o755); err != nil {
		t.Fatal(err)
	}

	j := Job{
		Items: []FileItem{
			{Path: srcs[0]}, {Path: srcs[1]}, {Path: srcs[2]},
		},
		OutputDir:  out,
		OutputMode: ModeIndividual,
		Overwrite:  PolicyIncrement,
	}

	// Cancel BEFORE Run starts. The runner should observe ctx.Err() at the top
	// of its per-item loop and produce zero EventItem events.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	ch := make(chan Event, 32)
	if err := Run(ctx, j, ch); err != nil {
		t.Fatalf("Run: %v", err)
	}
	close(ch)

	events := drain(ch)
	itemCount := 0
	for _, ev := range events {
		if ev.Kind == EventItem {
			itemCount++
		}
	}
	if itemCount != 0 {
		t.Errorf("expected 0 EventItem after pre-cancel, got %d", itemCount)
	}

	// No PNGs should have been produced.
	entries, _ := os.ReadDir(out)
	if len(entries) != 0 {
		t.Errorf("expected no produced PNGs, got %d", len(entries))
	}
}

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
