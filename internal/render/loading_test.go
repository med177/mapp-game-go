package render

import (
	"path/filepath"
	"testing"
)

func TestLoadingBackgroundLoadsFromScenarioDirectory(t *testing.T) {
	oldPath := loadingBackgroundPath
	oldImage := loadingBackground
	t.Cleanup(func() {
		loadingBackgroundPath = oldPath
		loadingBackground = oldImage
	})

	scenarioPath, err := filepath.Abs(filepath.Join("..", "..", "assets", "scenarios", "1300_ottoman_rise"))
	if err != nil {
		t.Fatalf("senaryo yolu çözülemedi: %v", err)
	}

	background := loadingBackgroundImage(scenarioPath)
	if background == nil {
		t.Fatal("yükleme ekranı senaryo arka planını yükleyemedi")
	}
	if gotW, gotH := background.Bounds().Dx(), background.Bounds().Dy(); gotW != 1408 || gotH != 768 {
		t.Fatalf("yükleme arka planı boyutu yanlış: got=%dx%d want=1408x768", gotW, gotH)
	}
}
