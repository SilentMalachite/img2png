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
