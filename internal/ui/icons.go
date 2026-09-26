package ui

import (
	"image"
	"image/color"
	"image/draw"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
)

type IconID string

const (
	IconNone      IconID = ""
	IconBack      IconID = "back"
	IconClose     IconID = "close"
	IconX         IconID = "x"
	IconMenu      IconID = "menu"
	IconBook      IconID = "book"
	IconMinus     IconID = "minus"
	IconPlus      IconID = "plus"
	IconPlay      IconID = "play"
	IconPause     IconID = "pause"
	IconNext      IconID = "next"
	IconSend      IconID = "send"
	IconTrash     IconID = "trash"
	IconCheck     IconID = "check"
	IconSword     IconID = "sword"
	IconSave      IconID = "save"
	IconLoad      IconID = "load"
	IconBuy       IconID = "buy"
	IconGold      IconID = "gold"
	IconSell      IconID = "sell"
	IconDiplomacy IconID = "diplomacy"
	IconNaval     IconID = "naval"
	IconReligion  IconID = "religion"
	IconExit      IconID = "exit"
	IconLogin     IconID = "login"
	IconDock      IconID = "dock"
	IconExport    IconID = "export"
	IconChange    IconID = "change"
	IconSplit     IconID = "split"
	IconForward   IconID = "forward"
	IconRevolt    IconID = "revolt"
)

const iconAssetDir = "assets/ui/icons"

var uiIconCache = map[IconID]*ebiten.Image{}

func init() {
	for _, id := range []IconID{
		IconBack,
		IconClose,
		IconX,
		IconMenu,
		IconBook,
		IconMinus,
		IconPlus,
		IconPlay,
		IconPause,
		IconNext,
		IconSend,
		IconTrash,
		IconCheck,
		IconSword,
		IconSave,
		IconLoad,
		IconBuy,
		IconGold,
		IconSell,
		IconDiplomacy,
		IconNaval,
		IconReligion,
		IconExit,
		IconLogin,
		IconDock,
		IconExport,
		IconChange,
		IconSplit,
		IconForward,
		IconRevolt,
	} {
		uiIconCache[id] = loadIconAsset(id)
	}
}

func iconImage(id IconID) *ebiten.Image {
	return uiIconCache[id]
}

func DrawIcon(screen *ebiten.Image, id IconID, x, y, size float64, tint color.Color) bool {
	src := iconImage(id)
	if src == nil {
		return false
	}
	sw, sh := src.Bounds().Dx(), src.Bounds().Dy()
	if sw == 0 || sh == 0 {
		return false
	}
	drawW, drawH := iconFitDimensions(sw, sh, size)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(drawW/float64(sw), drawH/float64(sh))
	op.GeoM.Translate(x+(size-drawW)/2, y+(size-drawH)/2)
	op.ColorScale.ScaleWithColor(tint)
	screen.DrawImage(src, op)
	return true
}

func iconFitDimensions(sw, sh int, size float64) (float64, float64) {
	if sw <= 0 || sh <= 0 || size <= 0 {
		return 0, 0
	}
	if sw >= sh {
		return size, size * float64(sh) / float64(sw)
	}
	return size * float64(sw) / float64(sh), size
}

func loadIconAsset(id IconID) *ebiten.Image {
	if id == IconNone {
		return nil
	}
	base := resolveIconAssetDir()
	if base == "" {
		return nil
	}
	assetID := id
	// IconX is used by compact destructive controls. Reuse the canonical
	// close asset so every X control has a visible icon without duplicating
	// the bitmap in the assets directory.
	if id == IconX {
		assetID = IconClose
	}
	path := filepath.Join(base, string(assetID)+".png")
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		return nil
	}
	return ebiten.NewImageFromImage(cropTransparentBorder(img))
}

func cropTransparentBorder(src image.Image) image.Image {
	bounds := src.Bounds()
	minX, minY := bounds.Max.X, bounds.Max.Y
	maxX, maxY := bounds.Min.X, bounds.Min.Y
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			_, _, _, alpha := src.At(x, y).RGBA()
			if alpha == 0 {
				continue
			}
			if x < minX {
				minX = x
			}
			if y < minY {
				minY = y
			}
			if x+1 > maxX {
				maxX = x + 1
			}
			if y+1 > maxY {
				maxY = y + 1
			}
		}
	}
	if minX >= maxX || minY >= maxY {
		return src
	}
	cropped := image.NewRGBA(image.Rect(0, 0, maxX-minX, maxY-minY))
	draw.Draw(cropped, cropped.Bounds(), src, image.Point{X: minX, Y: minY}, draw.Src)
	return cropped
}

func resolveIconAssetDir() string {
	candidates := []string{iconAssetDir}
	prefix := ""
	for i := 0; i < 5; i++ {
		prefix = filepath.Join(prefix, "..")
		candidates = append(candidates, filepath.Join(prefix, iconAssetDir))
	}
	for _, candidate := range candidates {
		if strings.TrimSpace(candidate) == "" {
			continue
		}
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate
		}
	}
	return ""
}
