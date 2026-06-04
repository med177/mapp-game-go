package render

import (
	"image/color"

	gameui "mapp-game-go/internal/ui"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

func drawUIOverlay(screen *ebiten.Image, c color.RGBA) {
	vector.FillRect(screen, 0, 0, float32(ScreenWidth), float32(ScreenHeight), c, false)
}

func drawUIPanelRect(screen *ebiten.Image, rect gameui.Rect, fill color.RGBA, border color.RGBA, borderWidth float32) {
	gameui.DrawPanel(screen, gameui.Panel{Rect: rect, Visible: true}, gameui.PanelStyle{
		BG:          fill,
		Border:      border,
		BorderWidth: borderWidth,
	})
}

func drawUIPanelTopBar(screen *ebiten.Image, rect gameui.Rect, h float32, c color.RGBA) {
	vector.FillRect(screen, float32(rect.X), float32(rect.Y), float32(rect.W), h, c, false)
}

func drawUIPanelFrame(screen *ebiten.Image, rect gameui.Rect, fill color.RGBA, border color.RGBA, borderWidth float32, topBarH float32) {
	drawUIPanelRect(screen, rect, fill, border, borderWidth)
	if topBarH > 0 {
		drawUIPanelTopBar(screen, rect, topBarH, border)
	}
}

func drawUIPanelTitle(screen *ebiten.Image, rect gameui.Rect, title string) {
	DrawTextCentered(screen, title, rect.X+rect.W/2, rect.Y, FaceLarge, ColorYellow)
}

func drawUISectionLabel(screen *ebiten.Image, x, y float64, label string) {
	DrawText(screen, label, x, y, FaceSmall, ColorGold)
}

func drawUIMutedText(screen *ebiten.Image, x, y float64, label string) {
	DrawText(screen, label, x, y, FaceSmall, ColorGray)
}

func drawUIInfoBlock(screen *ebiten.Image, x, y float64, lines []string, colors []color.Color) {
	for i, line := range lines {
		col := color.Color(ColorGray)
		if i < len(colors) && colors[i] != nil {
			col = colors[i]
		}
		DrawText(screen, line, x, y+float64(i*24), FaceSmall, col)
	}
}

func drawUIScreenChrome(screen *ebiten.Image, bg color.RGBA, title string, subtitle string) {
	screen.Fill(bg)
	vector.FillRect(screen, 0, 0, float32(ScreenWidth), 3, color.RGBA{180, 150, 60, 200}, false)
	vector.FillRect(screen, 0, float32(ScreenHeight)-3, float32(ScreenWidth), 3, color.RGBA{180, 150, 60, 200}, false)
	DrawTextCentered(screen, title, ScreenWidth/2, 40, FaceLarge, ColorYellow)
	if subtitle != "" {
		DrawTextCentered(screen, subtitle, ScreenWidth/2, 70, FaceSmall, ColorGray)
	}
}

func drawUICardRect(screen *ebiten.Image, rect gameui.Rect, fill color.RGBA, border color.RGBA, borderWidth float32) {
	vector.FillRect(screen, float32(rect.X), float32(rect.Y), float32(rect.W), float32(rect.H), fill, false)
	vector.StrokeRect(screen, float32(rect.X), float32(rect.Y), float32(rect.W), float32(rect.H), borderWidth, border, false)
}

func drawUICardAccent(screen *ebiten.Image, rect gameui.Rect, width float32, fill color.RGBA) {
	vector.FillRect(screen, float32(rect.X), float32(rect.Y), width, float32(rect.H), fill, false)
}

func drawUISeparator(screen *ebiten.Image, x1, y, x2 float32, width float32, c color.RGBA) {
	vector.StrokeLine(screen, x1, y, x2, y, width, c, false)
}

func drawUIProgressBar(screen *ebiten.Image, x, y, w, h float32, fill float64, bg color.RGBA, border color.RGBA, fg color.Color, borderWidth float32) {
	fill = clampF(fill)
	vector.FillRect(screen, x, y, w, h, bg, false)
	if fill > 0 {
		vector.FillRect(screen, x, y, float32(float64(w)*fill), h, fg, false)
	}
	if borderWidth > 0 {
		vector.StrokeRect(screen, x, y, w, h, borderWidth, border, false)
	}
}
