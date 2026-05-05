package render

import (
	"image"
	_ "image/png" // PNG decoder
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
)

// loadImage bir resim dosyasını diskten yükler ve bir ebiten.Image olarak döndürür.
// Hata durumunda programı sonlandırır, çünkü temel asset'ler olmadan oyun başlayamaz.
func loadImage(path string) *ebiten.Image {
	f, err := os.Open(path)
	if err != nil {
		log.Fatalf("Hata: Resim dosyası açılamadı %s: %v", path, err)
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		log.Fatalf("Hata: Resim dosyası decode edilemedi %s: %v", path, err)
	}
	return ebiten.NewImageFromImage(img)
}
