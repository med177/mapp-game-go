package render

import (
	"image/color"
	"sort"

	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	diplomStartY = 80.0
	diplomRowH   = 52.0
)

type diplomAction struct {
	label  string
	color  color.RGBA
	action ActionKind
}

var diplomActions = []diplomAction{
	{"Savaş", color.RGBA{180, 50, 50, 220}, ActionDeclareWar},
	{"Barış", color.RGBA{50, 120, 180, 220}, ActionProposePeace},
	{"İttifak", color.RGBA{50, 160, 80, 220}, ActionProposeAlliance},
	{"Ticaret", color.RGBA{160, 130, 50, 220}, ActionProposeTrade},
}

// diplomRowRect seçili satırdaki i. aksiyon butonunun dikdörtgenini döner.
func diplomActionRect(rowY float64, i int) (x, y, w, h float32) {
	btnW := float32(80)
	btnH := float32(22)
	gap := float32(6)
	rightEdge := float32(ScreenWidth) - 60
	x = rightEdge - float32(len(diplomActions))*(btnW+gap) + float32(i)*(btnW+gap)
	y = float32(rowY) + float32((diplomRowH-float64(btnH))/2)
	return x, y, btnW, btnH
}

// DrawDiplomacyPanel diplomasi panelini çizer.
func DrawDiplomacyPanel(screen *ebiten.Image, gs *state.GameState, focusIdx int) {
	overlay := ebiten.NewImage(int(ScreenWidth), int(ScreenHeight))
	overlay.Fill(color.RGBA{8, 6, 4, 220})
	screen.DrawImage(overlay, nil)

	DrawTextCentered(screen, "── Diplomasi ──", ScreenWidth/2, 24, FaceLarge, ColorYellow)
	drawDiplomacyCloseButton(screen)
	DrawText(screen, "Sol tık: Seç   Butonlarla aksiyon",
		30, 50, FaceSmall, ColorGray)

	factions := sortedFactions(gs)

	for i, fid := range factions {
		f := gs.Factions[fid]
		rel := gs.Relations[faction.RelationKey(gs.PlayerFactionID, fid)]

		y := diplomStartY + float64(i)*diplomRowH
		rowCol := color.RGBA{25, 20, 14, 200}
		if i == focusIdx {
			rowCol = color.RGBA{55, 45, 25, 230}
		}
		vector.FillRect(screen, 28, float32(y), float32(ScreenWidth)-56, float32(diplomRowH-4), rowCol, false)

		fc := color.RGBA{f.Color[0], f.Color[1], f.Color[2], 255}
		vector.FillRect(screen, 28, float32(y), 6, float32(diplomRowH-4), fc, false)

		DrawText(screen, f.NameTR, 42, y+6, FaceMed, ColorWhite)
		regionCount := len(gs.RegionsOwnedBy(fid))
		DrawText(screen, itoa(regionCount)+" bölge", 42, y+24, FaceSmall, ColorGray)

		if rel != nil {
			stanceCol, stanceTR := stanceDisplay(rel.Stance)
			DrawText(screen, stanceTR, 300, y+6, FaceMed, stanceCol)
			scoreCol := scoreColor(rel.Score)
			DrawText(screen, "İlişki: "+itoa(rel.Score), 300, y+24, FaceSmall, scoreCol)
		} else {
			DrawText(screen, "Tarafsız", 300, y+6, FaceMed, ColorGray)
		}

		armyCount := 0
		for _, a := range gs.Armies {
			if a.OwnerID == string(fid) {
				armyCount++
			}
		}
		DrawText(screen, itoa(armyCount)+" ordu", 500, y+6, FaceSmall, ColorGray)

		// Seçili satırda aksiyon butonları
		if i == focusIdx {
			for j, da := range diplomActions {
				bx, by, bw, bh := diplomActionRect(y, j)
				vector.FillRect(screen, bx, by, bw, bh, da.color, false)
				vector.StrokeRect(screen, bx, by, bw, bh, 1, panelBorder, false)
				tw := MeasureText(da.label, FaceSmall)
				DrawText(screen, da.label, float64(bx)+float64(bw)/2-tw/2, float64(by)+4, FaceSmall, ColorWhite)
			}
		}
	}
}

func diplomacyCloseRect() (x, y, w, h float32) {
	return float32(ScreenWidth) - 58, 20, 30, 26
}

func drawDiplomacyCloseButton(screen *ebiten.Image) {
	x, y, w, h := diplomacyCloseRect()
	vector.FillRect(screen, x, y, w, h, color.RGBA{45, 34, 25, 230}, false)
	vector.StrokeRect(screen, x, y, w, h, 1, panelBorder, false)
	tw := MeasureText("X", FaceSmall)
	DrawText(screen, "X", float64(x)+float64(w)/2-tw/2, float64(y)+6, FaceSmall, ColorGold)
}

func diplomacyCloseHit(mx, my float64) bool {
	x, y, w, h := diplomacyCloseRect()
	return mx >= float64(x) && mx <= float64(x+w) && my >= float64(y) && my <= float64(y+h)
}

// handleDiplomacyInput diplomasi paneli klavye ve fare girişini işler.
func (r *Renderer) handleDiplomacyInput() InputAction {
	factions := sortedFactions(r.gs)
	n := len(factions)

	mx, my := ebiten.CursorPosition()
	fx, fy := float64(mx), float64(my)

	// Hover → satır güncelle
	for i := range factions {
		y := diplomStartY + float64(i)*diplomRowH
		if fy >= y && fy <= y+diplomRowH-4 && fx >= 28 && fx <= ScreenWidth-56 {
			r.diplomacyFocus = i
			break
		}
	}

	// Sol tık → aksiyon butonu veya satır seçimi
	if r.mouseJustPressed(ebiten.MouseButtonLeft) {
		if diplomacyCloseHit(fx, fy) {
			r.showDiplomacy = false
			return InputAction{}
		}
		if r.diplomacyFocus < len(factions) {
			target := factions[r.diplomacyFocus]
			y := diplomStartY + float64(r.diplomacyFocus)*diplomRowH
			for j, da := range diplomActions {
				bx, by, bw, bh := diplomActionRect(y, j)
				if fx >= float64(bx) && fx <= float64(bx+bw) && fy >= float64(by) && fy <= float64(by+bh) {
					r.showDiplomacy = false
					return InputAction{Kind: da.action, TargetFaction: target}
				}
			}
		}
	}

	if r.keyJustPressed(ebiten.KeyArrowDown) && r.diplomacyFocus < n-1 {
		r.diplomacyFocus++
	}
	if r.keyJustPressed(ebiten.KeyArrowUp) && r.diplomacyFocus > 0 {
		r.diplomacyFocus--
	}
	if r.keyJustPressed(ebiten.KeyTab) || r.keyJustPressed(ebiten.KeyEscape) {
		r.showDiplomacy = false
		return InputAction{}
	}
	if r.diplomacyFocus < len(factions) {
		target := factions[r.diplomacyFocus]
		if r.keyJustPressed(ebiten.KeyW) {
			r.showDiplomacy = false
			return InputAction{Kind: ActionDeclareWar, TargetFaction: target}
		}
		if r.keyJustPressed(ebiten.KeyP) {
			r.showDiplomacy = false
			return InputAction{Kind: ActionProposePeace, TargetFaction: target}
		}
		if r.keyJustPressed(ebiten.KeyA) {
			r.showDiplomacy = false
			return InputAction{Kind: ActionProposeAlliance, TargetFaction: target}
		}
		if r.keyJustPressed(ebiten.KeyC) {
			r.showDiplomacy = false
			return InputAction{Kind: ActionProposeTrade, TargetFaction: target}
		}
	}
	return InputAction{}
}

func sortedFactions(gs *state.GameState) []faction.FactionID {
	var fids []faction.FactionID
	for fid := range gs.Factions {
		if fid == gs.PlayerFactionID {
			continue
		}
		fids = append(fids, fid)
	}
	sort.Slice(fids, func(i, j int) bool { return fids[i] < fids[j] })
	return fids
}

func stanceDisplay(s faction.DiplomaticStance) (color.Color, string) {
	switch s {
	case faction.StanceWar:
		return ColorRed, "⚔  Savaş"
	case faction.StanceAllied:
		return color.RGBA{60, 220, 60, 255}, "🤝 Müttefik"
	case faction.StanceTrade:
		return ColorGold, "📦 Ticaret"
	default:
		return ColorGray, "— Barış"
	}
}

func scoreColor(score int) color.Color {
	if score >= 50 {
		return color.RGBA{60, 220, 60, 255}
	}
	if score >= 0 {
		return ColorGray
	}
	if score >= -50 {
		return color.RGBA{220, 160, 60, 255}
	}
	return ColorRed
}
