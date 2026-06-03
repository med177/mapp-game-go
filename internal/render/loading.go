package render

import (
	"image/color"
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

func DrawLoadingScreen(screen *ebiten.Image, message string, progress int, tick int) {
	if message == "" {
		message = "Yükleniyor..."
	}
	screen.Fill(color.RGBA{7, 8, 12, 255})
	vector.FillRect(screen, 0, 0, float32(ScreenWidth), 3, color.RGBA{180, 150, 60, 200}, false)
	vector.FillRect(screen, 0, float32(ScreenHeight)-3, float32(ScreenWidth), 3, color.RGBA{180, 150, 60, 200}, false)

	cx := float32(ScreenWidth / 2)
	cy := float32(ScreenHeight/2 - 18)
	radius := float32(28)
	phase := float64(time.Now().UnixNano()%int64(60*time.Second)) / float64(time.Second) * 4.5
	for i := 0; i < 12; i++ {
		angle := float64(i)/12*math.Pi*2 + phase + float64(tick)*0.01
		alpha := uint8(60 + i*14)
		x := cx + float32(math.Cos(angle))*radius
		y := cy + float32(math.Sin(angle))*radius
		vector.FillCircle(screen, x, y, 3.6, color.RGBA{220, 180, 70, alpha}, true)
	}

	DrawTextCentered(screen, message, ScreenWidth/2, float64(cy)+46, FaceLarge, ColorYellow)
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

	DrawTextCentered(screen, itoa(progress)+"%", ScreenWidth/2, float64(cy)+86, FaceLarge, color.RGBA{232, 210, 132, 255})
	DrawTextCentered(screen, "Lütfen bekleyin", ScreenWidth/2, float64(cy)+114, FaceSmall, color.RGBA{150, 140, 110, 190})
}

func maxFloat32(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}
