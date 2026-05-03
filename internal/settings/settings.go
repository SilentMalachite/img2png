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
