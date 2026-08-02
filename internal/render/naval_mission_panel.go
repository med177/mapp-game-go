package render

import (
	"image/color"
	"sort"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/state"
	gameui "mapp-game-go/internal/ui"
	"mapp-game-go/internal/world"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	navalMissionPanelW           = float32(600)
	navalMissionPanelHeaderH     = float32(70)
	navalMissionPanelRowH        = float32(80)
	navalMissionPanelFooterH     = float32(56)
	navalMissionPanelVisibleRows = 7
	navalMissionButtonW          = float64(78)
	navalMissionButtonGap        = float64(6)
)

type navalMissionPanelLayout struct {
	panelX, panelY, panelW, panelH float32
	rowX, rowY, rowW               float32
	close                          gameui.Button
	clear                          gameui.Button
}

type navalMissionOption struct {
	kind        army.NavalMissionKind
	targetFleet army.ArmyID
	label       string
	description string
	effect      string
}

func navalMissionPanelLayoutFor(rowCount int) navalMissionPanelLayout {
	visibleRows := rowCount
	if visibleRows < 1 {
		visibleRows = 1
	}
	if visibleRows > navalMissionPanelVisibleRows {
		visibleRows = navalMissionPanelVisibleRows
	}
	panelH := navalMissionPanelHeaderH + float32(visibleRows)*navalMissionPanelRowH + navalMissionPanelFooterH
	panelX := float32(ScreenWidth)/2 - navalMissionPanelW/2
	panelY := float32(ScreenHeight)/2 - panelH/2
	close := gameui.NewButton(float64(panelX+navalMissionPanelW-42), float64(panelY+10), 28, 28, "").WithIcon(gameui.IconClose)
	close.IconSize = 13
	clear := gameui.NewButton(
		float64(panelX+18),
		float64(panelY+panelH-navalMissionPanelFooterH+11),
		commanderPanelButtonW,
		commanderPanelButtonH,
		"Görevi Kaldır",
	)
	return navalMissionPanelLayout{
		panelX: panelX, panelY: panelY, panelW: navalMissionPanelW, panelH: panelH,
		rowX: panelX + 18, rowY: panelY + navalMissionPanelHeaderH,
		rowW: navalMissionPanelW - 36, close: close, clear: clear,
	}
}

func navalMissionButtonRect(layout armyPanelLayout, hasMerchant bool) gameui.Rect {
	footerY := layout.panelY + layout.panelH - siegeFooterH
	buttonX := layout.gridX + 8
	if hasMerchant {
		buttonX += 132 + 8
	}
	return gameui.Rect{
		X: float64(buttonX), Y: float64(footerY + 4),
		W: navalMissionButtonW, H: float64(merchantRouteFooterButtonH),
	}
}

func playerNavalMissionEligible(gs *state.GameState, fleet *army.Army) bool {
	if gs == nil || fleet == nil || fleet.OwnerID != string(gs.PlayerFactionID) || !fleet.IsNaval {
		return false
	}
	return fleetHasWarshipUI(gs, fleet)
}

func fleetHasWarshipUI(gs *state.GameState, fleet *army.Army) bool {
	if gs == nil || fleet == nil {
		return false
	}
	for _, unit := range fleet.Units {
		if unitType := gs.UnitTypes[unit.TypeID]; unitType != nil && unitType.Category == army.CategoryNavalWar {
			return true
		}
	}
	return false
}

func navalMissionButtonHit(fx, fy float64, gs *state.GameState, aid army.ArmyID) bool {
	if gs == nil || aid == "" {
		return false
	}
	fleet := gs.Armies[aid]
	if !playerNavalMissionEligible(gs, fleet) {
		return false
	}
	return navalMissionButtonRect(armyPanelGeometry(), armyHasMerchantShip(gs, fleet)).Hit(fx, fy)
}

func drawNavalMissionFooter(screen *ebiten.Image, gs *state.GameState, fleet *army.Army, layout armyPanelLayout) {
	if !playerNavalMissionEligible(gs, fleet) {
		return
	}
	hasMerchant := armyHasMerchantShip(gs, fleet)
	rect := navalMissionButtonRect(layout, hasMerchant)
	button := gameui.NewButton(rect.X, rect.Y, rect.W, rect.H, "GÖREV")
	buttonBG := color.RGBA{50, 35, 12, 220}
	buttonBorder := color.RGBA{160, 120, 40, 220}
	if fleet.NavalMission != nil {
		buttonBG = color.RGBA{30, 70, 42, 235}
		buttonBorder = color.RGBA{105, 220, 145, 235}
	}
	gameui.DrawButton(screen, button, gameui.ButtonStyle{
		BG: buttonBG, Border: buttonBorder, Text: ColorGold, BorderWidth: 1, TextVariant: gameui.TextSmall,
	}, sharedTextRenderer{})

	status := "Görev yok"
	statusColor := color.RGBA{180, 180, 180, 220}
	if fleet.NavalMission != nil {
		status = "Görev: " + navalMissionLabelTR(fleet.NavalMission.Kind)
		if fleet.NavalMission.TargetRegionID != "" {
			if region := gs.Regions[fleet.NavalMission.TargetRegionID]; region != nil && region.NameTR != "" {
				status += " → " + region.NameTR
			}
		}
		statusColor = color.RGBA{160, 230, 175, 235}
	}
	footer := armyPanelFooterLayoutFor(gs, fleet, layout)
	if footer.missionStatus.W > 0 {
		drawArmyPanelFooterCenteredText(screen, footer.missionStatus, status, statusColor)
	}
}

func navalMissionOptions(gs *state.GameState, fleet *army.Army) []navalMissionOption {
	if !playerNavalMissionEligible(gs, fleet) {
		return nil
	}
	options := make([]navalMissionOption, 0, 6)
	if fleetHasWarshipUI(gs, fleet) {
		if fleet.IsAtSea() {
			options = append(options, navalMissionOption{
				kind: army.NavalMissionPatrol, label: "Devriye", description: "Bulunduğun denizi gözetle; görevli düşman abluka filosunu yakala.", effect: "Etki: aynı denizdeki abluka filosuyla otomatik savaşır; ticaret ve lojistik kesintisini dengeler.",
			})
			if gs.IsValidNavalBlockadeTarget(fleet, fleet.RegionID) {
				options = append(options, navalMissionOption{
					kind: army.NavalMissionBlockade, label: "Abluka", description: "Bulunduğun denizdeki düşman ticaretini baskıla; tek başına savaş başlatmaz.", effect: "Etki: savaş gemisi başına -%50 ticaret; azami -%100.",
				})
			}
		}
		for _, candidate := range playerTransportFleets(gs, fleet.ID) {
			options = append(options, navalMissionOption{
				kind: army.NavalMissionEscort, targetFleet: candidate.ID,
				label: "Escort → " + string(candidate.ID), description: "Yalnız seçili nakliye filosunu koru.", effect: "Etki: aynı denizde nakliyeye +%15 deniz savunması; azami +%30.",
			})
		}
	}
	return options
}

func playerTransportFleets(gs *state.GameState, exclude army.ArmyID) []*army.Army {
	if gs == nil {
		return nil
	}
	ids := make([]army.ArmyID, 0, len(gs.Armies))
	for id, fleet := range gs.Armies {
		if fleet == nil || id == exclude || fleet.OwnerID != string(gs.PlayerFactionID) || !fleet.IsNaval || fleet.TransportCapacity(gs.UnitTypes) <= 0 {
			continue
		}
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	result := make([]*army.Army, 0, len(ids))
	for _, id := range ids {
		result = append(result, gs.Armies[id])
	}
	return result
}

func (r *Renderer) openNavalMissionPanel() {
	if r == nil || r.gs == nil || r.SelectedArmy == "" {
		return
	}
	fleet := r.gs.Armies[r.SelectedArmy]
	if !playerNavalMissionEligible(r.gs, fleet) {
		return
	}
	r.navalMissionArmy = fleet.ID
	r.navalMissionTargeting = false
	r.navalMissionKind = ""
	r.showNavalMissionPanel = true
	r.bringOverlayPanelToFront(overlayPanelNavalMission)
}

func (r *Renderer) closeNavalMissionPanel() {
	if r == nil {
		return
	}
	r.showNavalMissionPanel = false
	r.navalMissionArmy = ""
	r.navalMissionTargeting = false
	r.navalMissionKind = ""
}

func (r *Renderer) navalMissionPanelRowCount() int {
	if r == nil || r.gs == nil {
		return 1
	}
	return len(navalMissionOptions(r.gs, r.gs.Armies[r.navalMissionArmy]))
}

func (r *Renderer) navalMissionPanelRect() gameui.Rect {
	layout := navalMissionPanelLayoutFor(r.navalMissionPanelRowCount())
	return gameui.Rect{X: float64(layout.panelX), Y: float64(layout.panelY), W: float64(layout.panelW), H: float64(layout.panelH)}
}

func navalMissionPanelRowRect(layout navalMissionPanelLayout, visibleRow int) gameui.Rect {
	y := layout.rowY + float32(visibleRow)*navalMissionPanelRowH
	return gameui.Rect{
		X: float64(layout.rowX), Y: float64(y + 5),
		W: float64(layout.rowW), H: float64(navalMissionPanelRowH - 10),
	}
}

func navalMissionPanelRowButton(rect gameui.Rect) gameui.Button {
	return gameui.NewButton(rect.X, rect.Y, rect.W, rect.H, "")
}

// navalMissionPanelInteractiveHit, cursor ve input tarafında aynı görünür
// satır/kapatma düğmesi geometry'sini kullanır.
func (r *Renderer) navalMissionPanelInteractiveHit(fx, fy float64) bool {
	if r == nil || !r.showNavalMissionPanel || r.gs == nil {
		return false
	}
	layout := navalMissionPanelLayoutFor(r.navalMissionPanelRowCount())
	if layout.close.HitTest(fx, fy) {
		return true
	}
	fleet := r.gs.Armies[r.navalMissionArmy]
	if fleet != nil && fleet.NavalMission != nil && layout.clear.HitTest(fx, fy) {
		return true
	}
	if !r.navalMissionPanelRect().Hit(fx, fy) {
		return false
	}

	rowCount := r.navalMissionPanelRowCount()
	maxScroll := rowCount - navalMissionPanelVisibleRows
	if maxScroll < 0 {
		maxScroll = 0
	}
	start := r.navalMissionScroll
	if start < 0 {
		start = 0
	}
	if start > maxScroll {
		start = maxScroll
	}
	visibleRow := int((fy - float64(layout.rowY+5)) / float64(navalMissionPanelRowH))
	if visibleRow < 0 || visibleRow >= navalMissionPanelVisibleRows {
		return false
	}
	row := visibleRow + start
	return row >= 0 && row < rowCount && navalMissionPanelRowRect(layout, visibleRow).Hit(fx, fy)
}

func (r *Renderer) drawNavalMissionPanel(screen *ebiten.Image) {
	if r == nil || !r.showNavalMissionPanel || r.gs == nil {
		return
	}
	layout := navalMissionPanelLayoutFor(r.navalMissionPanelRowCount())
	drawUIOverlay(screen, color.RGBA{0, 0, 0, 205})
	drawUIPanelFrame(screen, gameui.Rect{
		X: float64(layout.panelX), Y: float64(layout.panelY),
		W: float64(layout.panelW), H: float64(layout.panelH),
	}, panelBg, panelBorder, 1.5, 3)
	drawUILabel(screen, gameui.Rect{X: float64(layout.panelX + 20), Y: float64(layout.panelY + 23), W: float64(layout.panelW - 80)}, "Donanma Görevi", ColorGold, gameui.TextLarge, gameui.TextAlignStart)
	drawUIButtonWidget(screen, layout.close, tinyButtonStyle)

	fleet := r.gs.Armies[r.navalMissionArmy]
	options := navalMissionOptions(r.gs, fleet)
	rowCount := len(options)
	maxScroll := rowCount - navalMissionPanelVisibleRows
	if maxScroll < 0 {
		maxScroll = 0
	}
	start := 0
	if r.navalMissionScroll > maxScroll {
		r.navalMissionScroll = maxScroll
	}
	start = r.navalMissionScroll
	end := start + navalMissionPanelVisibleRows
	if end > rowCount {
		end = rowCount
	}
	for row := start; row < end; row++ {
		rowRect := navalMissionPanelRowRect(layout, row-start)
		bg := color.RGBA{28, 22, 14, 220}
		if row%2 == 0 {
			bg = color.RGBA{38, 29, 16, 230}
		}
		rowButton := navalMissionPanelRowButton(rowRect)
		gameui.DrawButton(screen, rowButton, gameui.ButtonStyle{
			BG: bg, Border: color.RGBA{92, 70, 34, 180}, Text: ColorWhite, BorderWidth: 1,
		}, sharedTextRenderer{})
		option := options[row]
		active := fleet != nil && fleet.NavalMission != nil && fleet.NavalMission.Kind == option.kind && fleet.NavalMission.TargetFleetID == option.targetFleet
		textColor := ColorWhite
		if active {
			textColor = color.RGBA{165, 235, 175, 255}
		}
		labelW := rowRect.W - 28
		if active {
			labelW -= 70
		}
		label := trimTextToWidth(option.label, FaceMed, labelW)
		description := trimTextToWidth(option.description, FaceSmall, rowRect.W-28)
		effect := trimTextToWidth(option.effect, FaceSmall, rowRect.W-28)
		drawUILabel(screen, gameui.Rect{X: rowRect.X + 14, Y: rowRect.Y + 7, W: labelW}, label, textColor, gameui.TextMedium, gameui.TextAlignStart)
		drawUILabel(screen, gameui.Rect{X: rowRect.X + 14, Y: rowRect.Y + 29, W: rowRect.W - 28}, description, ColorGray, gameui.TextSmall, gameui.TextAlignStart)
		effectColor := ColorGold
		if active {
			effectColor = color.RGBA{180, 240, 190, 255}
		}
		drawUILabel(screen, gameui.Rect{X: rowRect.X + 14, Y: rowRect.Y + 50, W: rowRect.W - 28}, effect, effectColor, gameui.TextSmall, gameui.TextAlignStart)
		if active {
			drawUILabel(screen, gameui.Rect{X: rowRect.X + rowRect.W - 62, Y: rowRect.Y + 14, W: 48}, "AKTİF", color.RGBA{165, 235, 175, 255}, gameui.TextSmall, gameui.TextAlignEnd)
		}
	}
	footerY := layout.panelY + layout.panelH - navalMissionPanelFooterH
	if fleet != nil && fleet.NavalMission != nil {
		gameui.DrawButton(screen, layout.clear, commanderUnassignButtonStyle(), sharedTextRenderer{})
	}
	if rowCount > navalMissionPanelVisibleRows {
		drawUILabel(screen, gameui.Rect{X: float64(layout.panelX + 18), Y: float64(footerY + 15), W: 190}, "Görevler: "+itoa(start+1)+"-"+itoa(end)+"/"+itoa(rowCount), ColorGray, gameui.TextSmall, gameui.TextAlignStart)
	}
	drawUILabel(screen, gameui.Rect{X: float64(layout.panelX + layout.panelW - 120), Y: float64(footerY + 15), W: 102}, "ESC: kapat", ColorGray, gameui.TextSmall, gameui.TextAlignEnd)
}

func (r *Renderer) drawNavalMissionTargetingOverlay(screen *ebiten.Image) {
	if r == nil || !r.navalMissionTargeting || r.gs == nil {
		return
	}
	fleet := r.gs.Armies[r.navalMissionArmy]
	for _, region := range r.gs.Regions {
		if !navalMissionTargetCandidate(r.gs, r.navalMissionKind, fleet, region) {
			continue
		}
		sx, sy := r.regionScreenPos(region)
		candidateColor := color.RGBA{220, 80, 70, 110}
		borderColor := color.RGBA{255, 145, 120, 220}
		if r.navalMissionKind == army.NavalMissionTransport {
			candidateColor = color.RGBA{226, 163, 48, 115}
			borderColor = color.RGBA{255, 216, 110, 230}
		}
		vector.FillCircle(screen, float32(sx), float32(sy), 9, candidateColor, false)
		vector.StrokeCircle(screen, float32(sx), float32(sy), 9, 1.5, borderColor, false)
	}
	vector.FillRect(screen, 0, 0, float32(ScreenWidth), 42, color.RGBA{24, 18, 8, 235}, false)
	label := navalMissionLabelTR(r.navalMissionKind)
	DrawTextCentered(screen, stringsUpperTR(label)+" HEDEFİ: haritada uygun bölgeye tıkla • ESC: iptal", float64(ScreenWidth)/2, 14, FaceSmall, ColorGold)
}

func navalMissionTargetCandidate(gs *state.GameState, kind army.NavalMissionKind, fleet *army.Army, region *world.Region) bool {
	if gs == nil || region == nil || region.IsLocked {
		return false
	}
	switch kind {
	case army.NavalMissionPatrol:
		return fleet != nil && fleet.IsAtSea() && region.IsSea && region.ID == fleet.RegionID
	case army.NavalMissionBlockade:
		return fleet != nil && fleet.IsAtSea() && region.ID == fleet.RegionID && gs.IsValidNavalBlockadeTarget(fleet, region.ID)
	case army.NavalMissionTransport:
		return !region.IsSea && region.CanLandEnter() && region.IsCoastal(gs.Regions)
	default:
		return false
	}
}

func navalMissionTargetCircleHit(cx, cy, mx, my float64) bool {
	dx := mx - cx
	dy := my - cy
	return dx*dx+dy*dy <= 12*12
}

func (r *Renderer) navalMissionTargetHovering(fx, fy float64) bool {
	if r == nil || !r.navalMissionTargeting || r.gs == nil {
		return false
	}
	fleet := r.gs.Armies[r.navalMissionArmy]
	for _, region := range r.gs.Regions {
		if !navalMissionTargetCandidate(r.gs, r.navalMissionKind, fleet, region) {
			continue
		}
		sx, sy := r.regionScreenPos(region)
		if navalMissionTargetCircleHit(sx, sy, fx, fy) {
			return true
		}
	}
	return false
}

func stringsUpperTR(value string) string {
	if value == "escort" {
		return "ESCORT"
	}
	if value == "abluka" {
		return "ABLUKA"
	}
	if value == "devriye" {
		return "DEVRİYE"
	}
	return "NAKLİYE"
}

func (r *Renderer) handleNavalMissionPanelInput() InputAction {
	if r == nil || (!r.showNavalMissionPanel && !r.navalMissionTargeting) || r.gs == nil {
		return InputAction{}
	}
	if r.keyJustPressed(ebiten.KeyEscape) {
		r.closeNavalMissionPanel()
		return InputAction{}
	}
	if r.navalMissionTargeting {
		return r.handleNavalMissionTargetInput()
	}
	layout := navalMissionPanelLayoutFor(r.navalMissionPanelRowCount())
	_, wheelY := ebiten.Wheel()
	if wheelY != 0 {
		r.navalMissionScroll -= int(wheelY)
		maxScroll := r.navalMissionPanelRowCount() - navalMissionPanelVisibleRows
		if maxScroll < 0 {
			maxScroll = 0
		}
		if r.navalMissionScroll < 0 {
			r.navalMissionScroll = 0
		}
		if r.navalMissionScroll > maxScroll {
			r.navalMissionScroll = maxScroll
		}
	}
	if !r.mouseJustPressed(ebiten.MouseButtonLeft) {
		return InputAction{}
	}
	mx, my := ebiten.CursorPosition()
	fx, fy := float64(mx), float64(my)
	fleet := r.gs.Armies[r.navalMissionArmy]
	if layout.clear.HitTest(fx, fy) && fleet != nil && fleet.NavalMission != nil {
		aid := r.navalMissionArmy
		r.closeNavalMissionPanel()
		return InputAction{Kind: ActionClearNavalMission, ArmyID: aid}
	}
	if layout.close.HitTest(fx, fy) || !r.navalMissionPanelRect().Hit(fx, fy) {
		r.closeNavalMissionPanel()
		return InputAction{}
	}
	row := int((fy - float64(layout.rowY+5)) / float64(navalMissionPanelRowH))
	if row < 0 || row >= navalMissionPanelVisibleRows || !navalMissionPanelRowRect(layout, row).Hit(fx, fy) {
		return InputAction{}
	}
	row += r.navalMissionScroll
	options := navalMissionOptions(r.gs, r.gs.Armies[r.navalMissionArmy])
	if row < 0 || row >= len(options) {
		return InputAction{}
	}
	option := options[row]
	if option.kind == army.NavalMissionEscort {
		aid := r.navalMissionArmy
		r.closeNavalMissionPanel()
		return InputAction{Kind: ActionAssignNavalMission, ArmyID: aid, BuildingID: string(option.kind), TargetArmyID: option.targetFleet}
	}
	if option.kind == army.NavalMissionPatrol || option.kind == army.NavalMissionBlockade {
		aid := r.navalMissionArmy
		fleet := r.gs.Armies[aid]
		if fleet == nil {
			r.closeNavalMissionPanel()
			return InputAction{}
		}
		r.closeNavalMissionPanel()
		return InputAction{Kind: ActionAssignNavalMission, ArmyID: aid, BuildingID: string(option.kind), TargetRegion: fleet.RegionID}
	}
	r.navalMissionTargeting = true
	r.navalMissionKind = option.kind
	r.showNavalMissionPanel = false
	return InputAction{}
}

func (r *Renderer) handleNavalMissionTargetInput() InputAction {
	if r == nil || r.gs == nil || !r.navalMissionTargeting {
		return InputAction{}
	}
	if !r.mouseJustPressed(ebiten.MouseButtonLeft) {
		return InputAction{}
	}
	mx, my := ebiten.CursorPosition()
	if topStatusPanelHit(float64(mx), float64(my)) || topDateHudHit(float64(mx), float64(my)) || bottomActionHudHit(float64(mx), float64(my)) || minimapHit(float64(mx), float64(my)) {
		return InputAction{}
	}
	wx, wy := r.screenToWorld(float64(mx), float64(my))
	if r.worldMap == nil {
		return InputAction{}
	}
	regionID := r.worldMap.RegionAt(int(wx), int(wy))
	if regionID == "" {
		return InputAction{}
	}
	mission := army.NavalMission{Kind: r.navalMissionKind, TargetRegionID: regionID}
	if ok, reason := r.gs.CanAssignNavalMission(r.navalMissionArmy, mission); !ok {
		r.ShowCombatResult(reason)
		return InputAction{}
	}
	aid := r.navalMissionArmy
	r.closeNavalMissionPanel()
	return InputAction{Kind: ActionAssignNavalMission, ArmyID: aid, BuildingID: string(mission.Kind), TargetRegion: regionID}
}

func navalMissionLabelTR(kind army.NavalMissionKind) string {
	switch kind {
	case army.NavalMissionPatrol:
		return "Devriye"
	case army.NavalMissionBlockade:
		return "Abluka"
	case army.NavalMissionEscort:
		return "Escort"
	case army.NavalMissionTransport:
		return "Nakliye"
	default:
		return "Görev yok"
	}
}
