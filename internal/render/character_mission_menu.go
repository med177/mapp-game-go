package render

import (
	"image/color"
	"math"

	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/scenario"
	gameui "mapp-game-go/internal/ui"
	"mapp-game-go/internal/world"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	characterMissionMenuW       = 244.0
	characterMissionMenuHeaderH = 38.0
	characterMissionMenuRowH    = 30.0
	characterMissionMenuPad     = 8.0
	characterMissionMenuMaxRows = 6
)

func characterMissionMenuRect(r *Renderer) (gameui.Rect, bool) {
	if r == nil || r.gs == nil || r.characterRosterMissionMenuCharacterID == "" {
		return gameui.Rect{}, false
	}
	owned := r.characterRosterOwned
	index := characterRosterIndexByID(owned, r.characterRosterMissionMenuCharacterID)
	if index < 0 {
		return gameui.Rect{}, false
	}
	column, _ := characterRosterColumns()
	viewport := characterRosterOwnedListViewport(column)
	row := characterRosterOwnedRowRect(viewport, index, r.characterRosterOwnedScroll, owned)
	itemCount := len(r.characterRosterMissionOptions)
	if r.characterRosterSelectedMissionID != "" {
		itemCount = len(r.characterRosterMissionTargets)
	}
	visibleRows := itemCount
	if visibleRows > characterMissionMenuMaxRows {
		visibleRows = characterMissionMenuMaxRows
	}
	if visibleRows < 1 {
		visibleRows = 1
	}
	height := characterMissionMenuHeaderH + float64(visibleRows)*characterMissionMenuRowH + characterMissionMenuPad
	panel := characterRosterPanelRect()
	width := math.Min(characterMissionMenuW, math.Max(120, panel.W-characterRosterPanelPad*2))
	x := row.X + row.W - width - 4
	if x < panel.X+characterRosterPanelPad {
		x = panel.X + characterRosterPanelPad
	}
	y := row.Y + 18
	if y+height > panel.Y+panel.H-characterRosterPanelPad {
		y = panel.Y + panel.H - height - characterRosterPanelPad
	}
	if y < panel.Y+characterRosterPanelPad {
		y = panel.Y + characterRosterPanelPad
	}
	return gameui.Rect{X: x, Y: y, W: width, H: height}, true
}

func characterMissionMenuCloseButton(menu gameui.Rect) gameui.Button {
	return gameui.NewCloseButton(menu.X+menu.W-27, menu.Y+characterMissionMenuPad, 22, 22)
}

func characterMissionMenuBackButton(menu gameui.Rect) gameui.Button {
	return gameui.NewButton(menu.X+8, menu.Y+characterMissionMenuPad, 22, 22, "").WithIcon(gameui.IconBack)
}

func characterMissionMenuTitleRect(menu gameui.Rect) gameui.Rect {
	return gameui.Rect{
		X: menu.X + 38,
		Y: menu.Y + (characterMissionMenuHeaderH-18)/2,
		W: menu.W - 73,
		H: 18,
	}
}

func characterMissionMenuOptionButton(menu gameui.Rect, index int, label string) gameui.Button {
	return gameui.NewButton(menu.X+8, menu.Y+characterMissionMenuHeaderH+float64(index)*characterMissionMenuRowH,
		menu.W-16, characterMissionMenuRowH-2, label)
}

func (r *Renderer) OpenCharacterMissionMenu(characterID string) bool {
	if r == nil || r.gs == nil || characterID == "" {
		return false
	}
	r.ensureCharacterRosterPanelCache()
	options := r.characterMissionsByCharacter[characterID]
	if len(options) == 0 {
		return false
	}
	r.characterRosterMissionMenuCharacterID = characterID
	r.characterRosterSelectedMissionID = ""
	r.characterRosterMissionOptions = append(r.characterRosterMissionOptions[:0], options...)
	r.characterRosterMissionTargets = nil
	r.characterRosterMissionScroll = 0
	r.characterRosterAssignMenuID = ""
	return true
}

func (r *Renderer) CloseCharacterMissionMenu() {
	if r == nil {
		return
	}
	r.characterRosterMissionMenuCharacterID = ""
	r.characterRosterSelectedMissionID = ""
	r.characterRosterMissionOptions = nil
	r.characterRosterMissionTargets = nil
	r.characterRosterMissionScroll = 0
}

func (r *Renderer) characterMissionOptionsFor(characterID string) []scenario.CharacterMissionDefinition {
	if r == nil {
		return nil
	}
	return r.characterMissionsByCharacter[characterID]
}

func (r *Renderer) selectedCharacterMission() (scenario.CharacterMissionDefinition, bool) {
	if r == nil || r.characterRosterSelectedMissionID == "" {
		return scenario.CharacterMissionDefinition{}, false
	}
	for _, mission := range r.characterRosterMissionOptions {
		if mission.ID == r.characterRosterSelectedMissionID {
			return mission, true
		}
	}
	return scenario.CharacterMissionDefinition{}, false
}

func (r *Renderer) characterMissionTargetLabel(targetID string) string {
	mission, ok := r.selectedCharacterMission()
	if !ok {
		return targetID
	}
	if mission.TargetKind == scenario.CharacterMissionTargetRegion {
		return regionDisplayName(r.gs, targetID)
	}
	return factionDisplayName(r.gs, targetID)
}

func (r *Renderer) confirmCharacterMissionStart(action InputAction, mission scenario.CharacterMissionDefinition, targetLabel string) {
	characterName := action.CharacterID
	if character := r.gs.Commanders[action.CharacterID]; character != nil && character.Name != "" {
		characterName = character.Name
	}
	message := characterName + " karakteri " + mission.NameTR + " görevine çıkacak"
	if targetLabel != "" {
		message += " (hedef: " + targetLabel + ")"
	}
	message += ". Süre: " + itoa(mission.DurationTurns) + " tur. Maliyet: " + itoa(mission.GoldCost) + " altın. Görev başlatılsın mı?"
	r.ShowConfirmDialog("Görevi Başlat", message, "Evet", "Hayır", action, nil)
}

func (r *Renderer) characterMissionMenuInteractiveHit(x, y float64) bool {
	menu, ok := characterMissionMenuRect(r)
	if !ok || !menu.Hit(x, y) {
		return false
	}
	if characterMissionMenuCloseButton(menu).HitTest(x, y) {
		return true
	}
	if r.characterRosterSelectedMissionID != "" && characterMissionMenuBackButton(menu).HitTest(x, y) {
		return true
	}
	if r.characterRosterSelectedMissionID == "" {
		owner := r.gs.Factions[r.gs.PlayerFactionID]
		for index, mission := range r.characterRosterMissionOptions {
			if index < r.characterRosterMissionScroll || index >= r.characterRosterMissionScroll+characterMissionMenuMaxRows {
				continue
			}
			button := characterMissionMenuOptionButton(menu, index-r.characterRosterMissionScroll, mission.NameTR)
			if button.HitTest(x, y) && owner != nil && owner.Gold >= mission.GoldCost {
				return true
			}
		}
		return false
	}
	for index, targetID := range r.characterRosterMissionTargets {
		if index < r.characterRosterMissionScroll || index >= r.characterRosterMissionScroll+characterMissionMenuMaxRows {
			continue
		}
		button := characterMissionMenuOptionButton(menu, index-r.characterRosterMissionScroll, r.characterMissionTargetLabel(targetID))
		if button.HitTest(x, y) {
			return true
		}
	}
	return false
}

func (r *Renderer) drawCharacterMissionMenu(screen *ebiten.Image) {
	menu, ok := characterMissionMenuRect(r)
	if !ok {
		return
	}
	drawUIPanelFrame(screen, menu, color.RGBA{18, 14, 10, 250}, color.RGBA{180, 140, 62, 255}, 1.2, 3)
	if r.characterRosterSelectedMissionID == "" {
		titleRect := characterMissionMenuTitleRect(menu)
		titleRect.X = menu.X + characterMissionMenuPad
		titleRect.W = menu.W - characterMissionMenuPad*2 - 30
		drawUILabel(screen, titleRect, "Görev seç", ColorGold, gameui.TextSmall, gameui.TextAlignStart)
	} else {
		drawUIButtonWidget(screen, characterMissionMenuBackButton(menu), tinyButtonStyle)
		mission, _ := r.selectedCharacterMission()
		titleRect := characterMissionMenuTitleRect(menu)
		drawUILabel(screen, titleRect, trimTextToWidth(mission.NameTR, FaceSmall, titleRect.W), ColorGold, gameui.TextSmall, gameui.TextAlignStart)
	}
	drawCloseButton(screen, characterMissionMenuCloseButton(menu))

	start := r.characterRosterMissionScroll
	var count int
	if r.characterRosterSelectedMissionID == "" {
		count = len(r.characterRosterMissionOptions)
	} else {
		count = len(r.characterRosterMissionTargets)
	}
	if start < 0 {
		start = 0
	}
	if start > count-characterMissionMenuMaxRows {
		start = count - characterMissionMenuMaxRows
	}
	if start < 0 {
		start = 0
	}
	end := start + characterMissionMenuMaxRows
	if end > count {
		end = count
	}
	for index := start; index < end; index++ {
		label := ""
		button := characterMissionMenuOptionButton(menu, index-start, label)
		button.Enabled = true
		if r.characterRosterSelectedMissionID == "" {
			mission := r.characterRosterMissionOptions[index]
			label = mission.NameTR + " · " + itoa(mission.DurationTurns) + " tur · " + itoa(mission.GoldCost) + " altın"
			owner := r.gs.Factions[r.gs.PlayerFactionID]
			button.Enabled = owner != nil && owner.Gold >= mission.GoldCost
		} else {
			label = r.characterMissionTargetLabel(r.characterRosterMissionTargets[index])
		}
		button.Label = trimTextToWidth(label, FaceSmall, button.W-12)
		drawUIButtonWidget(screen, button, tinyButtonStyle)
	}
	if count == 0 {
		message := "Uygun görev yok."
		if r.characterRosterSelectedMissionID != "" {
			message = "Uygun hedef yok."
		}
		drawUILabel(screen, gameui.Rect{X: menu.X + 8, Y: menu.Y + characterMissionMenuHeaderH + 8, W: menu.W - 16, H: 22}, message, ColorGray, gameui.TextSmall, gameui.TextAlignCenter)
	}
}

func (r *Renderer) handleCharacterMissionMenuClick(x, y float64) (InputAction, bool) {
	menu, ok := characterMissionMenuRect(r)
	if !ok || !menu.Hit(x, y) {
		return InputAction{}, false
	}
	if characterMissionMenuCloseButton(menu).HitTest(x, y) {
		r.CloseCharacterMissionMenu()
		return InputAction{}, true
	}
	if r.characterRosterSelectedMissionID != "" && characterMissionMenuBackButton(menu).HitTest(x, y) {
		r.characterRosterSelectedMissionID = ""
		r.characterRosterMissionTargets = nil
		r.characterRosterMissionScroll = 0
		return InputAction{}, true
	}
	if r.characterRosterSelectedMissionID == "" {
		owner := r.gs.Factions[r.gs.PlayerFactionID]
		for index, mission := range r.characterRosterMissionOptions {
			visibleIndex := index - r.characterRosterMissionScroll
			if visibleIndex < 0 || visibleIndex >= characterMissionMenuMaxRows {
				continue
			}
			button := characterMissionMenuOptionButton(menu, visibleIndex, mission.NameTR)
			if !button.HitTest(x, y) || owner == nil || owner.Gold < mission.GoldCost {
				continue
			}
			if mission.TargetKind == scenario.CharacterMissionTargetFaction || mission.TargetKind == scenario.CharacterMissionTargetRegion {
				r.characterRosterSelectedMissionID = mission.ID
				r.characterRosterMissionTargets = r.gs.CharacterMissionTargetIDs(r.gs.PlayerFactionID, mission.ID)
				r.characterRosterMissionScroll = 0
				return InputAction{}, true
			}
			characterID := r.characterRosterMissionMenuCharacterID
			action := InputAction{Kind: ActionStartCharacterMission, CharacterID: characterID, MissionID: mission.ID}
			r.CloseCharacterMissionMenu()
			r.confirmCharacterMissionStart(action, mission, "")
			return InputAction{}, true
		}
		return InputAction{}, true
	}
	for index, targetID := range r.characterRosterMissionTargets {
		visibleIndex := index - r.characterRosterMissionScroll
		if visibleIndex < 0 || visibleIndex >= characterMissionMenuMaxRows {
			continue
		}
		if characterMissionMenuOptionButton(menu, visibleIndex, r.characterMissionTargetLabel(targetID)).HitTest(x, y) {
			missionID := r.characterRosterSelectedMissionID
			characterID := r.characterRosterMissionMenuCharacterID
			mission, _ := r.selectedCharacterMission()
			targetLabel := r.characterMissionTargetLabel(targetID)
			r.CloseCharacterMissionMenu()
			action := InputAction{Kind: ActionStartCharacterMission, CharacterID: characterID, MissionID: missionID}
			if mission.TargetKind == scenario.CharacterMissionTargetRegion {
				action.TargetRegion = world.RegionID(targetID)
			} else {
				action.TargetFaction = faction.FactionID(targetID)
			}
			r.confirmCharacterMissionStart(action, mission, targetLabel)
			return InputAction{}, true
		}
	}
	return InputAction{}, true
}

func (r *Renderer) scrollCharacterMissionMenu(x, y, wheelY float64) bool {
	menu, ok := characterMissionMenuRect(r)
	if !ok || !menu.Hit(x, y) || wheelY == 0 {
		return false
	}
	count := len(r.characterRosterMissionOptions)
	if r.characterRosterSelectedMissionID != "" {
		count = len(r.characterRosterMissionTargets)
	}
	maxScroll := count - characterMissionMenuMaxRows
	if maxScroll < 0 {
		maxScroll = 0
	}
	r.characterRosterMissionScroll -= int(wheelY)
	if r.characterRosterMissionScroll < 0 {
		r.characterRosterMissionScroll = 0
	}
	if r.characterRosterMissionScroll > maxScroll {
		r.characterRosterMissionScroll = maxScroll
	}
	return true
}
