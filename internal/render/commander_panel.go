package render

import (
	"fmt"
	"image/color"
	"path/filepath"
	"strings"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/state"
	gameui "mapp-game-go/internal/ui"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	commanderPanelW       = 760.0
	commanderPanelH       = 500.0
	commanderPanelRowH    = 88.0
	commanderPanelListW   = 330.0
	commanderPanelButtonW = 150.0
	commanderPanelButtonH = 34.0
)

var (
	commanderPortraitCache = map[string]*ebiten.Image{}
	commanderPortraitTried = map[string]bool{}
)

func resetCommanderPortraitCache() {
	commanderPortraitCache = map[string]*ebiten.Image{}
	commanderPortraitTried = map[string]bool{}
}

func commanderPortraitAsset(asset string) *ebiten.Image {
	if asset == "" || ActiveScenarioPath == "" {
		return nil
	}
	path := commanderPortraitPath(asset)
	if path == "" {
		return nil
	}
	if img, ok := commanderPortraitCache[path]; ok {
		return img
	}
	if commanderPortraitTried[path] {
		return nil
	}
	commanderPortraitTried[path] = true
	img := tryLoadImage(path)
	if img != nil {
		commanderPortraitCache[path] = img
	}
	return img
}

func commanderPortrait(commander *army.Commander) *ebiten.Image {
	if commander == nil {
		return nil
	}
	return commanderPortraitAsset(commander.PortraitAsset)
}

func commanderPortraitPath(asset string) string {
	if asset == "" || ActiveScenarioPath == "" {
		return ""
	}
	clean := filepath.Clean(filepath.FromSlash(asset))
	slash := filepath.ToSlash(clean)
	if clean == "." || clean == "" || filepath.IsAbs(clean) || slash == ".." || strings.HasPrefix(slash, "../") {
		return ""
	}
	if strings.ContainsRune(slash, '/') {
		return filepath.Join(ActiveScenarioPath, "sprites", clean)
	}
	return filepath.Join(ActiveScenarioPath, "sprites", "commanders", clean)
}

func drawCommanderPortraitAsset(screen *ebiten.Image, asset string, x, y, w, h float64) {
	vector.FillRect(screen, float32(x), float32(y), float32(w), float32(h), color.RGBA{20, 16, 10, 230}, false)
	if img := commanderPortraitAsset(asset); img != nil {
		bounds := img.Bounds()
		if bounds.Dx() > 0 && bounds.Dy() > 0 {
			scale := w / float64(bounds.Dx())
			if heightScale := h / float64(bounds.Dy()); heightScale < scale {
				scale = heightScale
			}
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Scale(scale, scale)
			op.GeoM.Translate(x+(w-float64(bounds.Dx())*scale)/2, y+(h-float64(bounds.Dy())*scale)/2)
			screen.DrawImage(img, op)
		}
	} else {
		DrawTextCentered(screen, "Yok", x+w/2, y+h/2-8, FaceSmall, ColorGray)
	}
	vector.StrokeRect(screen, float32(x), float32(y), float32(w), float32(h), 1, color.RGBA{160, 120, 40, 220}, false)
}

func drawCommanderPortrait(screen *ebiten.Image, commander *army.Commander, x, y, w, h float64) {
	asset := ""
	if commander != nil {
		asset = commander.PortraitAsset
	}
	drawCommanderPortraitAsset(screen, asset, x, y, w, h)
}

func commanderPanelRect() gameui.Rect {
	return gameui.AnchorRect(
		gameui.Rect{W: ScreenWidth, H: ScreenHeight},
		commanderPanelW,
		commanderPanelH,
		gameui.AnchorCenter,
		gameui.AnchorMiddle,
		0,
		0,
	)
}

func commanderPanelCloseButton() gameui.Button {
	panel := commanderPanelRect()
	return gameui.NewButton(panel.X+panel.W-44, panel.Y+12, 30, 30, "").WithIcon(gameui.IconClose)
}

func commanderPanelUnassignButton(gs *state.GameState, aid army.ArmyID) (gameui.Button, bool) {
	if gs == nil {
		return gameui.Button{}, false
	}
	current := gs.Armies[aid]
	if current == nil || current.Commander == nil {
		return gameui.Button{}, false
	}
	panel := commanderPanelRect()
	return gameui.NewButton(panel.X+panel.W-commanderPanelButtonW-24, panel.Y+panel.H-54, commanderPanelButtonW, commanderPanelButtonH, "Komutanı Ayır"), true
}

func commanderPanelUnassignEmbarkedButton(gs *state.GameState, aid army.ArmyID) (gameui.Button, bool) {
	if gs == nil {
		return gameui.Button{}, false
	}
	current := gs.Armies[aid]
	if current == nil || current.EmbarkedCommander == nil {
		return gameui.Button{}, false
	}
	panel := commanderPanelRect()
	return gameui.NewButton(panel.X+24, panel.Y+panel.H-54, commanderPanelButtonW, commanderPanelButtonH, "Taşınanı Ayır"), true
}

func commanderPanelRow(index int) gameui.Rect {
	panel := commanderPanelRect()
	return gameui.Rect{
		X: panel.X + 24,
		Y: panel.Y + 104 + float64(index)*commanderPanelRowH,
		W: commanderPanelListW,
		H: commanderPanelRowH - 6,
	}
}

func (r *Renderer) DrawCommanderPanel(screen *ebiten.Image) {
	if r == nil || !r.showCommanderPanel || r.gs == nil {
		return
	}
	current := r.gs.Armies[r.commanderPanelArmy]
	if current == nil || current.OwnerID != string(r.gs.PlayerFactionID) {
		return
	}
	panel := commanderPanelRect()
	vector.FillRect(screen, 0, 0, float32(ScreenWidth), float32(ScreenHeight), color.RGBA{0, 0, 0, 185}, false)
	vector.FillRect(screen, float32(panel.X), float32(panel.Y), float32(panel.W), float32(panel.H), panelBg, false)
	drawPanelBorder(screen, float32(panel.X), float32(panel.Y), float32(panel.W), float32(panel.H))
	vector.FillRect(screen, float32(panel.X), float32(panel.Y), float32(panel.W), 3, panelBorder, false)

	DrawText(screen, "Komutan Atama", panel.X+24, panel.Y+20, FaceLarge, ColorYellow)
	DrawText(screen, "Komutan seçerek seçili orduya ata.", panel.X+24, panel.Y+48, FaceSmall, ColorGray)
	gameui.DrawButton(screen, commanderPanelCloseButton(), gameui.ButtonStyle{
		BG: panelBg, Border: panelBorder, Text: ColorWhite, BorderWidth: 1,
	}, sharedTextRenderer{})

	vector.StrokeLine(screen, float32(panel.X+commanderPanelListW+48), float32(panel.Y+84), float32(panel.X+commanderPanelListW+48), float32(panel.Y+panel.H-24), 1, panelBorder, false)
	DrawText(screen, "Boştaki Komutanlar — seçim için tıkla", panel.X+24, panel.Y+82, FaceMed, ColorGold)

	available := r.gs.AvailableCommanders(current.OwnerID)
	if len(available) == 0 {
		DrawText(screen, "Boşta komutan yok.", panel.X+24, panel.Y+122, FaceSmall, ColorGray)
	} else {
		for i, commander := range available {
			row := commanderPanelRow(i)
			rowBG := color.RGBA{35, 26, 14, 210}
			rowBorder := color.RGBA{100, 75, 30, 180}
			if i == r.commanderPanelFocus {
				rowBG = color.RGBA{75, 54, 20, 235}
				rowBorder = ColorGold
			}
			vector.FillRect(screen, float32(row.X), float32(row.Y), float32(row.W), float32(row.H), rowBG, false)
			vector.StrokeRect(screen, float32(row.X), float32(row.Y), float32(row.W), float32(row.H), 1, rowBorder, false)
			drawCommanderPortrait(screen, commander, row.X+8, row.Y+8, 64, 64)
			textX := row.X + 84
			DrawText(screen, commander.Name, textX, row.Y+12, FaceSmall, ColorWhite)
			DrawText(screen, fmt.Sprintf("Seviye %d  |  %d XP", commander.Level, commander.Experience), textX, row.Y+36, FaceSmall, ColorGray)
			drawCommanderTraitBadges(screen, commander, textX, row.Y+56, row.W-(textX-row.X)-12, commanderTraitBadgeOptions{MaxRows: 1})
		}
	}

	r.drawCommanderDetail(screen, current)
}

func (r *Renderer) drawCommanderDetail(screen *ebiten.Image, current *army.Army) {
	panel := commanderPanelRect()
	x := panel.X + commanderPanelListW + 78
	DrawText(screen, "Seçili Ordu", x, panel.Y+82, FaceMed, ColorGold)
	DrawText(screen, fmt.Sprintf("Birim: %d/%d", len(current.Units), army.MaxArmySize), x, panel.Y+112, FaceSmall, ColorWhite)
	detailCommander := current.Commander
	if detailCommander == nil {
		detailCommander = current.EmbarkedCommander
	}
	if detailCommander != nil {
		drawCommanderPortrait(screen, detailCommander, panel.X+panel.W-132, panel.Y+92, 96, 96)
	}

	if current.Commander == nil && current.EmbarkedCommander == nil {
		DrawText(screen, "Bu orduda komutan yok.", x, panel.Y+154, FaceSmall, ColorGray)
		DrawText(screen, "Soldaki listeden bir komutan seç.", x, panel.Y+178, FaceSmall, ColorGray)
	} else {
		y := panel.Y + 150
		if current.Commander != nil {
			y = drawCommanderProfile(screen, x, y, "Filo Komutanı", current.Commander)
		}
		if current.EmbarkedCommander != nil {
			drawCommanderProfile(screen, x, y, "Taşınan Kara Komutanı", current.EmbarkedCommander)
		}
	}

	if btn, ok := commanderPanelUnassignEmbarkedButton(r.gs, current.ID); ok {
		gameui.DrawButton(screen, btn, gameui.ButtonStyle{BG: color.RGBA{80, 35, 25, 230}, Border: color.RGBA{190, 90, 65, 255}, Text: ColorWhite, BorderWidth: 1}, sharedTextRenderer{})
	}
	if btn, ok := commanderPanelUnassignButton(r.gs, current.ID); ok {
		gameui.DrawButton(screen, btn, gameui.ButtonStyle{BG: color.RGBA{80, 35, 25, 230}, Border: color.RGBA{190, 90, 65, 255}, Text: ColorWhite, BorderWidth: 1}, sharedTextRenderer{})
	}
}

func commanderEffectSummary(commander *army.Commander) string {
	if commander == nil {
		return "Katkı yok."
	}
	parts := make([]string, 0, 5)
	if attack := commander.AttackModifier(); attack > 0 {
		parts = append(parts, fmt.Sprintf("Saldırı +%.0f%%", attack*100))
	}
	if defense := commander.DefenseModifier(); defense > 0 {
		parts = append(parts, fmt.Sprintf("Savunma +%.0f%%", defense*100))
	}
	if morale := commander.MoraleModifier(); morale > 0 {
		parts = append(parts, fmt.Sprintf("Moral +%.0f%%", morale*100))
	}
	if move := commander.MoveBonus(); move > 0 {
		parts = append(parts, fmt.Sprintf("Hareket +%d", move))
	}
	progress, breach := commander.SiegeBonuses()
	if progress > 0 || breach > 0 {
		parts = append(parts, fmt.Sprintf("Kuşatma +%d ilerleme / +%d gedik", progress, breach))
	}
	if len(parts) == 0 {
		return "Katkı yok."
	}
	return strings.Join(parts, "  |  ")
}

func commanderCombatEffectSummary(commander *army.Commander) string {
	if commander == nil {
		return "Savaş katkısı yok."
	}
	parts := make([]string, 0, 2)
	if attack := commander.AttackModifier(); attack > 0 {
		parts = append(parts, fmt.Sprintf("Saldırı +%.0f%%", attack*100))
	}
	if defense := commander.DefenseModifier(); defense > 0 {
		parts = append(parts, fmt.Sprintf("Savunma +%.0f%%", defense*100))
	}
	if len(parts) == 0 {
		return "Savaş katkısı yok."
	}
	return strings.Join(parts, "  |  ")
}

func commanderBattleEffectSummary(commander *army.Commander) string {
	if commander == nil {
		return "Muharebe katkısı yok."
	}
	parts := make([]string, 0, 3)
	if attack := commander.AttackModifier(); attack > 0 {
		parts = append(parts, fmt.Sprintf("Saldırı +%.0f%%", attack*100))
	}
	if defense := commander.DefenseModifier(); defense > 0 {
		parts = append(parts, fmt.Sprintf("Savunma +%.0f%%", defense*100))
	}
	if morale := commander.MoraleModifier(); morale > 0 {
		parts = append(parts, fmt.Sprintf("Moral +%.0f%%", morale*100))
	}
	if len(parts) == 0 {
		return "Muharebe katkısı yok."
	}
	return strings.Join(parts, "  |  ")
}

func commanderOperationalEffectSummary(commander *army.Commander) string {
	if commander == nil {
		return "Operasyon katkısı yok."
	}
	parts := make([]string, 0, 3)
	if morale := commander.MoraleModifier(); morale > 0 {
		parts = append(parts, fmt.Sprintf("Moral +%.0f%%", morale*100))
	}
	if move := commander.MoveBonus(); move > 0 {
		parts = append(parts, fmt.Sprintf("Hareket +%d", move))
	}
	progress, breach := commander.SiegeBonuses()
	if progress > 0 || breach > 0 {
		parts = append(parts, fmt.Sprintf("Kuşatma +%d/+%d", progress, breach))
	}
	if len(parts) == 0 {
		return "Operasyon katkısı yok."
	}
	return strings.Join(parts, "  |  ")
}

func commanderBattlePlanSummary(role string, commander *army.Commander) string {
	role = strings.TrimSpace(role)
	if commander == nil {
		return role + ": Yok"
	}
	return role + ": " + commander.Name + "  |  " + commanderBattleEffectSummary(commander)
}

func commanderSiegeSummary(commander *army.Commander) string {
	if commander == nil {
		return "Komutan yok; ek moral, hareket veya kuşatma bonusu yok."
	}
	return "Komutan: " + commander.Name + "  |  " + commanderOperationalEffectSummary(commander)
}

func drawCommanderProfile(screen *ebiten.Image, x, y float64, role string, commander *army.Commander) float64 {
	DrawText(screen, role, x, y, FaceSmall, ColorGold)
	DrawText(screen, commander.Name, x, y+20, FaceMed, ColorWhite)
	DrawText(screen, fmt.Sprintf("Seviye %d  |  %d XP  |  Savaş %d  |  Zafer %d", commander.Level, commander.Experience, commander.Battles, commander.Victories), x, y+42, FaceSmall, ColorGray)
	DrawText(screen, commanderCombatEffectSummary(commander), x, y+62, FaceSmall, color.RGBA{140, 210, 150, 255})
	DrawText(screen, commanderOperationalEffectSummary(commander), x, y+82, FaceSmall, color.RGBA{145, 185, 220, 255})
	DrawText(screen, "Uzmanlıklar", x, y+104, FaceSmall, ColorGold)
	bottomY := drawCommanderTraitBadges(screen, commander, x, y+122, 280, commanderTraitBadgeOptions{})
	return bottomY + 14
}

func (r *Renderer) handleCommanderPanelInput() InputAction {
	if r == nil || !r.showCommanderPanel {
		return InputAction{}
	}
	if r.keyJustPressed(ebiten.KeyEscape) {
		r.CloseCommanderPanel()
		return InputAction{}
	}
	mx, my := ebiten.CursorPosition()
	fx, fy := float64(mx), float64(my)
	if commanderPanelCloseButton().HitTest(fx, fy) && r.mouseJustPressed(ebiten.MouseButtonLeft) {
		r.CloseCommanderPanel()
		return InputAction{}
	}
	current := r.gs.Armies[r.commanderPanelArmy]
	if current == nil {
		r.CloseCommanderPanel()
		return InputAction{}
	}
	if btn, ok := commanderPanelUnassignButton(r.gs, current.ID); ok && btn.HitTest(fx, fy) && r.mouseJustPressed(ebiten.MouseButtonLeft) {
		return InputAction{Kind: ActionUnassignCommander, ArmyID: current.ID}
	}
	if btn, ok := commanderPanelUnassignEmbarkedButton(r.gs, current.ID); ok && btn.HitTest(fx, fy) && r.mouseJustPressed(ebiten.MouseButtonLeft) {
		return InputAction{Kind: ActionUnassignEmbarkedCommander, ArmyID: current.ID}
	}
	available := r.gs.AvailableCommanders(current.OwnerID)
	if r.keyJustPressed(ebiten.KeyArrowDown) && len(available) > 0 {
		r.commanderPanelFocus = (r.commanderPanelFocus + 1) % len(available)
		return InputAction{}
	}
	if r.keyJustPressed(ebiten.KeyArrowUp) && len(available) > 0 {
		r.commanderPanelFocus--
		if r.commanderPanelFocus < 0 {
			r.commanderPanelFocus = len(available) - 1
		}
		return InputAction{}
	}
	if (r.keyJustPressed(ebiten.KeyEnter) || r.keyJustPressed(ebiten.KeySpace)) && len(available) > 0 {
		return InputAction{Kind: ActionAssignCommander, ArmyID: current.ID, CommanderID: available[r.commanderPanelFocus].ID}
	}
	if r.mouseJustPressed(ebiten.MouseButtonLeft) {
		for i, commander := range available {
			if commanderPanelRow(i).Hit(fx, fy) {
				r.commanderPanelFocus = i
				return InputAction{Kind: ActionAssignCommander, ArmyID: current.ID, CommanderID: commander.ID}
			}
		}
	}
	return InputAction{}
}

func (r *Renderer) commanderPanelHovering(fx, fy float64) bool {
	if r == nil || !r.showCommanderPanel || r.gs == nil {
		return false
	}
	if commanderPanelCloseButton().HitTest(fx, fy) {
		return true
	}
	current := r.gs.Armies[r.commanderPanelArmy]
	if current == nil {
		return false
	}
	if btn, ok := commanderPanelUnassignButton(r.gs, current.ID); ok && btn.HitTest(fx, fy) {
		return true
	}
	if btn, ok := commanderPanelUnassignEmbarkedButton(r.gs, current.ID); ok && btn.HitTest(fx, fy) {
		return true
	}
	available := r.gs.AvailableCommanders(current.OwnerID)
	for i := range available {
		if commanderPanelRow(i).Hit(fx, fy) {
			return true
		}
	}
	return false
}

func (r *Renderer) OpenCommanderPanel(aid army.ArmyID) {
	if r == nil || r.gs == nil {
		return
	}
	current := r.gs.Armies[aid]
	if current == nil || current.OwnerID != string(r.gs.PlayerFactionID) {
		return
	}
	r.showCommanderPanel = true
	r.commanderPanelArmy = aid
	r.commanderPanelFocus = 0
	r.gs.SyncCommanderLinks()
}

func (r *Renderer) CloseCommanderPanel() {
	if r == nil {
		return
	}
	r.showCommanderPanel = false
	r.commanderPanelArmy = ""
	r.commanderPanelFocus = 0
}
