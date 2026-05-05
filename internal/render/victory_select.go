package render

import (
	"image/color"

	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type victoryOption struct {
	vtype   state.VictoryType
	title   string
	desc    string
	detail  string
}

var victoryOptions = []victoryOption{
	{
		state.VictoryDomination,
		"Toprak Hakimiyeti",
		"20+ bölge ve 5 kritik şehri ele geçir",
		"Roma, Konstantinopolis, Kahire, İstanbul, Paris",
	},
	{
		state.VictoryEconomic,
		"Ekonomik Güç",
		"Tur başı 500+ altın gelire ulaş ve 5 tur koru",
		"Ticaret ağları kur, bölgelerini geliştir",
	},
	{
		state.VictoryMilitary,
		"Askeri Üstünlük",
		"3 büyük fraksiyonu yenilgiye uğrat",
		"Düşman fraksiyonların son bölgelerini al",
	},
	{
		state.VictoryReligious,
		"Dinî Zafer",
		"Kudüs, Roma ve Mekke'yi aynı anda tut",
		"Kutsal şehirleri sürdürülebilir biçimde koru",
	},
}

// DrawVictorySelect zafer koşulu seçim ekranını çizer.
func DrawVictorySelect(screen *ebiten.Image, cursor int) {
	screen.Fill(color.RGBA{10, 10, 20, 255})

	cardW, cardH := 520.0, 100.0
	gap := 12.0
	n := float64(len(victoryOptions))
	totalH := n*cardH + (n-1)*gap
	headerH := 80.0

	startY := (ScreenHeight-(totalH+headerH))/2 + headerH
	cx := ScreenWidth/2 - cardW/2

	DrawTextCentered(screen, "ZAFER KOŞULUNU SEÇ", ScreenWidth/2, startY-headerH+10, FaceLarge, ColorYellow)
	DrawTextCentered(screen, "Nasıl kazanmak istiyorsun?", ScreenWidth/2, startY-headerH+38, FaceSmall, ColorGray)

	for i, opt := range victoryOptions {
		y := startY + float64(i)*(cardH+gap)

		bg := color.RGBA{25, 25, 45, 220}
		border := color.RGBA{80, 80, 120, 200}
		if i == cursor {
			bg = color.RGBA{50, 45, 90, 240}
			border = color.RGBA{200, 160, 60, 255}
		}

		vector.FillRect(screen, float32(cx), float32(y), float32(cardW), float32(cardH), bg, false)
		vector.StrokeRect(screen, float32(cx), float32(y), float32(cardW), float32(cardH), 2, border, false)

		titleCol := ColorWhite
		if i == cursor {
			titleCol = ColorYellow
		}
		DrawText(screen, opt.title, cx+18, y+14, FaceLarge, titleCol)
		DrawText(screen, opt.desc, cx+18, y+38, FaceMed, ColorGray)
		DrawText(screen, opt.detail, cx+18, y+60, FaceSmall, color.RGBA{140, 120, 80, 220})

		if i == cursor {
			DrawText(screen, "← SEÇİLİ", cx+cardW-110, y+14, FaceSmall, ColorGold)
		}
	}

	DrawTextCentered(screen, "[↑↓] Seç   [Enter] Onayla   [Esc] Geri", ScreenWidth/2, startY+totalH+20, FaceSmall, ColorGray)
}

// handleVictorySelectInput zafer seçim ekranı girişini işler.
func (r *Renderer) handleVictorySelectInput() InputAction {
	n := len(victoryOptions)

	// Hover ile kart vurgusunu güncelle
	mx, my := ebiten.CursorPosition()
	if i := r.victoryCardHoverIndex(float64(mx), float64(my)); i >= 0 {
		r.factionCursor = i
	}

	if r.keyJustPressed(ebiten.KeyArrowDown) {
		r.factionCursor = (r.factionCursor + 1) % n
	}
	if r.keyJustPressed(ebiten.KeyArrowUp) {
		r.factionCursor = (r.factionCursor - 1 + n) % n
	}
	if r.keyJustPressed(ebiten.KeyEnter) {
		opt := victoryOptions[r.factionCursor]
		return InputAction{Kind: ActionSelectVictory, BuildingID: string(opt.vtype)}
	}
	if r.mouseJustPressed(ebiten.MouseButtonLeft) {
		if i := r.victoryCardHoverIndex(float64(mx), float64(my)); i >= 0 {
			opt := victoryOptions[i]
			return InputAction{Kind: ActionSelectVictory, BuildingID: string(opt.vtype)}
		}
	}
	if r.keyJustPressed(ebiten.KeyEscape) {
		r.factionCursor = 0
		return InputAction{Kind: ActionBack}
	}
	return InputAction{}
}

// victoryRegions zafer tipine göre hedef bölgeleri döner.
func victoryRegions(vtype state.VictoryType) []world.RegionID {
	switch vtype {
	case state.VictoryReligious:
		return []world.RegionID{"jerusalem", "rome", "mecca"}
	case state.VictoryDomination:
		return []world.RegionID{"constantinople", "rome", "paris", "cairo", "jerusalem"}
	}
	return nil
}
