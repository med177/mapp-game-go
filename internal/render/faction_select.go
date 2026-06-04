package render

import (
	"image/color"
	"sort"

	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/religion"
	"mapp-game-go/internal/state"
	gameui "mapp-game-go/internal/ui"

	"github.com/hajimehoshi/ebiten/v2"
)

func buildFactionCardButtons(gs *state.GameState) []gameui.Button {
	factions := selectableFactions(gs)
	cols := 3
	rows := (len(factions) + cols - 1) / cols
	cardW := 350.0
	cardH := 110.0
	padX := 30.0
	padY := 12.0
	headerH := 70.0
	grid := centeredGridRect(cols, rows, cardW, cardH, padX, padY, headerH)
	buttons := make([]gameui.Button, 0, len(factions))
	for i, fid := range factions {
		col := i % cols
		row := i / cols
		r := gridCellRect(grid, cardW, cardH, padX, padY, col, row)
		label := ""
		if f := gs.Factions[fid]; f != nil {
			label = f.NameTR
		}
		buttons = append(buttons, gameui.NewButton(r.X, r.Y, r.W, r.H, label))
	}
	return buttons
}

// DrawFactionSelect fraksiyon seçim ekranını çizer.
func DrawFactionSelect(screen *ebiten.Image, gs *state.GameState, cursor int) {
	drawUIScreenChrome(screen, color.RGBA{10, 8, 5, 255}, "MAPP — Fraksiyon Seç", "Fraksiyon kartını seçmek için tıkla")

	factions := selectableFactions(gs)
	cols := 3
	rows := (len(factions) + cols - 1) / cols
	cardW := float32(350)
	cardH := float32(110)
	padX := float32(30)
	padY := float32(12)

	headerH := 70.0
	grid := centeredGridRect(cols, rows, float64(cardW), float64(cardH), float64(padX), float64(padY), headerH)

	drawBackButton(screen)

	for i, fid := range factions {
		f := gs.Factions[fid]
		col := i % cols
		row := i / cols
		cell := gridCellRect(grid, float64(cardW), float64(cardH), float64(padX), float64(padY), col, row)
		x := float32(cell.X)
		y := float32(cell.Y)

		fc := color.RGBA{f.Color[0], f.Color[1], f.Color[2], 255}
		bgCol := color.RGBA{22, 18, 12, 220}
		borderCol := color.RGBA{80, 65, 40, 200}
		if i == cursor {
			bgCol = color.RGBA{45, 36, 20, 240}
			borderCol = fc
		}

		drawUICardRect(screen, cell, bgCol, borderCol, 2)

		// Renk şeridi
		drawUICardAccent(screen, cell, 8, fc)

		// İsim
		nameCol := ColorWhite
		if i == cursor {
			nameCol = ColorYellow
		}
		DrawText(screen, f.NameTR, float64(x+16), float64(y+12), FaceLarge, nameCol)

		// Din
		DrawText(screen, religion.DisplayNameTR(f.Religion), float64(x+16), float64(y+36), FaceSmall, ColorGray)

		// Bölge sayısı ve başlangıç altını
		regionCount := len(gs.RegionsOwnedBy(fid))
		DrawText(screen, itoa(regionCount)+" bölge", float64(x+16), float64(y+54), FaceSmall, ColorGold)

		if i == cursor {
			DrawText(screen, "← SEÇİLİ", float64(x+cardW-90), float64(y+12), FaceSmall, fc)
		}
	}
}

func selectableFactions(gs *state.GameState) []faction.FactionID {
	var fids []faction.FactionID
	for fid, f := range gs.Factions {
		if f.IsPlayable {
			fids = append(fids, fid)
		}
	}
	sort.Slice(fids, func(i, j int) bool { return fids[i] < fids[j] })
	return fids
}
