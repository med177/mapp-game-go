package render

import (
	"image"
	"image/color"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/state"
	gameui "mapp-game-go/internal/ui"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	characterRosterRowH       = 132.0
	characterRosterColumnGap  = 10.0
	characterRosterColumnHead = 122.0
	characterRosterPanelPad   = 10.0
)

var characterCouncilRoleIDs = [3]string{state.CouncilRoleHeadOfHouse, state.CouncilRoleTreasurer, state.CouncilRoleDiplomat}

func characterRosterHUDButtonRect() [4]float32 {
	return topHudUtilityButtonRect(1)
}

func buildCharacterRosterHUDButton() gameui.Button {
	r := characterRosterHUDButtonRect()
	button := buttonFromRectF32(r, "").WithIcon(gameui.IconPeople)
	button.IconSize = 21
	return button
}

func characterRosterHUDButtonHit(mx, my float64) bool {
	return buildCharacterRosterHUDButton().HitTest(mx, my)
}

func drawCharacterRosterHUDButton(screen *ebiten.Image, open bool) {
	button := buildCharacterRosterHUDButton()
	centerX := float32(button.X + button.W/2)
	centerY := float32(button.Y + button.H/2)
	fill := color.RGBA{38, 34, 28, 220}
	border := color.RGBA{110, 100, 82, 220}
	if open {
		fill = color.RGBA{144, 76, 36, 250}
		border = color.RGBA{255, 208, 104, 255}
	}
	vector.FillCircle(screen, centerX, centerY, float32(button.W/2), fill, true)
	vector.StrokeCircle(screen, centerX, centerY, float32(button.W/2-1), 1.5, border, true)
	gameui.DrawIcon(screen, button.Icon, button.X+(button.W-button.IconSize)/2, button.Y+(button.H-button.IconSize)/2, button.IconSize, ColorWhite)
}

func characterRosterPanelRect() gameui.Rect {
	return activeWarsPanelRect()
}

func characterRosterCloseButton() gameui.Button {
	panel := characterRosterPanelRect()
	return gameui.NewCloseButton(panel.X+panel.W-38, panel.Y+9, 26, 24)
}

func characterRosterColumns() (owned, candidates gameui.Rect) {
	panel := characterRosterPanelRect()
	content := gameui.Rect{
		X: panel.X + characterRosterPanelPad,
		Y: panel.Y + activeWarsPanelHeaderH,
		W: panel.W - characterRosterPanelPad*2,
		H: panel.H - activeWarsPanelHeaderH - characterRosterPanelPad,
	}
	ownedW := (content.W - characterRosterColumnGap) * 0.5
	if ownedW < 0 {
		ownedW = 0
	}
	owned = gameui.Rect{X: content.X, Y: content.Y, W: ownedW, H: content.H}
	candidates = gameui.Rect{
		X: owned.X + owned.W + characterRosterColumnGap,
		Y: content.Y,
		W: content.W - owned.W - characterRosterColumnGap,
		H: content.H,
	}
	return owned, candidates
}

func characterRosterListViewport(column gameui.Rect) gameui.Rect {
	return gameui.Rect{
		X: column.X + 5,
		Y: column.Y + characterRosterColumnHead,
		W: column.W - 10,
		H: column.H - characterRosterColumnHead - 5,
	}
}

func characterRosterOwnedListViewport(column gameui.Rect) gameui.Rect {
	const listTop = 34.0
	return gameui.Rect{
		X: column.X + 5,
		Y: column.Y + listTop,
		W: column.W - 10,
		H: column.H - listTop - 5,
	}
}

func characterRosterVisibleRows(viewport gameui.Rect) int {
	rows := int(viewport.H / characterRosterRowH)
	if rows < 1 {
		return 1
	}
	return rows
}

func characterRosterMaxScroll(count int, viewport gameui.Rect) int {
	maxScroll := count - characterRosterVisibleRows(viewport)
	if maxScroll < 0 {
		return 0
	}
	return maxScroll
}

func clampCharacterRosterScroll(scroll, count int, viewport gameui.Rect) int {
	if scroll < 0 {
		return 0
	}
	if maxScroll := characterRosterMaxScroll(count, viewport); scroll > maxScroll {
		return maxScroll
	}
	return scroll
}

func characterRosterRowRect(viewport gameui.Rect, visibleIndex int) gameui.Rect {
	return gameui.Rect{
		X: viewport.X,
		Y: viewport.Y + float64(visibleIndex)*characterRosterRowH,
		W: viewport.W,
		H: characterRosterRowH - 4,
	}
}

func characterRosterRowBorderRect(row gameui.Rect) gameui.Rect {
	const borderInset = 1.0
	return gameui.Rect{
		X: row.X + borderInset,
		Y: row.Y + borderInset,
		W: row.W - borderInset*2,
		H: row.H - borderInset*2,
	}
}

func characterRosterPortraitRect(row gameui.Rect) gameui.Rect {
	size := float64(armyPanelCommanderPortrait)
	return gameui.Rect{X: row.X + 6, Y: row.Y + 4, W: size, H: size}
}

func characterRosterTraitRowRect(row gameui.Rect) gameui.Rect {
	return gameui.Rect{X: row.X + 8, Y: row.Y + row.H - 24, W: row.W - 16, H: 18}
}

func characterRosterRowAt(x, y float64, count, scroll int, viewport gameui.Rect) int {
	if count == 0 || !viewport.Hit(x, y) {
		return -1
	}
	scroll = clampCharacterRosterScroll(scroll, count, viewport)
	visibleRows := characterRosterVisibleRows(viewport)
	for visibleIndex := 0; visibleIndex < visibleRows; visibleIndex++ {
		index := scroll + visibleIndex
		if index >= count {
			break
		}
		if characterRosterRowRect(viewport, visibleIndex).Hit(x, y) {
			return index
		}
	}
	return -1
}

func characterAdoptButton(column gameui.Rect, rowIndex, scroll int) gameui.Button {
	viewport := characterRosterListViewport(column)
	row := characterRosterRowRect(viewport, rowIndex-scroll)
	button := gameui.NewButton(row.X+row.W-34, row.Y+row.H-62, 28, 28, "").WithIcon(gameui.IconPlus)
	button.IconSize = 20
	return button
}

func characterReleaseButton(column gameui.Rect, rowIndex, scroll int) gameui.Button {
	viewport := characterRosterOwnedListViewport(column)
	row := characterRosterRowRect(viewport, rowIndex-scroll)
	button := gameui.NewButton(row.X+row.W-34, row.Y+59, 28, 28, "").WithIcon(gameui.IconClose)
	button.IconSize = 20
	return button
}

func (r *Renderer) confirmCharacterAdoption(character *army.Commander) {
	if r == nil || character == nil {
		return
	}
	r.ShowConfirmDialog(
		"Karakteri Haneye Kat",
		character.Name+" hanenize katılsın mı?",
		"Evet",
		"Hayır",
		InputAction{Kind: ActionAdoptNeutralCharacter, CharacterID: character.ID},
		nil,
	)
}

func (r *Renderer) confirmCharacterRelease(character *army.Commander) {
	if r == nil || character == nil {
		return
	}
	message := character.Name + " hanenizden çıkarılıp tarafsız adaylara geri bırakılsın mı? Ordu ve konsey atamaları kaldırılacak."
	r.ShowConfirmDialog(
		"Karakteri Serbest Bırak",
		message,
		"Evet",
		"Hayır",
		InputAction{Kind: ActionReleaseCharacter, CharacterID: character.ID},
		nil,
	)
}

func characterCouncilAssignButton(column gameui.Rect, rowIndex, scroll int) gameui.Button {
	viewport := characterRosterOwnedListViewport(column)
	row := characterRosterRowRect(viewport, rowIndex-scroll)
	return gameui.NewButton(row.X+row.W-78, row.Y+5, 74, 22, "Ata")
}

func characterCouncilRemoveDutyButton(column gameui.Rect, rowIndex, scroll int) gameui.Button {
	viewport := characterRosterOwnedListViewport(column)
	row := characterRosterRowRect(viewport, rowIndex-scroll)
	return gameui.NewButton(row.X+row.W-78, row.Y+5, 74, 22, "Kaldır")
}

func characterRosterDutyActionButton(column gameui.Rect, rowIndex, scroll int, character *army.Commander) gameui.Button {
	if character != nil && character.AssignedCouncilRole != "" {
		return characterCouncilRemoveDutyButton(column, rowIndex, scroll)
	}
	return characterCouncilAssignButton(column, rowIndex, scroll)
}

func characterRosterAssignmentMenuRect(column gameui.Rect, rowIndex, scroll int) gameui.Rect {
	viewport := characterRosterOwnedListViewport(column)
	row := characterRosterRowRect(viewport, rowIndex-scroll)
	const menuWidth = 116.0
	return gameui.Rect{X: row.X + row.W - menuWidth - 4, Y: row.Y + 31, W: menuWidth, H: 3 * 22}
}

func characterRosterAssignmentOptionButton(column gameui.Rect, rowIndex, scroll, optionIndex int) gameui.Button {
	menu := characterRosterAssignmentMenuRect(column, rowIndex, scroll)
	if optionIndex < 0 || optionIndex >= len(characterCouncilRoleIDs) {
		return gameui.Button{}
	}
	role := characterCouncilRoleIDs[optionIndex]
	return gameui.NewButton(menu.X+3, menu.Y+float64(optionIndex)*22+1, menu.W-6, 20, characterCouncilRoleNameTR(role))
}

func characterCouncilRoleNameTR(role string) string {
	switch role {
	case state.CouncilRoleHeadOfHouse:
		return "Hane Lideri"
	case state.CouncilRoleTreasurer:
		return "Hazine Sorumlusu"
	case state.CouncilRoleDiplomat:
		return "Diplomat"
	default:
		return "Görev"
	}
}

func characterCouncilRoleBenefitTR(role string) string {
	switch role {
	case state.CouncilRoleHeadOfHouse:
		return "Hane yönetimi"
	case state.CouncilRoleTreasurer:
		return "Vergi geliri +%5"
	case state.CouncilRoleDiplomat:
		return "Heyet ilişki bonusu +3"
	default:
		return ""
	}
}

func characterRosterAssignmentOptionAt(x, y float64, column gameui.Rect, rowIndex, scroll int) int {
	for i := range characterCouncilRoleIDs {
		if characterRosterAssignmentOptionButton(column, rowIndex, scroll, i).HitTest(x, y) {
			return i
		}
	}
	return -1
}

func characterRosterPanelInteractiveHit(r *Renderer, x, y float64) bool {
	if r == nil || r.gs == nil || !r.showCharacterRoster {
		return false
	}
	r.ensureCharacterRosterPanelCache()
	if characterRosterCloseButton().HitTest(x, y) {
		return true
	}
	ownedColumn, candidateColumn := characterRosterColumns()
	candidates := r.characterRosterCandidates
	viewport := characterRosterListViewport(candidateColumn)
	r.characterRosterNeutralScroll = clampCharacterRosterScroll(r.characterRosterNeutralScroll, len(candidates), viewport)
	if index := characterRosterRowAt(x, y, len(candidates), r.characterRosterNeutralScroll, viewport); index >= 0 {
		return characterAdoptButton(candidateColumn, index, r.characterRosterNeutralScroll).HitTest(x, y)
	}
	ownedRoster := r.characterRosterOwned
	ownedViewport := characterRosterOwnedListViewport(ownedColumn)
	r.characterRosterOwnedScroll = clampCharacterRosterScroll(r.characterRosterOwnedScroll, len(ownedRoster), ownedViewport)
	if menuIndex := characterRosterIndexByID(ownedRoster, r.characterRosterAssignMenuID); menuIndex >= r.characterRosterOwnedScroll &&
		menuIndex < r.characterRosterOwnedScroll+characterRosterVisibleRows(ownedViewport) {
		if option := characterRosterAssignmentOptionAt(x, y, ownedColumn, menuIndex, r.characterRosterOwnedScroll); option >= 0 {
			return true
		}
	}
	if index := characterRosterRowAt(x, y, len(ownedRoster), r.characterRosterOwnedScroll, ownedViewport); index >= 0 {
		if characterReleaseButton(ownedColumn, index, r.characterRosterOwnedScroll).HitTest(x, y) {
			return r.characterRosterAssignMenuID != ownedRoster[index].ID
		}
		character := ownedRoster[index]
		dutyButton := characterRosterDutyActionButton(ownedColumn, index, r.characterRosterOwnedScroll, character)
		if dutyButton.HitTest(x, y) &&
			(character.AssignedCouncilRole != "" ||
				r.gs.CanAssignCharacterToCouncilRole(r.gs.PlayerFactionID, character.ID, state.CouncilRoleHeadOfHouse)) {
			return true
		}
	}
	return false
}

func characterRosterIndexByID(characters []*army.Commander, characterID string) int {
	if characterID == "" {
		return -1
	}
	for index, character := range characters {
		if character != nil && character.ID == characterID {
			return index
		}
	}
	return -1
}

func handleCharacterRosterPanelClick(r *Renderer, x, y float64) (InputAction, bool) {
	if r == nil || r.gs == nil || !r.showCharacterRoster {
		return InputAction{}, false
	}
	if characterRosterCloseButton().HitTest(x, y) {
		r.CloseCharacterRosterPanel()
		return InputAction{}, true
	}
	ownedColumn, candidateColumn := characterRosterColumns()
	ownedRoster := r.characterRosterOwned
	candidates := r.characterRosterCandidates
	ownedViewport := characterRosterOwnedListViewport(ownedColumn)
	candidateViewport := characterRosterListViewport(candidateColumn)
	r.characterRosterOwnedScroll = clampCharacterRosterScroll(r.characterRosterOwnedScroll, len(ownedRoster), ownedViewport)
	r.characterRosterNeutralScroll = clampCharacterRosterScroll(r.characterRosterNeutralScroll, len(candidates), candidateViewport)

	menuIndex := characterRosterIndexByID(ownedRoster, r.characterRosterAssignMenuID)
	if menuIndex >= r.characterRosterOwnedScroll &&
		menuIndex < r.characterRosterOwnedScroll+characterRosterVisibleRows(ownedViewport) {
		if option := characterRosterAssignmentOptionAt(x, y, ownedColumn, menuIndex, r.characterRosterOwnedScroll); option >= 0 {
			character := ownedRoster[menuIndex]
			r.characterRosterAssignMenuID = ""
			if r.gs.CanAssignCharacterToCouncilRole(r.gs.PlayerFactionID, character.ID, characterCouncilRoleIDs[option]) {
				return InputAction{Kind: ActionAssignCouncilRole, CharacterID: character.ID, CouncilRole: characterCouncilRoleIDs[option]}, true
			}
			return InputAction{}, true
		}
		if characterCouncilAssignButton(ownedColumn, menuIndex, r.characterRosterOwnedScroll).HitTest(x, y) {
			r.characterRosterAssignMenuID = ""
			return InputAction{}, true
		}
		r.characterRosterAssignMenuID = ""
	}

	if index := characterRosterRowAt(x, y, len(ownedRoster), r.characterRosterOwnedScroll, ownedViewport); index >= 0 {
		character := ownedRoster[index]
		if characterReleaseButton(ownedColumn, index, r.characterRosterOwnedScroll).HitTest(x, y) {
			r.confirmCharacterRelease(character)
			return InputAction{}, true
		}
		if character.AssignedCouncilRole != "" &&
			characterRosterDutyActionButton(ownedColumn, index, r.characterRosterOwnedScroll, character).HitTest(x, y) {
			return InputAction{Kind: ActionClearCouncilRole, CharacterID: character.ID, CouncilRole: character.AssignedCouncilRole}, true
		}
		if character.AssignedCouncilRole == "" &&
			r.gs.CanAssignCharacterToCouncilRole(r.gs.PlayerFactionID, character.ID, state.CouncilRoleHeadOfHouse) &&
			characterCouncilAssignButton(ownedColumn, index, r.characterRosterOwnedScroll).HitTest(x, y) {
			r.characterRosterAssignMenuID = character.ID
			return InputAction{}, true
		}
	}
	if index := characterRosterRowAt(x, y, len(candidates), r.characterRosterNeutralScroll, candidateViewport); index >= 0 &&
		characterAdoptButton(candidateColumn, index, r.characterRosterNeutralScroll).HitTest(x, y) {
		r.confirmCharacterAdoption(candidates[index])
		return InputAction{}, true
	}
	return InputAction{}, false
}

func drawCharacterRosterScrollbar(screen *ebiten.Image, viewport gameui.Rect, count, scroll int) {
	maxScroll := characterRosterMaxScroll(count, viewport)
	if maxScroll <= 0 || viewport.H <= 0 {
		return
	}
	track := gameui.Rect{X: viewport.X + viewport.W - 4, Y: viewport.Y + 2, W: 3, H: viewport.H - 4}
	thumbH := track.H * float64(characterRosterVisibleRows(viewport)) / float64(count)
	if thumbH < 18 {
		thumbH = 18
	}
	if thumbH > track.H {
		thumbH = track.H
	}
	scroll = clampCharacterRosterScroll(scroll, count, viewport)
	thumbY := track.Y
	if track.H > thumbH {
		thumbY += (track.H - thumbH) * float64(scroll) / float64(maxScroll)
	}
	drawRoundedRect(screen, float32(track.X), float32(track.Y), float32(track.W), float32(track.H), 2, color.RGBA{70, 65, 55, 180})
	drawRoundedRect(screen, float32(track.X), float32(thumbY), float32(track.W), float32(thumbH), 2, color.RGBA{210, 175, 85, 230})
}

func drawCharacterRosterRows(screen *ebiten.Image, gs *state.GameState, column, viewport gameui.Rect, characters []*army.Commander, scroll int, candidate bool, assignmentMenuCharacterID string) {
	if len(characters) == 0 {
		message := "Karakter bulunmuyor."
		if candidate {
			message = "Tarafsız aday yok."
		}
		drawUILabel(screen, viewport, message, ColorGray, gameui.TextSmall, gameui.TextAlignCenter)
		return
	}

	scroll = clampCharacterRosterScroll(scroll, len(characters), viewport)
	visibleRows := characterRosterVisibleRows(viewport)
	end := scroll + visibleRows
	if end > len(characters) {
		end = len(characters)
	}
	left, top := int(viewport.X), int(viewport.Y)
	right, bottom := int(viewport.X+viewport.W), int(viewport.Y+viewport.H)
	if right <= left || bottom <= top {
		return
	}
	body := screen.SubImage(image.Rect(left, top, right, bottom)).(*ebiten.Image)
	for index := scroll; index < end; index++ {
		character := characters[index]
		row := characterRosterRowRect(viewport, index-scroll)
		rowBG := color.RGBA{35, 26, 14, 210}
		rowBorder := color.RGBA{100, 75, 30, 180}
		vector.FillRect(body, float32(row.X), float32(row.Y), float32(row.W), float32(row.H), rowBG, false)
		border := characterRosterRowBorderRect(row)
		vector.StrokeRect(body, float32(border.X), float32(border.Y), float32(border.W), float32(border.H), 1, rowBorder, false)
		portrait := characterRosterPortraitRect(row)
		drawCommanderPortrait(body, character, portrait.X, portrait.Y, portrait.W, portrait.H)

		textX := portrait.X + portrait.W + 10
		textW := row.X + row.W - textX - 8
		var adopt gameui.Button
		if candidate {
			adopt = characterAdoptButton(column, index, scroll)
			textW = adopt.X - textX - 5
		} else if character.AssignedCouncilRole != "" ||
			gs.CanAssignCharacterToCouncilRole(gs.PlayerFactionID, character.ID, state.CouncilRoleHeadOfHouse) {
			button := characterRosterDutyActionButton(column, index, scroll, character)
			if available := button.X - textX - 5; available < textW {
				textW = available
			}
		}
		drawUILabel(body, gameui.Rect{X: textX, Y: row.Y + 8, W: textW, H: 18}, trimTextToWidth(character.Name, FaceSmall, textW), ColorWhite, gameui.TextSmall, gameui.TextAlignStart)
		kind := "Karakter"
		if character.Type == army.CharacterTypeAdministrator {
			kind = "Yönetici"
		}
		if !candidate && character.AssignedArmyID != "" {
			kind = "Orduya atanmış"
		}
		detail := kind
		bonus := ""
		if candidate {
			detail = "Tarafsız aday"
		} else if character.AssignedCouncilRole != "" {
			detail = characterCouncilRoleNameTR(character.AssignedCouncilRole)
			bonus = characterCouncilRoleBenefitTR(character.AssignedCouncilRole)
		}
		drawUILabel(body, gameui.Rect{X: textX, Y: row.Y + 29, W: textW, H: 17}, trimTextToWidth(detail, FaceSmall, textW), ColorGray, gameui.TextSmall, gameui.TextAlignStart)
		if bonus != "" {
			drawUILabel(body, gameui.Rect{X: textX, Y: row.Y + 49, W: textW, H: 17}, trimTextToWidth(bonus, FaceSmall, textW), color.RGBA{165, 205, 145, 255}, gameui.TextSmall, gameui.TextAlignStart)
		}
		traitRow := characterRosterTraitRowRect(row)
		drawCharacterTraitBadges(body, character, traitRow.X, traitRow.Y, traitRow.W, characterTraitBadgeOptions{MaxRows: 1})
		if candidate {
			drawUIButtonWidget(body, adopt, applyTinyButtonStyle)
		} else {
			if character.ID != assignmentMenuCharacterID {
				drawUIButtonWidget(body, characterReleaseButton(column, index, scroll), dangerTinyButtonStyle)
			}
			if character.AssignedCouncilRole != "" {
				drawUIButtonWidget(body, characterRosterDutyActionButton(column, index, scroll, character), dangerTinyButtonStyle)
			} else if gs.CanAssignCharacterToCouncilRole(gs.PlayerFactionID, character.ID, state.CouncilRoleHeadOfHouse) {
				drawUIButtonWidget(body, characterRosterDutyActionButton(column, index, scroll, character), tinyButtonStyle)
			}
			if character.ID == assignmentMenuCharacterID {
				menu := characterRosterAssignmentMenuRect(column, index, scroll)
				drawUICardRect(body, menu, color.RGBA{24, 20, 14, 250}, color.RGBA{174, 132, 58, 255}, 1)
				for optionIndex := range characterCouncilRoleIDs {
					drawUIButtonWidget(body, characterRosterAssignmentOptionButton(column, index, scroll, optionIndex), tinyButtonStyle)
				}
			}
		}
	}
	drawCharacterRosterScrollbar(screen, viewport, len(characters), scroll)
}

func (r *Renderer) drawCharacterRosterPanel(screen *ebiten.Image) {
	if r == nil || r.gs == nil || !r.showCharacterRoster {
		return
	}
	r.ensureCharacterRosterPanelCache()
	panel := characterRosterPanelRect()
	drawUIPanelFrame(screen, panel, color.RGBA{12, 10, 8, 244}, color.RGBA{154, 112, 54, 255}, 1.5, 5)
	drawUILabel(screen, gameui.Rect{X: panel.X + characterRosterPanelPad, Y: panel.Y + 9, W: panel.W - 60}, "Karakterler", color.RGBA{255, 220, 118, 255}, gameui.TextMedium, gameui.TextAlignStart)
	drawCloseButton(screen, characterRosterCloseButton())

	ownedColumn, candidateColumn := characterRosterColumns()
	ownedRoster := r.characterRosterOwned
	candidates := r.characterRosterCandidates
	ownedViewport := characterRosterOwnedListViewport(ownedColumn)
	candidateViewport := characterRosterListViewport(candidateColumn)
	r.characterRosterOwnedScroll = clampCharacterRosterScroll(r.characterRosterOwnedScroll, len(ownedRoster), ownedViewport)
	r.characterRosterNeutralScroll = clampCharacterRosterScroll(r.characterRosterNeutralScroll, len(candidates), candidateViewport)

	drawUICardRect(screen, ownedColumn, color.RGBA{20, 16, 11, 220}, color.RGBA{92, 72, 40, 190}, 1)
	drawUICardRect(screen, candidateColumn, color.RGBA{20, 16, 11, 220}, color.RGBA{92, 72, 40, 190}, 1)
	drawUILabel(screen, gameui.Rect{X: ownedColumn.X + 8, Y: ownedColumn.Y + 7, W: ownedColumn.W - 16}, "Hane Üyeleri ("+itoa(len(ownedRoster))+")", ColorGold, gameui.TextSmall, gameui.TextAlignStart)
	drawUILabel(screen, gameui.Rect{X: candidateColumn.X + 8, Y: candidateColumn.Y + 7, W: candidateColumn.W - 16}, "Adaylar ("+itoa(len(candidates))+")", ColorGold, gameui.TextSmall, gameui.TextAlignStart)
	drawUILabel(screen, gameui.Rect{X: candidateColumn.X + 8, Y: candidateColumn.Y + 34, W: candidateColumn.W - 16, H: 32}, "Ücretsiz katılabilirler", ColorGray, gameui.TextSmall, gameui.TextAlignStart)
	drawCharacterRosterRows(screen, r.gs, ownedColumn, ownedViewport, ownedRoster, r.characterRosterOwnedScroll, false, r.characterRosterAssignMenuID)
	drawCharacterRosterRows(screen, r.gs, candidateColumn, candidateViewport, candidates, r.characterRosterNeutralScroll, true, "")
}

func (r *Renderer) handleCharacterRosterPanelInput() (InputAction, bool) {
	if r == nil || r.gs == nil || !r.showCharacterRoster {
		return InputAction{}, false
	}
	r.ensureCharacterRosterPanelCache()
	if r.keyJustPressed(ebiten.KeyEscape) {
		if r.characterRosterAssignMenuID != "" {
			r.characterRosterAssignMenuID = ""
			return InputAction{}, true
		}
		r.CloseCharacterRosterPanel()
		return InputAction{}, true
	}
	mx, my := ebiten.CursorPosition()
	x, y := float64(mx), float64(my)
	panel := characterRosterPanelRect()
	if !panel.Hit(x, y) {
		if r.mouseJustPressed(ebiten.MouseButtonLeft) {
			r.characterRosterAssignMenuID = ""
		}
		return InputAction{}, false
	}
	ownedColumn, candidateColumn := characterRosterColumns()
	ownedRoster := r.characterRosterOwned
	candidates := r.characterRosterCandidates
	ownedViewport := characterRosterOwnedListViewport(ownedColumn)
	candidateViewport := characterRosterListViewport(candidateColumn)
	if _, wheelY := ebiten.Wheel(); wheelY != 0 {
		r.characterRosterAssignMenuID = ""
		switch {
		case ownedViewport.Hit(x, y):
			r.characterRosterOwnedScroll = clampCharacterRosterScroll(r.characterRosterOwnedScroll-int(wheelY), len(ownedRoster), ownedViewport)
		case candidateViewport.Hit(x, y):
			r.characterRosterNeutralScroll = clampCharacterRosterScroll(r.characterRosterNeutralScroll-int(wheelY), len(candidates), candidateViewport)
		}
		return InputAction{}, true
	}
	if r.mouseJustPressed(ebiten.MouseButtonLeft) {
		if action, handled := handleCharacterRosterPanelClick(r, x, y); handled {
			return action, true
		}
	}
	leftPressed := ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
	leftWasPressed := r.prevMouse[ebiten.MouseButtonLeft]
	if leftPressed || leftWasPressed {
		r.prevMouse[ebiten.MouseButtonLeft] = leftPressed
		return InputAction{}, true
	}
	return InputAction{}, true
}

func (r *Renderer) OpenCharacterRosterPanel() {
	if r == nil || r.gs == nil || r.gs.PlayerFactionID == "" {
		return
	}
	r.showCharacterRoster = true
	r.characterRosterAssignMenuID = ""
	r.characterRosterOwnedScroll = 0
	r.characterRosterNeutralScroll = 0
	r.RefreshCharacterRosterPanel()
	r.bringOverlayPanelToFront(overlayPanelCharacterRoster)
}

func (r *Renderer) ensureCharacterRosterPanelCache() {
	if r != nil && !r.characterRosterCacheValid {
		r.RefreshCharacterRosterPanel()
	}
}

// RefreshCharacterRosterPanel rebuilds character lists after a state mutation.
func (r *Renderer) RefreshCharacterRosterPanel() {
	if r == nil || r.gs == nil {
		return
	}
	r.characterRosterOwned = r.gs.CharacterRoster(string(r.gs.PlayerFactionID))
	r.characterRosterCandidates = r.gs.NeutralCharacterCandidates()
	r.characterRosterCacheValid = true
}

func (r *Renderer) CloseCharacterRosterPanel() {
	if r == nil {
		return
	}
	r.showCharacterRoster = false
	r.characterRosterAssignMenuID = ""
	r.characterRosterOwnedScroll = 0
	r.characterRosterNeutralScroll = 0
	r.characterRosterOwned = nil
	r.characterRosterCandidates = nil
	r.characterRosterCacheValid = false
}

func (r *Renderer) characterRosterPanelHovering(x, y float64) bool {
	return characterRosterPanelInteractiveHit(r, x, y)
}
