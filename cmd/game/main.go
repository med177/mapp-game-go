package main

import (
	"log"

	"mapp-game-go/internal/game"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/joho/godotenv"
)

func main() {
	// .env dosyasını yüklemeyi dene (varsa)
	_ = godotenv.Load()

	ebiten.SetWindowTitle("Mapp — Orta Çağ Strateji")
	ebiten.SetWindowSize(1920, 1080)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.MaximizeWindow()

	g := game.New()
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
