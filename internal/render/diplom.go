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
	diplomRowH = 58.0
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

func minF(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

type rectF struct {
	x float64
	y float64
	w float64
	h float64
}

func listPageRect() rectF {
	w := minF(ScreenWidth-80, 1100)
	h := ScreenHeight - 190
	if h < 240 {
		h = 240
	}
	x := (ScreenWidth - w) / 2
	y := (ScreenHeight - h) / 2
	return rectF{x: x, y: y, w: w, h: h}
}

func offerPageRect() rectF {
	w := minF(ScreenWidth-120, 760)
	h := minF(ScreenHeight-180, 600)
	if h < 360 {
		h = 360
	}
	x := (ScreenWidth - w) / 2
	y := (ScreenHeight - h) / 2
	return rectF{x: x, y: y, w: w, h: h}
}

func listRowStartY() float64 {
	return listPageRect().y + 52
}

func diplomVisibleRows() int {
	r := listPageRect()
	usable := r.h - 70
	rows := int(usable / diplomRowH)
	if rows < 1 {
		return 1
	}
	return rows
}

func diplomMaxScroll(total int) int {
	max := total - diplomVisibleRows()
	if max < 0 {
		return 0
	}
	return max
}

func clampDiplomScroll(total, scroll int) int {
	if scroll < 0 {
		return 0
	}
	max := diplomMaxScroll(total)
	if scroll > max {
		return max
	}
	return scroll
}

func clampDiplomFocus(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func ensureDiplomFocusVisible(total, focus, scroll int) int {
	scroll = clampDiplomScroll(total, scroll)
	visible := diplomVisibleRows()
	if focus < scroll {
		return focus
	}
	if focus >= scroll+visible {
		return focus - visible + 1
	}
	return scroll
}

func diplomActionRect(i int) (x, y, w, h float32) {
	p := offerPageRect()
	btnW := float32(p.w - 40)
	btnH := float32(42)
	gap := float32(12)
	x = float32(p.x + 20)
	y = float32(p.y + 190 + float64(i)*(float64(btnH)+float64(gap)))
	return x, y, btnW, btnH
}

func diplomSendRect() (x, y, w, h float32) {
	p := offerPageRect()
	w = float32((p.w - 52) / 2)
	h = 40
	x = float32(p.x + p.w - 20 - float64(w))
	y = float32(p.y + p.h - 64)
	return x, y, w, h
}

func diplomBackRect() (x, y, w, h float32) {
	p := offerPageRect()
	w = float32((p.w - 52) / 2)
	h = 40
	x = float32(p.x + 20)
	y = float32(p.y + p.h - 64)
	return x, y, w, h
}

// DrawDiplomacyPanel diplomasi panelini çizer.
func DrawDiplomacyPanel(screen *ebiten.Image, gs *state.GameState, focusIdx, scroll, actionFocus int, target faction.FactionID) {
	overlay := ebiten.NewImage(int(ScreenWidth), int(ScreenHeight))
	overlay.Fill(color.RGBA{8, 6, 4, 220})
	screen.DrawImage(overlay, nil)

	DrawTextCentered(screen, "── Diplomasi ──", ScreenWidth/2, 24, FaceLarge, ColorYellow)
	drawDiplomacyCloseButton(screen)

	factions := sortedFactions(gs)
	scroll = clampDiplomScroll(len(factions), scroll)
	focusIdx = clampDiplomFocus(focusIdx, 0, len(factions)-1)
	start := scroll
	end := start + diplomVisibleRows()
	if end > len(factions) {
		end = len(factions)
	}

	if target == "" {
		drawDiplomacyListPage(screen, gs, factions, focusIdx, start, end)
	} else {
		drawDiplomacyOfferPanel(screen, gs, target, actionFocus)
	}

	if target == "" && len(factions) > end-start {
		info := "Liste: " + itoa(start+1) + "-" + itoa(end) + "/" + itoa(len(factions))
		DrawText(screen, info, listPageRect().x+8, listPageRect().y+listPageRect().h-18, FaceSmall, ColorGray)
	}
}

func drawDiplomacyListPage(screen *ebiten.Image, gs *state.GameState, factions []faction.FactionID, focusIdx, start, end int) {
	r := listPageRect()
	vector.FillRect(screen, float32(r.x), float32(r.y), float32(r.w), float32(r.h), color.RGBA{18, 16, 12, 210}, false)
	vector.StrokeRect(screen, float32(r.x), float32(r.y), float32(r.w), float32(r.h), 1, panelBorder, false)
	DrawText(screen, "Devlet seçin", r.x+14, r.y+14, FaceSmall, ColorGray)

	for row, i := 0, start; i < end; i, row = i+1, row+1 {
		fid := factions[i]
		f := gs.Factions[fid]
		rel := gs.Relations[faction.RelationKey(gs.PlayerFactionID, fid)]

		y := listRowStartY() + float64(row)*diplomRowH
		rowCol := color.RGBA{25, 20, 14, 200}
		if i == focusIdx {
			rowCol = color.RGBA{70, 58, 30, 235}
		}
		vector.FillRect(screen, float32(r.x+8), float32(y), float32(r.w-16), float32(diplomRowH-4), rowCol, false)

		fc := color.RGBA{f.Color[0], f.Color[1], f.Color[2], 255}
		vector.FillRect(screen, float32(r.x+8), float32(y), 6, float32(diplomRowH-4), fc, false)

		DrawText(screen, f.NameTR, r.x+16, y+7, FaceMed, ColorWhite)
		regionCount := len(gs.RegionsOwnedBy(fid))
		DrawText(screen, itoa(regionCount)+" bölge", r.x+16, y+29, FaceSmall, ColorGray)

		statusX := r.x + r.w - 240
		if rel != nil {
			stanceCol, stanceTR := stanceDisplay(rel.Stance)
			DrawText(screen, stanceTR, statusX, y+7, FaceMed, stanceCol)
			scoreCol := scoreColor(rel.Score)
			DrawText(screen, "İlişki: "+itoa(rel.Score), statusX, y+29, FaceSmall, scoreCol)
		} else {
			DrawText(screen, "Tarafsız", statusX, y+7, FaceMed, ColorGray)
		}
	}
}

func drawDiplomacyOfferPanel(screen *ebiten.Image, gs *state.GameState, target faction.FactionID, actionFocus int) {
	f := gs.Factions[target]
	if f == nil {
		return
	}
	p := offerPageRect()
	vector.FillRect(screen, float32(p.x), float32(p.y), float32(p.w), float32(p.h), color.RGBA{16, 14, 10, 220}, false)
	vector.StrokeRect(screen, float32(p.x), float32(p.y), float32(p.w), float32(p.h), 1, panelBorder, false)

	DrawText(screen, "Teklif Paneli", p.x+20, p.y+20, FaceLarge, ColorGold)
	bx, by, bw, bh := diplomBackRect()
	vector.FillRect(screen, bx, by, bw, bh, color.RGBA{70, 70, 70, 230}, false)
	vector.StrokeRect(screen, bx, by, bw, bh, 1, panelBorder, false)
	backLabel := "← Geri"
	blw := MeasureText(backLabel, FaceMed)
	DrawText(screen, backLabel, float64(bx)+float64(bw)/2-blw/2, float64(by)+10, FaceMed, ColorWhite)
	DrawText(screen, "Hedef: "+f.NameTR, p.x+20, p.y+52, FaceMed, ColorWhite)

	rel := gs.Relations[faction.RelationKey(gs.PlayerFactionID, target)]
	relScore := 0
	relStance := faction.StancePeace
	if rel != nil {
		relScore = rel.Score
		relStance = rel.Stance
	}
	DrawText(screen, "Durum: "+stanceDisplayText(relStance), p.x+20, p.y+76, FaceSmall, ColorGray)
	DrawText(screen, "İlişki Skoru: "+itoa(relScore), p.x+20, p.y+96, FaceSmall, scoreColor(relScore))

	DrawText(screen, "Teklif Türü", p.x+20, p.y+126, FaceMed, ColorGray)
	for i, da := range diplomActions {
		chance, status := estimateDiplomacyChance(gs, target, da.action)
		bx, by, bw, bh := diplomActionRect(i)
		bg := da.color
		if i != actionFocus {
			bg.A = 170
		}
		vector.FillRect(screen, bx, by, bw, bh, bg, false)
		vector.StrokeRect(screen, bx, by, bw, bh, 1, panelBorder, false)
		DrawText(screen, da.label, float64(bx)+14, float64(by)+7, FaceMed, ColorWhite)
		chanceText := "%" + itoa(chance)
		cw := MeasureText(chanceText, FaceMed)
		DrawText(screen, chanceText, float64(bx)+float64(bw)-cw-14, float64(by)+7, FaceMed, ColorWhite)
		DrawText(screen, status, float64(bx)+14, float64(by)+25, FaceSmall, color.RGBA{235, 230, 210, 230})
	}

	_, lastBY, _, lastBH := diplomActionRect(len(diplomActions) - 1)
	selected := "Seçili teklif: " + diplomActions[actionFocus].label
	slw := MeasureText(selected, FaceSmall)
	selectedY := float64(lastBY + lastBH + 24)
	DrawText(screen, selected, p.x+p.w/2-slw/2, selectedY, FaceSmall, ColorGray)

	sx, sy, sw, sh := diplomSendRect()
	vector.FillRect(screen, sx, sy, sw, sh, color.RGBA{48, 130, 72, 235}, false)
	vector.StrokeRect(screen, sx, sy, sw, sh, 1, panelBorder, false)
	sendLabel := "Teklif Gönder"
	lw := MeasureText(sendLabel, FaceMed)
	DrawText(screen, sendLabel, float64(sx)+float64(sw)/2-lw/2, float64(sy)+10, FaceMed, ColorWhite)
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

func diplomacyListRowAt(gs *state.GameState, scroll int, fx, fy float64) int {
	factions := sortedFactions(gs)
	start := clampDiplomScroll(len(factions), scroll)
	end := start + diplomVisibleRows()
	if end > len(factions) {
		end = len(factions)
	}
	lr := listPageRect()
	for row, i := 0, start; i < end; i, row = i+1, row+1 {
		y := listRowStartY() + float64(row)*diplomRowH
		if fy >= y && fy <= y+diplomRowH-4 && fx >= lr.x+8 && fx <= lr.x+lr.w-8 {
			return i
		}
	}
	return -1
}

// handleDiplomacyInput diplomasi paneli klavye ve fare girişini işler.
func (r *Renderer) handleDiplomacyInput() InputAction {
	factions := sortedFactions(r.gs)
	n := len(factions)
	if n == 0 {
		return InputAction{}
	}
	r.diplomacyScroll = clampDiplomScroll(n, r.diplomacyScroll)
	r.diplomacyFocus = clampDiplomFocus(r.diplomacyFocus, 0, n-1)
	r.diplomacyActionFocus = clampDiplomFocus(r.diplomacyActionFocus, 0, len(diplomActions)-1)
	r.diplomacyScroll = ensureDiplomFocusVisible(n, r.diplomacyFocus, r.diplomacyScroll)

	mx, my := ebiten.CursorPosition()
	fx, fy := float64(mx), float64(my)
	_, wheelY := ebiten.Wheel()
	if wheelY > 0 {
		r.diplomacyScroll--
	}
	if wheelY < 0 {
		r.diplomacyScroll++
	}
	r.diplomacyScroll = clampDiplomScroll(n, r.diplomacyScroll)

	hoverIdx := diplomacyListRowAt(r.gs, r.diplomacyScroll, fx, fy)
	if hoverIdx >= 0 {
		r.diplomacyFocus = hoverIdx
	}

	if r.mouseJustPressed(ebiten.MouseButtonLeft) {
		if diplomacyCloseHit(fx, fy) {
			r.showDiplomacy = false
			r.diplomacyTargetFaction = ""
			return InputAction{}
		}
		if r.diplomacyTargetFaction == "" {
			if hoverIdx >= 0 && r.diplomacyFocus < len(factions) {
				r.diplomacyTargetFaction = factions[r.diplomacyFocus]
				r.diplomacyActionFocus = 0
				return InputAction{}
			}
		} else {
			bx, by, bw, bh := diplomBackRect()
			if fx >= float64(bx) && fx <= float64(bx+bw) && fy >= float64(by) && fy <= float64(by+bh) {
				r.diplomacyTargetFaction = ""
				return InputAction{}
			}
			for j := range diplomActions {
				ax, ay, aw, ah := diplomActionRect(j)
				if fx >= float64(ax) && fx <= float64(ax+aw) && fy >= float64(ay) && fy <= float64(ay+ah) {
					r.diplomacyActionFocus = j
					return InputAction{}
				}
			}
			sx, sy, sw, sh := diplomSendRect()
			if fx >= float64(sx) && fx <= float64(sx+sw) && fy >= float64(sy) && fy <= float64(sy+sh) {
				target := r.diplomacyTargetFaction
				r.showDiplomacy = false
				r.diplomacyTargetFaction = ""
				return InputAction{Kind: diplomActions[r.diplomacyActionFocus].action, TargetFaction: target}
			}
		}
	}

	if r.keyJustPressed(ebiten.KeyArrowDown) && r.diplomacyFocus < n-1 {
		r.diplomacyFocus++
		r.diplomacyScroll = ensureDiplomFocusVisible(n, r.diplomacyFocus, r.diplomacyScroll)
	}
	if r.keyJustPressed(ebiten.KeyArrowUp) && r.diplomacyFocus > 0 {
		r.diplomacyFocus--
		r.diplomacyScroll = ensureDiplomFocusVisible(n, r.diplomacyFocus, r.diplomacyScroll)
	}
	if r.diplomacyTargetFaction != "" {
		if r.keyJustPressed(ebiten.KeyArrowRight) && r.diplomacyActionFocus < len(diplomActions)-1 {
			r.diplomacyActionFocus++
		}
		if r.keyJustPressed(ebiten.KeyArrowLeft) && r.diplomacyActionFocus > 0 {
			r.diplomacyActionFocus--
		}
	}
	if r.keyJustPressed(ebiten.KeyTab) || r.keyJustPressed(ebiten.KeyEscape) {
		if r.diplomacyTargetFaction != "" {
			r.diplomacyTargetFaction = ""
		} else {
			r.showDiplomacy = false
		}
		return InputAction{}
	}
	if r.keyJustPressed(ebiten.KeyEnter) {
		if r.diplomacyTargetFaction == "" {
			if r.diplomacyFocus < len(factions) {
				r.diplomacyTargetFaction = factions[r.diplomacyFocus]
				r.diplomacyActionFocus = 0
				return InputAction{}
			}
		} else {
			target := r.diplomacyTargetFaction
			r.showDiplomacy = false
			r.diplomacyTargetFaction = ""
			return InputAction{Kind: diplomActions[r.diplomacyActionFocus].action, TargetFaction: target}
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
		if f := gs.Factions[fid]; f == nil || f.IsEliminated {
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
		return ColorRed, "WAR Savas"
	case faction.StanceAllied:
		return color.RGBA{60, 220, 60, 255}, "ALLY Muttefik"
	case faction.StanceTrade:
		return ColorGold, "TRADE Ticaret"
	default:
		return ColorGray, "Baris"
	}
}

func stanceDisplayText(s faction.DiplomaticStance) string {
	_, label := stanceDisplay(s)
	return label
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

func estimateDiplomacyChance(gs *state.GameState, target faction.FactionID, action ActionKind) (int, string) {
	rel := gs.Relations[faction.RelationKey(gs.PlayerFactionID, target)]
	score := 0
	stance := faction.StancePeace
	if rel != nil {
		score = rel.Score
		stance = rel.Stance
	}
	playerRegions := len(gs.RegionsOwnedBy(gs.PlayerFactionID))
	targetRegions := len(gs.RegionsOwnedBy(target))
	regionDelta := playerRegions - targetRegions

	chance := 50 + score/2
	switch action {
	case ActionDeclareWar:
		if stance == faction.StanceWar {
			chance = 0
		} else {
			chance = 100
		}
	case ActionProposePeace:
		if stance != faction.StanceWar {
			chance = 0
		} else {
			chance = 35 + (-score / 2) + regionDelta*4
		}
	case ActionProposeAlliance:
		if stance == faction.StanceWar {
			chance = 0
		} else {
			chance = 15 + score + regionDelta*2
		}
	case ActionProposeTrade:
		if stance == faction.StanceWar {
			chance = 0
		} else {
			chance = 40 + score + regionDelta
		}
	}
	if chance < 0 {
		chance = 0
	}
	if chance > 100 {
		chance = 100
	}
	switch {
	case chance == 0:
		return chance, "Geçersiz / Mümkün değil"
	case chance >= 75:
		return chance, "Yüksek kabul olasılığı"
	case chance >= 45:
		return chance, "Orta kabul olasılığı"
	default:
		return chance, "Düşük kabul olasılığı"
	}
}
