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
