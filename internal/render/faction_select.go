package render

import (
	"image/color"
	"sort"

	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// DrawFactionSelect fraksiyon seçim ekranını çizer.
func DrawFactionSelect(screen *ebiten.Image, gs *state.GameState, cursor int) {
	screen.Fill(color.RGBA{10, 8, 5, 255})

	factions := selectableFactions(gs)
	cols := 3
	rows := (len(factions) + cols - 1) / cols
	cardW := float32(350)
	cardH := float32(110)
	padX := float32(30)
	padY := float32(12)

	gridW := cardW*float32(cols) + padX*float32(cols-1)
	gridH := cardH*float32(rows) + padY*float32(rows-1)
	headerH := float32(70) // başlık + ipucu için alan

	startX := float32(ScreenWidth)/2 - gridW/2
	startY := float32(ScreenHeight)/2 - (gridH+headerH)/2

	DrawTextCentered(screen, "MAPP — Fraksiyon Seç", ScreenWidth/2, float64(startY)+4, FaceLarge, ColorYellow)
	DrawTextCentered(screen, "[↑↓] Seç   [Enter] Onayla   [F11] Tam Ekran", ScreenWidth/2, float64(startY)+30, FaceSmall, ColorGray)

	startY += headerH

	for i, fid := range factions {
		f := gs.Factions[fid]
		col := i % cols
		row := i / cols
		x := startX + float32(col)*(cardW+padX)
		y := startY + float32(row)*(cardH+padY)

		fc := color.RGBA{f.Color[0], f.Color[1], f.Color[2], 255}
		bgCol := color.RGBA{22, 18, 12, 220}
		borderCol := color.RGBA{80, 65, 40, 200}
		if i == cursor {
			bgCol = color.RGBA{45, 36, 20, 240}
			borderCol = fc
		}

		vector.FillRect(screen, x, y, cardW, cardH, bgCol, false)
		vector.StrokeLine(screen, x, y, x+cardW, y, 2, borderCol, false)
		vector.StrokeLine(screen, x, y+cardH, x+cardW, y+cardH, 2, borderCol, false)
		vector.StrokeLine(screen, x, y, x, y+cardH, 2, borderCol, false)
		vector.StrokeLine(screen, x+cardW, y, x+cardW, y+cardH, 2, borderCol, false)

		// Renk şeridi
		vector.FillRect(screen, x, y, 8, cardH, fc, false)

		// İsim
		nameCol := ColorWhite
		if i == cursor {
			nameCol = ColorYellow
		}
		DrawText(screen, f.NameTR, float64(x+16), float64(y+12), FaceLarge, nameCol)

		// Din
		DrawText(screen, religionTR(f.Religion), float64(x+16), float64(y+36), FaceSmall, ColorGray)

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

func religionTR(r faction.Religion) string {
	switch r {
	case faction.ReligionCatholic:
		return "Katolik"
	case faction.ReligionOrthodox:
		return "Ortodoks"
	case faction.ReligionSunni:
		return "Sünni İslam"
	case faction.ReligionShia:
		return "Şii İslam"
	}
	return string(r)
}
