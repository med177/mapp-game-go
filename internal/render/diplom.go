package render

import (
	"image/color"
	"sort"

	"mapp-game-go/internal/diplomacy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	gameui "mapp-game-go/internal/ui"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
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
	{diplomacy.ActionLabelTR(diplomacy.ActionDeclareWar), color.RGBA{180, 50, 50, 220}, ActionDeclareWar},
	{diplomacy.ActionLabelTR(diplomacy.ActionProposePeace), color.RGBA{50, 120, 180, 220}, ActionProposePeace},
	{diplomacy.ActionLabelTR(diplomacy.ActionProposeAlliance), color.RGBA{50, 160, 80, 220}, ActionProposeAlliance},
	{diplomacy.ActionLabelTR(diplomacy.ActionProposeTrade), color.RGBA{160, 130, 50, 220}, ActionProposeTrade},
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

type diplomacyActionButton struct {
	Index  int
	Button gameui.Button
}

func buildDiplomacyCloseButton() gameui.Button {
	x, y, w, h := diplomacyCloseRect()
	return gameui.NewButton(float64(x), float64(y), float64(w), float64(h), "X")
}

func buildDiplomacyListView(gs *state.GameState, focusIdx, scroll int) gameui.ListView {
	factions := sortedFactions(gs)
	items := make([]string, 0, len(factions))
	for _, fid := range factions {
		if f := gs.Factions[fid]; f != nil {
			items = append(items, f.NameTR)
		}
	}
	r := listPageRect()
	list := gameui.NewListView(r.x+8, listRowStartY(), r.w-16, float64(diplomVisibleRows())*diplomRowH, diplomRowH, diplomVisibleRows(), items)
	list.Scroll = clampDiplomScroll(len(items), scroll)
	list.Selected = clampDiplomFocus(focusIdx, 0, len(items)-1)
	return list
}

func buildDiplomacyBackButton() gameui.Button {
	x, y, w, h := diplomBackRect()
	return gameui.NewButton(float64(x), float64(y), float64(w), float64(h), "← Geri")
}

func buildDiplomacySendButton() gameui.Button {
	x, y, w, h := diplomSendRect()
	return gameui.NewButton(float64(x), float64(y), float64(w), float64(h), "Teklif Gönder")
}

func buildDiplomacyActionButtons() []diplomacyActionButton {
	out := make([]diplomacyActionButton, 0, len(diplomActions))
	for i, da := range diplomActions {
		x, y, w, h := diplomActionRect(i)
		out = append(out, diplomacyActionButton{
			Index:  i,
			Button: gameui.NewButton(float64(x), float64(y), float64(w), float64(h), da.label),
		})
	}
	return out
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

	list := buildDiplomacyListView(gs, focusIdx, start)
	for row, i := 0, list.Scroll; i < end; i, row = i+1, row+1 {
		fid := factions[i]
		f := gs.Factions[fid]
		rel := gs.Relations[faction.RelationKey(gs.PlayerFactionID, fid)]

		y := list.Rect.Y + float64(row)*diplomRowH
		rowCol := color.RGBA{25, 20, 14, 200}
		if i == focusIdx {
			rowCol = color.RGBA{70, 58, 30, 235}
		}
		vector.FillRect(screen, float32(list.Rect.X), float32(y), float32(list.Rect.W), float32(diplomRowH-4), rowCol, false)

		fc := color.RGBA{f.Color[0], f.Color[1], f.Color[2], 255}
		vector.FillRect(screen, float32(list.Rect.X), float32(y), 6, float32(diplomRowH-4), fc, false)

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
	drawDiplomacyButton(screen, buildDiplomacyBackButton(), color.RGBA{70, 70, 70, 230}, panelBorder, FaceMed, 10)
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
	for _, btn := range buildDiplomacyActionButtons() {
		i := btn.Index
		da := diplomActions[i]
		chance, status := estimateDiplomacyChance(gs, target, da.action)
		bg := da.color
		if i != actionFocus {
			bg.A = 170
		}
		drawDiplomacyButton(screen, btn.Button, bg, panelBorder, FaceMed, 7)
		bx, by, bw, _ := diplomActionRect(i)
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

	drawDiplomacyButton(screen, buildDiplomacySendButton(), color.RGBA{48, 130, 72, 235}, panelBorder, FaceMed, 10)
}

func diplomacyCloseRect() (x, y, w, h float32) {
	return float32(ScreenWidth) - 58, 20, 30, 26
}

func drawDiplomacyCloseButton(screen *ebiten.Image) {
	drawDiplomacyButton(screen, buildDiplomacyCloseButton(), color.RGBA{45, 34, 25, 230}, panelBorder, FaceSmall, 6)
}

func drawDiplomacyButton(screen *ebiten.Image, btn gameui.Button, bg color.RGBA, border color.Color, face *text.GoTextFace, textOffsetY float64) {
	vector.FillRect(screen, float32(btn.X), float32(btn.Y), float32(btn.W), float32(btn.H), bg, false)
	vector.StrokeRect(screen, float32(btn.X), float32(btn.Y), float32(btn.W), float32(btn.H), 1, border, false)
	tw := MeasureText(btn.Label, face)
	DrawText(screen, btn.Label, btn.X+btn.W/2-tw/2, btn.Y+textOffsetY, face, ColorWhite)
}

// handleDiplomacyInput diplomasi paneli klavye ve fare girişini işler.
func (r *Renderer) handleDiplomacyInput(input gameui.InputState) InputAction {
	factions := sortedFactions(r.gs)
	n := len(factions)
	if n == 0 {
		return InputAction{}
	}
	r.diplomacyScroll = clampDiplomScroll(n, r.diplomacyScroll)
	r.diplomacyFocus = clampDiplomFocus(r.diplomacyFocus, 0, n-1)
	r.diplomacyActionFocus = clampDiplomFocus(r.diplomacyActionFocus, 0, len(diplomActions)-1)
	r.diplomacyScroll = ensureDiplomFocusVisible(n, r.diplomacyFocus, r.diplomacyScroll)
	if input.LeftJustPressed && !diplomacyPanelPointerHit(input.MouseX, input.MouseY, r.gs, r.diplomacyFocus, r.diplomacyScroll, r.diplomacyActionFocus, r.diplomacyTargetFaction) {
		r.showDiplomacy = false
		r.diplomacyTargetFaction = ""
		return InputAction{}
	}
	if buildDiplomacyCloseButton().HandleInput(input) {
		r.showDiplomacy = false
		r.diplomacyTargetFaction = ""
		return InputAction{}
	}
	if r.diplomacyTargetFaction == "" {
		list := buildDiplomacyListView(r.gs, r.diplomacyFocus, r.diplomacyScroll)
		if list.HandleInput(input) {
			r.diplomacyScroll = list.Scroll
			if list.Selected >= 0 {
				r.diplomacyFocus = list.Selected
				if input.LeftJustPressed && r.diplomacyFocus < len(factions) {
					r.diplomacyTargetFaction = factions[r.diplomacyFocus]
					r.diplomacyActionFocus = 0
				}
			}
			return InputAction{}
		}
	} else {
		if buildDiplomacyBackButton().HandleInput(input) {
			r.diplomacyTargetFaction = ""
			return InputAction{}
		}
		for _, btn := range buildDiplomacyActionButtons() {
			if btn.Button.HandleInput(input) {
				r.diplomacyActionFocus = btn.Index
				return InputAction{}
			}
		}
		if buildDiplomacySendButton().HandleInput(input) {
			target := r.diplomacyTargetFaction
			r.showDiplomacy = false
			r.diplomacyTargetFaction = ""
			return InputAction{Kind: diplomActions[r.diplomacyActionFocus].action, TargetFaction: target}
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

func diplomacyPanelPointerHit(mx, my float64, gs *state.GameState, focusIdx, scroll, actionFocus int, target faction.FactionID) bool {
	if buildDiplomacyCloseButton().HitTest(mx, my) {
		return true
	}
	if target != "" {
		if buildDiplomacyBackButton().HitTest(mx, my) || buildDiplomacySendButton().HitTest(mx, my) {
			return true
		}
		for _, btn := range buildDiplomacyActionButtons() {
			if btn.Button.HitTest(mx, my) {
				return true
			}
		}
		p := offerPageRect()
		return gameui.Rect{X: p.x, Y: p.y, W: p.w, H: p.h}.Hit(mx, my)
	}
	list := buildDiplomacyListView(gs, focusIdx, scroll)
	if list.HitTest(mx, my) {
		return true
	}
	p := listPageRect()
	return gameui.Rect{X: p.x, Y: p.y, W: p.w, H: p.h}.Hit(mx, my)
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
		return ColorRed, faction.DiplomaticStanceBadgeTR(s)
	case faction.StanceAllied:
		return color.RGBA{60, 220, 60, 255}, faction.DiplomaticStanceBadgeTR(s)
	case faction.StanceTrade:
		return ColorGold, faction.DiplomaticStanceBadgeTR(s)
	default:
		return ColorGray, faction.DiplomaticStanceLabelTR(faction.NormalizeStance(s))
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
