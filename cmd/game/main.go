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
)

//go:embed mapp_window_icon.png
var windowIconPNG []byte

func loadWindowIcon() {
	img, _, err := image.Decode(bytes.NewReader(windowIconPNG))
	if err != nil {
		log.Printf("Icon decode failed: %v", err)
		return
	}
	ebiten.SetWindowIcon([]image.Image{ebiten.NewImageFromImage(img)})
}

func main() {
	// .env dosyasını yüklemeyi dene (varsa)
	_ = godotenv.Load()

	ebiten.SetWindowTitle("Mapp — Orta Çağ Strateji")
	ebiten.SetWindowSize(1920, 1080)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.MaximizeWindow()
	loadWindowIcon()

	g := game.New()
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
