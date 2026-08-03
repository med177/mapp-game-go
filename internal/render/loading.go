package render

import (
	"image/color"
	"math"
	"path/filepath"
	"time"

	gameui "mapp-game-go/internal/ui"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

var (
	loadingBackgroundPath string
	loadingBackground     *ebiten.Image
)

func DrawLoadingScreen(screen *ebiten.Image, message string, progress int, tick int, scenarioPath string) {
	if message == "" {
		message = "Yükleniyor..."
	}
	if background := loadingBackgroundImage(scenarioPath); background != nil {
		drawUIImageCover(screen, background)
		drawUIOverlay(screen, color.RGBA{0, 0, 0, 92})
	} else {
		screen.Fill(color.RGBA{7, 8, 12, 255})
	}
	vector.FillRect(screen, 0, 0, float32(ScreenWidth), 3, color.RGBA{180, 150, 60, 200}, false)
	vector.FillRect(screen, 0, float32(ScreenHeight)-3, float32(ScreenWidth), 3, color.RGBA{180, 150, 60, 200}, false)

	cx := float32(ScreenWidth / 2)
	// Yükleme grubunun tamamı ekranın alt-orta bölümünde tutulur.
	cy := float32(ScreenHeight - 150)
	radius := float32(28)
	phase := float64(time.Now().UnixNano()%int64(60*time.Second)) / float64(time.Second) * 4.5
	for i := 0; i < 12; i++ {
		angle := float64(i)/12*math.Pi*2 + phase + float64(tick)*0.01
		alpha := uint8(60 + i*14)
		x := cx + float32(math.Cos(angle))*radius
		y := cy + float32(math.Sin(angle))*radius
		vector.FillCircle(screen, x, y, 3.6, color.RGBA{220, 180, 70, alpha}, true)
	}

	drawUILabel(screen, gameui.Rect{X: 0, Y: float64(cy) + 46, W: ScreenWidth}, message, ColorYellow, gameui.TextLarge, gameui.TextAlignCenter)
	barW := float32(280)
	barH := float32(12)
	barX := cx - barW/2
	barY := cy + 68
	vector.FillRect(screen, barX, barY, barW, barH, color.RGBA{28, 24, 18, 220}, false)
	vector.StrokeRect(screen, barX, barY, barW, barH, 1, color.RGBA{130, 110, 60, 220}, false)
	fillW := barW * float32(progress) / 100
	if fillW > 0 {
		vector.FillRect(screen, barX+1, barY+1, maxFloat32(fillW-2, 0), barH-2, color.RGBA{214, 176, 68, 235}, false)
	}

	drawUILabel(screen, gameui.Rect{X: 0, Y: float64(cy) + 86, W: ScreenWidth}, itoa(progress)+"%", color.RGBA{232, 210, 132, 255}, gameui.TextLarge, gameui.TextAlignCenter)
	drawUILabel(screen, gameui.Rect{X: 0, Y: float64(cy) + 114, W: ScreenWidth}, "Lütfen bekleyin", color.RGBA{150, 140, 110, 190}, gameui.TextSmall, gameui.TextAlignCenter)
}

func loadingBackgroundImage(scenarioPath string) *ebiten.Image {
	if scenarioPath == loadingBackgroundPath {
		return loadingBackground
	}

	loadingBackgroundPath = scenarioPath
	loadingBackground = nil
	if scenarioPath != "" {
		loadingBackground = tryLoadImage(filepath.Join(scenarioPath, "scenario_bg.png"))
	}
	return loadingBackground
}

func maxFloat32(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}
