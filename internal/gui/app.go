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
