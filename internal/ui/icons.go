package ui

import (
	"image"
	"image/color"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
)

type IconID string

const (
	IconNone   IconID = ""
	IconBack   IconID = "back"
	IconClose  IconID = "close"
	IconMenu   IconID = "menu"
	IconBook   IconID = "book"
	IconMinus  IconID = "minus"
	IconPlus   IconID = "plus"
	IconPlay   IconID = "play"
	IconPause  IconID = "pause"
	IconNext   IconID = "next"
	IconSend   IconID = "send"
	IconTrash  IconID = "trash"
	IconCheck  IconID = "check"
	IconSword  IconID = "sword"
	IconSave   IconID = "save"
	IconLoad   IconID = "load"
	IconBuy    IconID = "buy"
	IconSell   IconID = "sell"
	IconExit   IconID = "exit"
)

const iconAssetDir = "assets/ui/icons"

var uiIconCache = map[IconID]*ebiten.Image{}

func init() {
	for _, id := range []IconID{
		IconBack,
		IconClose,
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
		IconSell,
		IconExit,
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
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(size/float64(sw), size/float64(sh))
	op.GeoM.Translate(x, y)
	op.ColorScale.ScaleWithColor(tint)
	screen.DrawImage(src, op)
	return true
}

func loadIconAsset(id IconID) *ebiten.Image {
	if id == IconNone {
		return nil
	}
	base := resolveIconAssetDir()
	if base == "" {
		return nil
	}
	path := filepath.Join(base, string(id)+".png")
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		return nil
	}
	return ebiten.NewImageFromImage(img)
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
