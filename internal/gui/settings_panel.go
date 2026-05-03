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
