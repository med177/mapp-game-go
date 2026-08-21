package render

import (
	"bytes"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/gofont/goregular"
)

var (
	FaceSmall     *text.GoTextFace // 12px — yardımcı metinler
	FaceTiny      *text.GoTextFace // 10px — sıkışık durum etiketleri
	FaceMicro     *text.GoTextFace // 8px — dar marker rozetleri
	FaceMed       *text.GoTextFace // 14px — genel UI
	FaceLarge     *text.GoTextFace // 18px — başlıklar
	FaceLargeBold *text.GoTextFace // 18px — kritik vurgular

	fontSource     *text.GoTextFaceSource
	boldFontSource *text.GoTextFaceSource
)

func init() {
	var err error
	fontSource, err = text.NewGoTextFaceSource(bytes.NewReader(goregular.TTF))
	if err != nil {
		log.Fatalf("Font yüklenemedi: %v", err)
	}
	boldFontSource, err = text.NewGoTextFaceSource(bytes.NewReader(gobold.TTF))
	if err != nil {
		log.Fatalf("Kalın font yüklenemedi: %v", err)
	}
	FaceSmall = &text.GoTextFace{Source: fontSource, Size: 12}
	FaceTiny = &text.GoTextFace{Source: fontSource, Size: 10}
	FaceMicro = &text.GoTextFace{Source: fontSource, Size: 8}
	FaceMed = &text.GoTextFace{Source: fontSource, Size: 14}
	FaceLarge = &text.GoTextFace{Source: fontSource, Size: 18}
	FaceLargeBold = &text.GoTextFace{Source: boldFontSource, Size: 18}
}

// DrawText ekrana renkli metin yazar.
func DrawText(screen *ebiten.Image, str string, x, y float64, face *text.GoTextFace, col color.Color) {
	op := &text.DrawOptions{}
	op.GeoM.Translate(x, y)
	op.ColorScale.ScaleWithColor(col)
	text.Draw(screen, str, face, op)
}

// MeasureText metnin piksel genişliğini ölçer.
func MeasureText(str string, face *text.GoTextFace) float64 {
	w, _ := text.Measure(str, face, 0)
	return w
}

func markerBadgeTextFace(label string, badgeWidth float64) *text.GoTextFace {
	if MeasureText(label, FaceTiny) > badgeWidth-5 {
		return FaceMicro
	}
	return FaceTiny
}

// DrawTextCentered metni x konumuna göre ortalar.
func DrawTextCentered(screen *ebiten.Image, str string, cx, y float64, face *text.GoTextFace, col color.Color) {
	w := MeasureText(str, face)
	DrawText(screen, str, cx-w/2, y, face, col)
}

var (
	ColorWhite  = color.RGBA{255, 255, 255, 255}
	ColorYellow = color.RGBA{255, 220, 60, 255}
	ColorGold   = color.RGBA{220, 180, 40, 255}
	ColorRed    = color.RGBA{230, 60, 60, 255}
	ColorGreen  = color.RGBA{80, 210, 80, 255}
	ColorGray   = color.RGBA{180, 180, 180, 255}
	ColorDark   = color.RGBA{30, 25, 20, 255}
)
