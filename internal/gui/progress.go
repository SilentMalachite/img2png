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

	onCancel func()
	openPath string
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
