package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
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

func NewSettingsPanel(tr *i18n.Translator, init PanelState, parent fyne.Window) *SettingsPanel {
	p := &SettingsPanel{tr: tr}
	p.outputDir = widget.NewEntry()
	p.outputDir.SetText(init.OutputDir)

	browse := widget.NewButton(tr.T("button.browse"), func() {
		dialog.ShowFolderOpen(func(u fyne.ListableURI, err error) {
			if err != nil || u == nil {
				return
			}
			p.outputDir.SetText(u.Path())
		}, parent)
	})
	outputDirRow := container.NewBorder(nil, nil, nil, browse, p.outputDir)

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
		outputDirRow,
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
