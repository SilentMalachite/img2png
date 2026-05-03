// Package i18n provides a tiny key→label translator for the GUI.
// Supported languages: Japanese ("ja"), English ("en"). English is the
// fallback for any unknown key or unknown language.
package i18n

import "strings"

// Translator looks up labels by key.
type Translator struct {
	lang  string
	table map[string]string
}

// New returns a translator for the given language code ("ja" / "en"). Any
// other value falls back to English.
func New(lang string) *Translator {
	t := &Translator{lang: "en", table: tableEN}
	if lang == "ja" {
		t.lang = "ja"
		t.table = tableJA
	}
	return t
}

// T returns the localized label for key, or key itself if no entry exists.
func (t *Translator) T(key string) string {
	if v, ok := t.table[key]; ok {
		return v
	}
	if t.lang != "en" {
		if v, ok := tableEN[key]; ok {
			return v
		}
	}
	return key
}

// Lang returns the active language code.
func (t *Translator) Lang() string { return t.lang }

// DetectFromLocale parses a POSIX/Windows locale string and returns one of
// "ja" or "en" (the only supported values).
func DetectFromLocale(loc string) string {
	loc = strings.ToLower(strings.TrimSpace(loc))
	if loc == "" {
		return "en"
	}
	// Take the language portion before any '_', '-', or '.'.
	for _, sep := range []string{".", "_", "-"} {
		if i := strings.Index(loc, sep); i >= 0 {
			loc = loc[:i]
		}
	}
	if loc == "ja" {
		return "ja"
	}
	return "en"
}

// Tables. Add new keys here in alphabetic order.
var tableEN = map[string]string{
	"button.add_files":   "+ Add Files",
	"button.add_folder":  "+ Add Folder",
	"button.cancel":      "Cancel",
	"button.clear":       "Clear",
	"button.convert":     "Convert",
	"button.open_output": "Open output folder",
	"label.drop":         "Drop files or folders here",
	"label.output_dir":   "Output folder",
	"label.output_mode":  "Folder input",
	"label.overwrite":    "On duplicate name",
	"mode.individual":    "Individual PNG",
	"mode.zip":           "ZIP",
	"policy.increment":   "Append number",
	"policy.overwrite":   "Overwrite",
	"policy.skip":        "Skip",
	"status.done":        "Done",
	"status.failed":      "Failed",
	"status.pending":     "Pending",
	"status.running":     "Running",
	"status.skipped":     "Skipped",
	"summary.canceled":   "Canceled",
	"summary.results":    "Done: %d   Skipped: %d   Failed: %d",
	"window.title":       "img2png",
}

var tableJA = map[string]string{
	"button.add_files":   "＋ ファイル追加",
	"button.add_folder":  "＋ フォルダ追加",
	"button.cancel":      "キャンセル",
	"button.clear":       "クリア",
	"button.convert":     "変換開始",
	"button.open_output": "出力フォルダを開く",
	"label.drop":         "ファイル / フォルダをここにドロップ",
	"label.output_dir":   "出力先",
	"label.output_mode":  "フォルダ入力時",
	"label.overwrite":    "同名ファイル時",
	"mode.individual":    "個別PNG",
	"mode.zip":           "ZIPにまとめる",
	"policy.increment":   "連番付与",
	"policy.overwrite":   "上書き",
	"policy.skip":        "スキップ",
	"status.done":        "完了",
	"status.failed":      "失敗",
	"status.pending":     "待機中",
	"status.running":     "処理中",
	"status.skipped":     "スキップ",
	"summary.canceled":   "キャンセルされました",
	"summary.results":    "成功 %d / スキップ %d / 失敗 %d",
	"window.title":       "img2png",
}
