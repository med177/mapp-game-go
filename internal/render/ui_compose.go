package render

import (
	"image/color"
	"math"

	gameui "mapp-game-go/internal/ui"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

func drawUILabel(screen *ebiten.Image, rect gameui.Rect, text string, col color.Color, variant gameui.TextVariant, align gameui.TextAlign) {
	gameui.NewTextLabel(rect, text, col, variant, align).Draw(screen, renderText)
}

func drawUIWrappedLabel(screen *ebiten.Image, rect gameui.Rect, text string, col color.Color, variant gameui.TextVariant, lineStep float64, maxLines int) {
	label := gameui.NewWrappedLabel(rect, text, col, variant, lineStep)
	label.MaxLines = maxLines
	label.Ellipsis = maxLines > 0
	label.Draw(screen, renderText)
}

func drawUIWrappedLabelAligned(screen *ebiten.Image, rect gameui.Rect, text string, col color.Color, variant gameui.TextVariant, lineStep float64, maxLines int, align gameui.TextAlign) {
	label := gameui.NewWrappedLabel(rect, text, col, variant, lineStep)
	label.MaxLines = maxLines
	label.Ellipsis = maxLines > 0
	label.Align = align
	label.Draw(screen, renderText)
}

func drawUIOutlinedLabel(screen *ebiten.Image, rect gameui.Rect, text string, fill color.Color, outline color.Color, variant gameui.TextVariant, align gameui.TextAlign) {
	gameui.NewOutlinedLabel(rect, text, fill, outline, variant, align).Draw(screen, renderText)
}

func drawUIRichTextBlock(screen *ebiten.Image, rect gameui.Rect, lines []gameui.RichTextLine, lineStep float64) {
	gameui.NewRichTextBlock(rect, lines, lineStep).Draw(screen, renderText)
}

func drawUIKeyValueWidget(screen *ebiten.Image, row gameui.KeyValueRow) {
	row.Draw(screen, renderText)
}

func drawUITableRow(screen *ebiten.Image, row gameui.TableRow) {
	row.Draw(screen, renderText)
}

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
	drawUILabel(screen, gameui.Rect{X: rect.X, Y: rect.Y, W: rect.W}, title, ColorYellow, gameui.TextLarge, gameui.TextAlignCenter)
}

func drawUISectionLabel(screen *ebiten.Image, x, y float64, label string) {
	drawUILabel(screen, gameui.Rect{X: x, Y: y}, label, ColorGold, gameui.TextSmall, gameui.TextAlignStart)
}

func drawUICenteredSectionLabel(screen *ebiten.Image, centerX, y float64, label string) {
	drawUILabel(screen, gameui.Rect{X: centerX, Y: y}, label, ColorGold, gameui.TextSmall, gameui.TextAlignCenter)
}

func drawUIMutedText(screen *ebiten.Image, x, y float64, label string) {
	drawUILabel(screen, gameui.Rect{X: x, Y: y}, label, ColorGray, gameui.TextSmall, gameui.TextAlignStart)
}

func drawUIKeyValueRow(screen *ebiten.Image, x, y, w float64, label, value string, labelColor color.Color, valueColor color.Color) {
	if labelColor == nil {
		labelColor = ColorGray
	}
	if valueColor == nil {
		valueColor = ColorWhite
	}
	value = trimTextToWidth(value, FaceMed, w)
	row := gameui.NewKeyValueRow(gameui.Rect{X: x, Y: y, W: w}, label, value)
	row.LabelColor = labelColor
	row.ValueColor = valueColor
	row.LabelVariant = gameui.TextSmall
	row.ValueVariant = gameui.TextMedium
	drawUIKeyValueWidget(screen, row)
}

func drawUIInfoBlock(screen *ebiten.Image, x, y float64, lines []string, colors []color.Color) {
	for i, line := range lines {
		col := color.Color(ColorGray)
		if i < len(colors) && colors[i] != nil {
			col = colors[i]
		}
		drawUILabel(screen, gameui.Rect{X: x, Y: y + float64(i*24)}, line, col, gameui.TextSmall, gameui.TextAlignStart)
	}
}

func drawUIScreenChrome(screen *ebiten.Image, bg color.RGBA, title string, subtitle string) {
	screen.Fill(bg)
	drawUIScreenChromeOverlay(screen, title, subtitle)
}

func drawUIScreenChromeOverlay(screen *ebiten.Image, title string, subtitle string) {
	vector.FillRect(screen, 0, 0, float32(ScreenWidth), 3, color.RGBA{180, 150, 60, 200}, false)
	vector.FillRect(screen, 0, float32(ScreenHeight)-3, float32(ScreenWidth), 3, color.RGBA{180, 150, 60, 200}, false)
	drawUILabel(screen, gameui.Rect{X: 0, Y: 40, W: ScreenWidth}, title, ColorYellow, gameui.TextLarge, gameui.TextAlignCenter)
	if subtitle != "" {
		drawUILabel(screen, gameui.Rect{X: 0, Y: 70, W: ScreenWidth}, subtitle, ColorGray, gameui.TextSmall, gameui.TextAlignCenter)
	}
}

func drawUIImageCover(screen, image *ebiten.Image) {
	if image == nil {
		return
	}

	bounds := image.Bounds()
	sourceW := float64(bounds.Dx())
	sourceH := float64(bounds.Dy())
	if sourceW <= 0 || sourceH <= 0 {
		return
	}

	scale := math.Max(ScreenWidth/sourceW, ScreenHeight/sourceH)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate((ScreenWidth-sourceW*scale)/2, (ScreenHeight-sourceH*scale)/2)
	screen.DrawImage(image, op)
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
