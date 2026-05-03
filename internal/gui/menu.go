// internal/gui/menu.go
package gui

import (
	"fyne.io/fyne/v2"

	"github.com/SilentMalachite/img2png/internal/i18n"
)

// BuildMainMenu returns a *fyne.MainMenu with Language and Help submenus.
// onLangChange is called with "ja", "en", or "" (auto/system). onAbout is
// called when the user selects About.
func BuildMainMenu(tr *i18n.Translator, onLangChange func(code string), onAbout func()) *fyne.MainMenu {
	lang := fyne.NewMenu("Language",
		fyne.NewMenuItem("Auto", func() { onLangChange("") }),
		fyne.NewMenuItem("English", func() { onLangChange("en") }),
		fyne.NewMenuItem("日本語", func() { onLangChange("ja") }),
	)
	help := fyne.NewMenu("Help",
		fyne.NewMenuItem("About img2png", onAbout),
	)
	return fyne.NewMainMenu(lang, help)
}
