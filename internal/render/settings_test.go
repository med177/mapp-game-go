package render

import (
	"encoding/json"
	"testing"
)

func TestSettingsFullscreenJSONRoundTrip(t *testing.T) {
	for _, fullscreen := range []bool{false, true} {
		data, err := json.Marshal(Settings{Fullscreen: fullscreen})
		if err != nil {
			t.Fatalf("ayarlar JSON'a çevrilemedi: %v", err)
		}

		loaded := DefaultSettings()
		if err := json.Unmarshal(data, &loaded); err != nil {
			t.Fatalf("ayarlar JSON'dan okunamadı: %v", err)
		}
		if loaded.Fullscreen != fullscreen {
			t.Fatalf("fullscreen değeri korunmadı: got=%v want=%v", loaded.Fullscreen, fullscreen)
		}
	}
}

func TestLegacySettingsRemainWindowedByDefault(t *testing.T) {
	loaded := DefaultSettings()
	if err := json.Unmarshal([]byte(`{"Difficulty":1,"MusicOn":true}`), &loaded); err != nil {
		t.Fatalf("eski ayar formatı okunamadı: %v", err)
	}
	if loaded.Fullscreen {
		t.Fatal("fullscreen alanı olmayan eski ayarlar pencereli kalmalı")
	}
}

func TestDisplayModeLabels(t *testing.T) {
	if got := displayModeLabelTR(false); got != "Pencereli" {
		t.Fatalf("pencereli etiket yanlış: %q", got)
	}
	if got := displayModeLabelTR(true); got != "Tam Ekran" {
		t.Fatalf("tam ekran etiketi yanlış: %q", got)
	}
}
