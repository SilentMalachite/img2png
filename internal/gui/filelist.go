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
