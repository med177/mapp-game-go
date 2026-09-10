package main

import (
	"bytes"
	_ "embed"
	"image"
	_ "image/png"
	"log"

	"mapp-game-go/internal/game"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/joho/godotenv"
	"golang.org/x/image/draw"
)

//go:embed mapp_window_icon.png
var windowIconPNG []byte

func loadWindowIcon() {
	img, _, err := image.Decode(bytes.NewReader(windowIconPNG))
	if err != nil {
		log.Printf("Icon decode failed: %v", err)
		return
	}

	// Windows görev çubuğu ve başlık çubuğu farklı ikon boyutları ister.
	// Tek bir 512x512 görsel vermek bazı Windows sürümlerinde varsayılan
	// pencere ikonuna geri dönülmesine neden olabiliyor.
	iconSizes := []int{16, 32, 48, 64, 128, 256}
	icons := make([]image.Image, 0, len(iconSizes))
	for _, size := range iconSizes {
		resized := image.NewNRGBA(image.Rect(0, 0, size, size))
		draw.CatmullRom.Scale(resized, resized.Bounds(), img, img.Bounds(), draw.Src, nil)
		icons = append(icons, ebiten.NewImageFromImage(resized))
	}
	ebiten.SetWindowIcon(icons)
}

func main() {
	// .env dosyasını yüklemeyi dene (varsa)
	_ = godotenv.Load()

	loadWindowIcon()
	ebiten.SetWindowTitle("Mapp Game — Harita Strateji Oyunu")
	ebiten.SetWindowSize(1920, 1080)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.MaximizeWindow()
	// Pencerenin X düğmesine basıldığında uygulamanın kapanmasını oyunun
	// onay modalına bırak. Ebitengine aksi halde pencereyi hemen kapatır.
	ebiten.SetWindowClosingHandled(true)

	g := game.New()
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
