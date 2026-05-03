// internal/i18n/i18n_test.go
package i18n

import "testing"

func TestT_KnownKey_Japanese(t *testing.T) {
	tr := New("ja")
	if got := tr.T("button.convert"); got != "変換開始" {
		t.Errorf("ja button.convert = %q, want 変換開始", got)
	}
}

func TestT_KnownKey_English(t *testing.T) {
	tr := New("en")
	if got := tr.T("button.convert"); got != "Convert" {
		t.Errorf("en button.convert = %q, want Convert", got)
	}
}

func TestT_UnknownKey_ReturnsKey(t *testing.T) {
	tr := New("en")
	if got := tr.T("nope.nope"); got != "nope.nope" {
		t.Errorf("unknown key should round-trip, got %q", got)
	}
}

func TestNew_BlankLanguage_FallsBackToEnglish(t *testing.T) {
	// "" means "no preference, use detector" — tests force fallback by passing
	// an unknown locale.
	tr := New("klingon")
	if got := tr.T("button.convert"); got != "Convert" {
		t.Errorf("unknown lang should fall back to English, got %q", got)
	}
}

func TestDetectFromLocale(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"ja_JP.UTF-8", "ja"},
		{"ja-JP", "ja"},
		{"en_US.UTF-8", "en"},
		{"fr_FR", "en"}, // unsupported → english
		{"", "en"},
	}
	for _, tc := range cases {
		if got := DetectFromLocale(tc.in); got != tc.want {
			t.Errorf("DetectFromLocale(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
