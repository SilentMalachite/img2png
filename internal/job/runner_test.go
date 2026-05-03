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
