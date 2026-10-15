package render

import (
	"image"
	"image/color"
	"math"
	"strings"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/scenario"
	"mapp-game-go/internal/state"
	gameui "mapp-game-go/internal/ui"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	characterRosterRowH             = 136.0
	characterRosterCardH            = 128.0
	characterRosterColumnGap        = 10.0
	characterRosterCandidateListTop = 74.0
	characterRosterPanelPad         = 10.0
	characterRosterOwnedListTop     = 34.0
	characterRosterPanelMaxH        = 6*characterRosterRowH + activeWarsPanelHeaderH +
		characterRosterPanelPad + characterRosterOwnedListTop + 5
	characterRosterAssignMenuWidth = 156.0
	characterRosterAssignMenuPad   = 6.0
	characterRosterAssignOptionH   = 30.0
	characterRosterDutyButtonW     = 48.0
	characterRosterCardPadding     = 6.0
	characterRosterScrollbarGap    = 4.0
	characterRosterScrollbarW      = 3.0
	characterRosterScrollbarRight  = 5.0
	characterRosterDutyGroupPad    = 8.0
	characterRosterDutyGroupGap    = characterRosterDutyGroupPad*2 - (characterRosterRowH - characterRosterCardH)
	characterAbilityPopupWidth     = 320.0
	characterAbilityPopupPad       = 10.0
	characterAbilityPopupTitleH    = 22.0
	characterAbilityPopupLineH     = 18.0
	characterAbilityBadgeSize      = 24.0
	characterAbilityBadgeInset     = 2.0
)

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

func characterRosterHUDCounts(gs *state.GameState) (openRoles, activeMissions int) {
	if gs == nil || gs.PlayerFactionID == "" {
		return 0, 0
	}

	isRosterCharacter := func(character *army.Commander) bool {
		if character == nil || character.OwnerID != string(gs.PlayerFactionID) ||
			!character.ActiveInYear(gs.Year) || gs.DeadCommanderIDs[character.ID] ||
			gs.DismissedCommanderIDs[character.ID] {
			return false
		}
		switch character.Type {
		case army.CharacterTypeAdministrator, army.CharacterTypeCharacter, army.CharacterTypeCommander:
			return true
		default:
			return false
		}
	}

	roles := characterCouncilRoles(gs)
	for index, role := range roles {
		if role.ID == "" {
			continue
		}
		duplicate := false
		for _, previous := range roles[:index] {
			if previous.ID == role.ID {
				duplicate = true
				break
			}
		}
		if duplicate {
			continue
		}

		assigned := false
		for _, character := range gs.Commanders {
			if isRosterCharacter(character) && character.AssignedCouncilRole == role.ID {
				assigned = true
				break
			}
		}
		if !assigned {
			openRoles++
		}
	}

	for _, character := range gs.Commanders {
		if isRosterCharacter(character) && character.ActiveMissionID != "" {
			activeMissions++
		}
	}
	return openRoles, activeMissions
}

func drawCharacterRosterHUDCountBadge(screen *ebiten.Image, centerX, centerY float32, count int, fill color.RGBA) {
	const radius = 8.0
	vector.FillCircle(screen, centerX, centerY, radius, fill, true)
	drawUILabel(screen, gameui.Rect{
		X: float64(centerX - radius), Y: float64(centerY - 7),
		W: radius * 2, H: 14,
	}, itoa(count), ColorWhite, gameui.TextSmall, gameui.TextAlignCenter)
}

func drawCharacterRosterHUDButton(screen *ebiten.Image, gs *state.GameState, open bool) {
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

	openRoles, activeMissions := characterRosterHUDCounts(gs)
	if openRoles > 0 {
		drawCharacterRosterHUDCountBadge(screen, centerX+11, centerY-15, openRoles, color.RGBA{170, 42, 32, 255})
	}
	if activeMissions > 0 {
		drawCharacterRosterHUDCountBadge(screen, centerX+11, centerY+15, activeMissions, color.RGBA{218, 126, 35, 255})
	}
}

func characterRosterPanelRect() gameui.Rect {
	panel := activeWarsPanelRect()
	panel.H = characterRosterPanelMaxH
	widthIncrease := math.Min(40, panel.X-10)
	if widthIncrease > 0 {
		panel.X -= widthIncrease
		panel.W += widthIncrease
	}
	if remaining := ScreenHeight - panel.Y - 16; remaining < panel.H {
		panel.H = remaining
	}
	if panel.H < 150 {
		panel.H = 150
	}
	return panel
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
	widthIncrease := panel.W - activeWarsPanelRect().W
	ownedW := (content.W-characterRosterColumnGap)*0.5 + widthIncrease*0.5
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
		Y: column.Y + characterRosterCandidateListTop,
		W: column.W - 5 - characterRosterScrollbarGap - characterRosterScrollbarW - characterRosterScrollbarRight,
		H: column.H - characterRosterCandidateListTop - 5,
	}
}

func characterRosterOwnedListViewport(column gameui.Rect) gameui.Rect {
	return gameui.Rect{
		X: column.X + 5,
		Y: column.Y + characterRosterOwnedListTop,
		W: column.W - 5 - characterRosterScrollbarGap - characterRosterScrollbarW - characterRosterScrollbarRight,
		H: column.H - characterRosterOwnedListTop - 5,
	}
}

func characterRosterVisibleRows(viewport gameui.Rect) int {
	rows := int(math.Ceil(viewport.H / characterRosterRowH))
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
		H: characterRosterCardH,
	}
}

func characterRosterHasDutyGroupDivider(characters []*army.Commander, index int) bool {
	if index <= 0 || index >= len(characters) || characters[index-1] == nil || characters[index] == nil ||
		characters[index-1].AssignedCouncilRole == "" || characters[index].AssignedCouncilRole != "" {
		return false
	}
	return true
}

func characterRosterOwnedRowRect(viewport gameui.Rect, index, scroll int, characters []*army.Commander) gameui.Rect {
	row := characterRosterRowRect(viewport, index-scroll)
	for boundary := scroll + 1; boundary <= index; boundary++ {
		if characterRosterHasDutyGroupDivider(characters, boundary) {
			row.Y += characterRosterDutyGroupGap
		}
	}
	return row
}

func characterRosterDutyGroupDividerY(viewport gameui.Rect, characters []*army.Commander, index, scroll int) (float64, bool) {
	if !characterRosterHasDutyGroupDivider(characters, index) {
		return 0, false
	}
	row := characterRosterOwnedRowRect(viewport, index, scroll, characters)
	return row.Y - (characterRosterRowH-row.H+characterRosterDutyGroupGap)/2, true
}

func characterRosterOwnedVisibleRows(viewport gameui.Rect, characters []*army.Commander, scroll int) int {
	if len(characters) == 0 {
		return 0
	}
	visibleRows := 0
	for index := scroll; index < len(characters); index++ {
		row := characterRosterOwnedRowRect(viewport, index, scroll, characters)
		if visibleRows > 0 && row.Y >= viewport.Y+viewport.H {
			break
		}
		visibleRows++
	}
	if visibleRows == 0 {
		return 1
	}
	return visibleRows
}

func characterRosterMaxOwnedScroll(viewport gameui.Rect, characters []*army.Commander) int {
	for scroll := 0; scroll < len(characters); scroll++ {
		lastRow := characterRosterOwnedRowRect(viewport, len(characters)-1, scroll, characters)
		if lastRow.Y+lastRow.H <= viewport.Y+viewport.H {
			return scroll
		}
	}
	if len(characters) > 0 {
		return len(characters) - 1
	}
	return 0
}

func clampCharacterRosterOwnedScroll(scroll int, viewport gameui.Rect, characters []*army.Commander) int {
	if scroll < 0 {
		return 0
	}
	if maxScroll := characterRosterMaxOwnedScroll(viewport, characters); scroll > maxScroll {
		return maxScroll
	}
	return scroll
}

func characterRosterOwnedRowAt(x, y float64, characters []*army.Commander, scroll int, viewport gameui.Rect) int {
	if len(characters) == 0 || !viewport.Hit(x, y) {
		return -1
	}
	scroll = clampCharacterRosterOwnedScroll(scroll, viewport, characters)
	end := scroll + characterRosterOwnedVisibleRows(viewport, characters, scroll)
	if end > len(characters) {
		end = len(characters)
	}
	for index := scroll; index < end; index++ {
		if characterRosterOwnedRowRect(viewport, index, scroll, characters).Hit(x, y) {
			return index
		}
	}
	return -1
}

func characterRosterDetailRect(row gameui.Rect, textX float64) gameui.Rect {
	return gameui.Rect{X: textX, Y: row.Y + 34, W: row.X + row.W - textX - characterRosterCardPadding, H: 17}
}

func characterRosterCandidateCostLines(character *army.Commander) [2]string {
	if character == nil {
		return [2]string{}
	}
	return [2]string{
		"Alım: " + itoa(character.AdoptionCostGold) + " altın (tek seferlik)",
		"Bakım: " + itoa(character.GoldUpkeep) + " altın/tur",
	}
}

func characterRosterMissionDetailLines(missionName string, turnsRemaining int) [2]string {
	return [2]string{
		"Görevde: " + missionName,
		"Kalan süre: " + itoa(turnsRemaining) + " tur",
	}
}

func characterRosterUpkeepLabel(character *army.Commander, role scenario.CouncilRoleDefinition) string {
	if role.NoCost {
		return "Tur başı bakım yok"
	}
	if character == nil {
		return "Bakım: 0 altın/tur"
	}
	return "Bakım: " + itoa(character.GoldUpkeep) + " altın/tur"
}

func characterRosterCouncilManageable(character *army.Commander) bool {
	return character != nil &&
		(character.Type == army.CharacterTypeAdministrator || character.Type == army.CharacterTypeCharacter ||
			character.Type == army.CharacterTypeCommander)
}

func characterRosterReleaseable(character *army.Commander) bool {
	return character != nil && character.ActiveMissionID == "" &&
		(character.Type == army.CharacterTypeAdministrator || character.Type == army.CharacterTypeCharacter)
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
	return gameui.Rect{X: row.X + characterRosterCardPadding, Y: row.Y + characterRosterCardPadding, W: size, H: size}
}

func characterRosterAbilityBadgeRect(portrait gameui.Rect) gameui.Rect {
	size := characterAbilityBadgeSize
	return gameui.Rect{
		X: portrait.X + portrait.W - size + characterAbilityBadgeInset,
		Y: portrait.Y + portrait.H - size + characterAbilityBadgeInset,
		W: size,
		H: size,
	}
}

func drawCharacterRosterAbilityBadge(screen *ebiten.Image, portrait gameui.Rect) {
	badge := characterRosterAbilityBadgeRect(portrait)
	vector.FillRect(screen, float32(badge.X), float32(badge.Y), float32(badge.W), float32(badge.H), color.RGBA{23, 31, 39, 245}, false)
	vector.StrokeRect(screen, float32(badge.X), float32(badge.Y), float32(badge.W), float32(badge.H), 1, color.RGBA{180, 205, 220, 255}, false)
	gameui.DrawIcon(
		screen, gameui.IconAbility,
		badge.X+characterAbilityBadgeInset, badge.Y+characterAbilityBadgeInset,
		badge.W-characterAbilityBadgeInset*2, color.White,
	)
}

func characterRosterTraitRowRect(row gameui.Rect) gameui.Rect {
	return gameui.Rect{
		X: row.X + characterRosterCardPadding,
		Y: row.Y + row.H - 24,
		W: row.W - characterRosterCardPadding*2,
		H: 18,
	}
}

func characterRosterLineRectBeforeButton(row gameui.Rect, textX, y, height float64, button gameui.Button) gameui.Rect {
	line := gameui.Rect{
		X: textX, Y: y, W: row.X + row.W - textX - characterRosterCardPadding, H: height,
	}
	if line.Y < button.Y+button.H && line.Y+line.H > button.Y &&
		button.X > line.X && button.X < line.X+line.W {
		line.W = button.X - line.X - 4
	}
	return line
}

func characterAbilityPopupRect(mx, my float64, abilityCount int) gameui.Rect {
	height := characterAbilityPopupPad*2 + characterAbilityPopupTitleH +
		float64(abilityCount)*characterAbilityPopupLineH
	x, y, width, height := tooltipRect(mx, my, characterAbilityPopupWidth, height)
	return gameui.Rect{X: x, Y: y, W: width, H: height}
}

func characterRosterScrollbarTrack(viewport gameui.Rect) gameui.Rect {
	return gameui.Rect{
		X: viewport.X + viewport.W + characterRosterScrollbarGap,
		Y: viewport.Y + 2,
		W: characterRosterScrollbarW,
		H: viewport.H - 4,
	}
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

func characterRosterOptionalCharacters(rosters [][]*army.Commander) []*army.Commander {
	if len(rosters) == 0 {
		return nil
	}
	return rosters[0]
}

func characterReleaseButton(column gameui.Rect, rowIndex, scroll int, rosters ...[]*army.Commander) gameui.Button {
	viewport := characterRosterOwnedListViewport(column)
	row := characterRosterOwnedRowRect(viewport, rowIndex, scroll, characterRosterOptionalCharacters(rosters))
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
		character.Name+" hanenize katılsın mı? Tek seferlik ücret: "+itoa(character.AdoptionCostGold)+
			" altın; bakım: "+itoa(character.GoldUpkeep)+" altın/tur.",
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

func characterCouncilAssignButton(column gameui.Rect, rowIndex, scroll int, rosters ...[]*army.Commander) gameui.Button {
	viewport := characterRosterOwnedListViewport(column)
	row := characterRosterOwnedRowRect(viewport, rowIndex, scroll, characterRosterOptionalCharacters(rosters))
	return gameui.NewButton(row.X+row.W-characterRosterCardPadding-characterRosterDutyButtonW, row.Y+5, characterRosterDutyButtonW, 22, "Ata")
}

func characterCouncilRemoveDutyButton(column gameui.Rect, rowIndex, scroll int, rosters ...[]*army.Commander) gameui.Button {
	viewport := characterRosterOwnedListViewport(column)
	row := characterRosterOwnedRowRect(viewport, rowIndex, scroll, characterRosterOptionalCharacters(rosters))
	return gameui.NewButton(row.X+row.W-characterRosterCardPadding-characterRosterDutyButtonW, row.Y+5, characterRosterDutyButtonW, 22, "Kaldır")
}

func characterRosterDutyActionButton(column gameui.Rect, rowIndex, scroll int, character *army.Commander, rosters ...[]*army.Commander) gameui.Button {
	if character != nil && character.AssignedCouncilRole != "" {
		return characterCouncilRemoveDutyButton(column, rowIndex, scroll, rosters...)
	}
	return characterCouncilAssignButton(column, rowIndex, scroll, rosters...)
}

func characterRosterDutyRemovable(character *army.Commander) bool {
	return character != nil && character.AssignedCouncilRole != "" && character.ActiveMissionID == ""
}

func characterRosterMissionButton(column gameui.Rect, rowIndex, scroll int, rosters ...[]*army.Commander) gameui.Button {
	viewport := characterRosterOwnedListViewport(column)
	characters := characterRosterOptionalCharacters(rosters)
	row := characterRosterOwnedRowRect(viewport, rowIndex, scroll, characters)
	label := "Görev"
	if rowIndex >= 0 && rowIndex < len(characters) && characters[rowIndex] != nil &&
		len(characters[rowIndex].PendingAdventureAbilities) > 0 {
		label = "Yetenek"
	}
	return gameui.NewButton(
		row.X+row.W-characterRosterCardPadding-characterRosterDutyButtonW,
		row.Y+row.H-characterRosterCardPadding-18,
		characterRosterDutyButtonW,
		18,
		label,
	)
}

func characterRosterAssignmentMenuRectWithCount(column gameui.Rect, rowIndex, scroll, roleCount int, rosters ...[]*army.Commander) gameui.Rect {
	viewport := characterRosterOwnedListViewport(column)
	row := characterRosterOwnedRowRect(viewport, rowIndex, scroll, characterRosterOptionalCharacters(rosters))
	return gameui.Rect{
		X: row.X + row.W - characterRosterAssignMenuWidth - 4,
		Y: row.Y + 31,
		W: characterRosterAssignMenuWidth,
		H: characterRosterAssignMenuPad*2 + float64(roleCount)*characterRosterAssignOptionH,
	}
}

func characterRosterAssignmentOptionButtonForRole(column gameui.Rect, rowIndex, scroll, optionIndex, roleCount int, roleName string, rosters ...[]*army.Commander) gameui.Button {
	menu := characterRosterAssignmentMenuRectWithCount(column, rowIndex, scroll, roleCount, rosters...)
	if optionIndex < 0 || optionIndex >= roleCount {
		return gameui.Button{}
	}
	return gameui.NewButton(
		menu.X+characterRosterAssignMenuPad,
		menu.Y+characterRosterAssignMenuPad+float64(optionIndex)*characterRosterAssignOptionH,
		menu.W-characterRosterAssignMenuPad*2,
		characterRosterAssignOptionH,
		roleName,
	)
}

func characterCouncilRoles(gs *state.GameState) []scenario.CouncilRoleDefinition {
	if gs == nil || gs.CouncilConfig.Roles == nil {
		return scenario.DefaultCharacterCouncilConfig().Roles
	}
	return gs.CouncilConfig.Roles
}

func characterCanBeAssignedToAnyCouncilRole(gs *state.GameState, characterID string) bool {
	if gs == nil {
		return false
	}
	for _, role := range characterCouncilRoles(gs) {
		if gs.CanAssignCharacterToCouncilRole(gs.PlayerFactionID, characterID, role.ID) {
			return true
		}
	}
	return false
}

func characterCouncilEffectsLinesTR(effects scenario.CouncilEffects) []string {
	var parts []string
	if percent := councilEffectPercent(effects.IncomeMod); percent != 0 {
		parts = append(parts, "Bölgesel vergi "+formatSignedCouncilPercent(percent))
	}
	if effects.SatisfactionPoints != 0 {
		parts = append(parts, "Memnuniyet "+formatSignedCouncilPoints(effects.SatisfactionPoints))
	}
	if percent := councilEffectPercent(effects.MilitaryPowerMod); percent != 0 {
		parts = append(parts, "Askerî güç "+formatSignedCouncilPercent(percent))
	}
	if percent := councilEffectPercent(effects.CommanderPowerMod); percent != 0 {
		parts = append(parts, "Komuta gücü "+formatSignedCouncilPercent(percent))
	}
	if effects.RelationBonusPoints != 0 {
		parts = append(parts, "Heyet "+formatSignedCouncilPoints(effects.RelationBonusPoints))
	}
	if percent := councilEffectPercent(effects.ResourceProductionMod); percent != 0 {
		parts = append(parts, "Kaynak üretimi "+formatSignedCouncilPercent(percent))
	}
	if percent := councilEffectPercent(effects.ResearchSpeedMod); percent != 0 {
		parts = append(parts, "Araştırma hızı "+formatSignedCouncilPercent(percent))
	}
	return parts
}

func characterAdventureAbilityLinesTR(abilities []army.CharacterAbility) []string {
	var parts []string
	for _, ability := range abilities {
		if ability.NameTR == "" {
			continue
		}
		label := ability.NameTR
		effectName := ""
		switch {
		case ability.Scope == "faction" && ability.Effect == "military_power":
			effectName = "askerî güç"
		case ability.Scope == "army" && ability.Effect == "army_power":
			effectName = "ordu gücü"
		case ability.Scope == "navy" && ability.Effect == "navy_power":
			effectName = "donanma gücü"
		}
		if effectName != "" && ability.Modifier != 0 {
			percent := int(math.Round(ability.Modifier * 100))
			label += " (" + formatSignedCouncilPercent(percent) + " " + effectName + ")"
		}
		parts = append(parts, label)
	}
	return parts
}

func characterAdventureAbilityEffectsSummaryTR(abilities []army.CharacterAbility) string {
	var effects []string
	for _, ability := range abilities {
		effectName := ""
		switch {
		case ability.Scope == "faction" && ability.Effect == "military_power":
			effectName = "askerî güç"
		case ability.Scope == "army" && ability.Effect == "army_power":
			effectName = "ordu gücü"
		case ability.Scope == "navy" && ability.Effect == "navy_power":
			effectName = "donanma gücü"
		}
		if effectName == "" || ability.Modifier == 0 {
			continue
		}
		percent := int(math.Round(ability.Modifier * 100))
		effects = append(effects, formatSignedCouncilPercent(percent)+" "+effectName)
	}
	return strings.Join(effects, " · ")
}

func characterRosterAbilityPortraitAt(r *Renderer, x, y float64) (*army.Commander, []string) {
	if r == nil || r.gs == nil || !r.showCharacterRoster {
		return nil, nil
	}
	r.ensureCharacterRosterPanelCache()
	ownedColumn, candidateColumn := characterRosterColumns()
	ownedViewport := characterRosterOwnedListViewport(ownedColumn)
	r.characterRosterOwnedScroll = clampCharacterRosterOwnedScroll(
		r.characterRosterOwnedScroll, ownedViewport, r.characterRosterOwned,
	)
	ownedVisible := characterRosterOwnedVisibleRows(
		ownedViewport, r.characterRosterOwned, r.characterRosterOwnedScroll,
	)
	for index := r.characterRosterOwnedScroll; index < len(r.characterRosterOwned) &&
		index < r.characterRosterOwnedScroll+ownedVisible; index++ {
		character := r.characterRosterOwned[index]
		if character == nil {
			continue
		}
		portrait := characterRosterPortraitRect(characterRosterOwnedRowRect(
			ownedViewport, index, r.characterRosterOwnedScroll, r.characterRosterOwned,
		))
		if ownedViewport.Hit(x, y) && portrait.Hit(x, y) {
			if lines := characterAdventureAbilityLinesTR(character.AdventureAbilities); len(lines) > 0 {
				return character, lines
			}
		}
	}

	candidateViewport := characterRosterListViewport(candidateColumn)
	r.characterRosterNeutralScroll = clampCharacterRosterScroll(
		r.characterRosterNeutralScroll, len(r.characterRosterCandidates), candidateViewport,
	)
	candidateVisible := characterRosterVisibleRows(candidateViewport)
	for visibleIndex := 0; visibleIndex < candidateVisible; visibleIndex++ {
		index := r.characterRosterNeutralScroll + visibleIndex
		if index >= len(r.characterRosterCandidates) {
			break
		}
		character := r.characterRosterCandidates[index]
		if character == nil {
			continue
		}
		portrait := characterRosterPortraitRect(characterRosterRowRect(candidateViewport, visibleIndex))
		if candidateViewport.Hit(x, y) && portrait.Hit(x, y) {
			if lines := characterAdventureAbilityLinesTR(character.AdventureAbilities); len(lines) > 0 {
				return character, lines
			}
		}
	}
	return nil, nil
}

func drawCharacterRosterAbilityPopup(screen *ebiten.Image, r *Renderer) {
	if screen == nil || r == nil {
		return
	}
	mx, my := ebiten.CursorPosition()
	character, lines := characterRosterAbilityPortraitAt(r, float64(mx), float64(my))
	if character == nil || len(lines) == 0 {
		return
	}
	popup := characterAbilityPopupRect(float64(mx), float64(my), len(lines))
	drawTooltipBox(screen, popup.X, popup.Y, popup.W, popup.H)
	drawUILabel(screen, gameui.Rect{
		X: popup.X + characterAbilityPopupPad,
		Y: popup.Y + characterAbilityPopupPad,
		W: popup.W - characterAbilityPopupPad*2,
		H: characterAbilityPopupTitleH,
	}, "Yetenekleri", ColorGold, gameui.TextSmall, gameui.TextAlignStart)
	for index, line := range lines {
		lineRect := gameui.Rect{
			X: popup.X + characterAbilityPopupPad,
			Y: popup.Y + characterAbilityPopupPad + characterAbilityPopupTitleH +
				float64(index)*characterAbilityPopupLineH,
			W: popup.W - characterAbilityPopupPad*2,
			H: characterAbilityPopupLineH,
		}
		drawUILabel(screen, lineRect, trimTextToWidth(line, FaceSmall, lineRect.W),
			ColorWhite, gameui.TextSmall, gameui.TextAlignStart)
	}
}

func formatSignedCouncilPercent(value int) string {
	sign := "+"
	if value < 0 {
		sign = "-"
		value = -value
	}
	return sign + "%" + itoa(value)
}

func formatSignedCouncilPoints(value int) string {
	sign := "+"
	if value < 0 {
		sign = "-"
		value = -value
	}
	return sign + itoa(value)
}

func councilEffectPercent(mod float64) int {
	if mod == 0 {
		return 0
	}
	return int(math.Round((mod - 1) * 100))
}

func characterRosterAssignmentOptionAtForRoles(x, y float64, column gameui.Rect, rowIndex, scroll int, roles []scenario.CouncilRoleDefinition, rosters ...[]*army.Commander) int {
	for i, role := range roles {
		button := characterRosterAssignmentOptionButtonForRole(column, rowIndex, scroll, i, len(roles), role.NameTR, rosters...)
		if button.HitTest(x, y) {
			return i
		}
	}
	return -1
}

func characterRosterCanAssignRole(gs *state.GameState, character *army.Commander, role scenario.CouncilRoleDefinition) bool {
	return gs != nil && character != nil &&
		gs.CanAssignCharacterToCouncilRole(gs.PlayerFactionID, character.ID, role.ID)
}

func characterRosterPanelInteractiveHit(r *Renderer, x, y float64) bool {
	if r == nil || r.gs == nil || !r.showCharacterRoster {
		return false
	}
	r.ensureCharacterRosterPanelCache()
	if characterRosterCloseButton().HitTest(x, y) {
		return true
	}
	if character, _ := characterRosterAbilityPortraitAt(r, x, y); character != nil {
		return true
	}
	if r.characterMissionMenuInteractiveHit(x, y) {
		return true
	}
	ownedColumn, candidateColumn := characterRosterColumns()
	candidates := r.characterRosterCandidates
	roles := characterCouncilRoles(r.gs)
	viewport := characterRosterListViewport(candidateColumn)
	r.characterRosterNeutralScroll = clampCharacterRosterScroll(r.characterRosterNeutralScroll, len(candidates), viewport)
	if index := characterRosterRowAt(x, y, len(candidates), r.characterRosterNeutralScroll, viewport); index >= 0 {
		return characterAdoptButton(candidateColumn, index, r.characterRosterNeutralScroll).HitTest(x, y)
	}
	ownedRoster := r.characterRosterOwned
	ownedViewport := characterRosterOwnedListViewport(ownedColumn)
	r.characterRosterOwnedScroll = clampCharacterRosterOwnedScroll(r.characterRosterOwnedScroll, ownedViewport, ownedRoster)
	if menuIndex := characterRosterIndexByID(ownedRoster, r.characterRosterAssignMenuID); menuIndex >= r.characterRosterOwnedScroll &&
		menuIndex < r.characterRosterOwnedScroll+characterRosterOwnedVisibleRows(ownedViewport, ownedRoster, r.characterRosterOwnedScroll) {
		if characterRosterCouncilManageable(ownedRoster[menuIndex]) {
			if option := characterRosterAssignmentOptionAtForRoles(x, y, ownedColumn, menuIndex, r.characterRosterOwnedScroll, roles, ownedRoster); option >= 0 {
				return characterRosterCanAssignRole(r.gs, ownedRoster[menuIndex], roles[option])
			}
		}
	}
	if index := characterRosterOwnedRowAt(x, y, ownedRoster, r.characterRosterOwnedScroll, ownedViewport); index >= 0 {
		character := ownedRoster[index]
		if characterRosterReleaseable(character) && character.AssignedCouncilRole == "" &&
			characterReleaseButton(ownedColumn, index, r.characterRosterOwnedScroll, ownedRoster).HitTest(x, y) {
			return r.characterRosterAssignMenuID != character.ID
		}
		dutyButton := characterRosterDutyActionButton(ownedColumn, index, r.characterRosterOwnedScroll, character, ownedRoster)
		if characterRosterCouncilManageable(character) && dutyButton.HitTest(x, y) &&
			(characterRosterDutyRemovable(character) ||
				characterCanBeAssignedToAnyCouncilRole(r.gs, character.ID)) {
			return true
		}
		if character.ActiveMissionID == "" && (len(character.PendingAdventureAbilities) > 0 || len(r.characterMissionOptionsFor(character.ID)) > 0) &&
			characterRosterMissionButton(ownedColumn, index, r.characterRosterOwnedScroll, ownedRoster).HitTest(x, y) {
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
	if r.characterRosterMissionMenuCharacterID != "" {
		if action, handled := r.handleCharacterMissionMenuClick(x, y); handled {
			return action, true
		}
		r.CloseCharacterMissionMenu()
	}
	ownedColumn, candidateColumn := characterRosterColumns()
	ownedRoster := r.characterRosterOwned
	candidates := r.characterRosterCandidates
	roles := characterCouncilRoles(r.gs)
	ownedViewport := characterRosterOwnedListViewport(ownedColumn)
	candidateViewport := characterRosterListViewport(candidateColumn)
	r.characterRosterOwnedScroll = clampCharacterRosterOwnedScroll(r.characterRosterOwnedScroll, ownedViewport, ownedRoster)
	r.characterRosterNeutralScroll = clampCharacterRosterScroll(r.characterRosterNeutralScroll, len(candidates), candidateViewport)

	menuIndex := characterRosterIndexByID(ownedRoster, r.characterRosterAssignMenuID)
	if menuIndex >= r.characterRosterOwnedScroll &&
		menuIndex < r.characterRosterOwnedScroll+characterRosterOwnedVisibleRows(ownedViewport, ownedRoster, r.characterRosterOwnedScroll) &&
		characterRosterCouncilManageable(ownedRoster[menuIndex]) {
		if option := characterRosterAssignmentOptionAtForRoles(x, y, ownedColumn, menuIndex, r.characterRosterOwnedScroll, roles, ownedRoster); option >= 0 {
			character := ownedRoster[menuIndex]
			if characterRosterCanAssignRole(r.gs, character, roles[option]) {
				r.characterRosterAssignMenuID = ""
				return InputAction{Kind: ActionAssignCouncilRole, CharacterID: character.ID, CouncilRole: roles[option].ID}, true
			}
			return InputAction{}, true
		}
		if characterCouncilAssignButton(ownedColumn, menuIndex, r.characterRosterOwnedScroll, ownedRoster).HitTest(x, y) {
			r.characterRosterAssignMenuID = ""
			return InputAction{}, true
		}
		r.characterRosterAssignMenuID = ""
	}

	if index := characterRosterOwnedRowAt(x, y, ownedRoster, r.characterRosterOwnedScroll, ownedViewport); index >= 0 {
		character := ownedRoster[index]
		if character.ActiveMissionID == "" && (len(character.PendingAdventureAbilities) > 0 || len(r.characterMissionOptionsFor(character.ID)) > 0) &&
			characterRosterMissionButton(ownedColumn, index, r.characterRosterOwnedScroll, ownedRoster).HitTest(x, y) {
			r.OpenCharacterMissionMenu(character.ID)
			return InputAction{}, true
		}
		if characterRosterReleaseable(character) && character.AssignedCouncilRole == "" &&
			characterReleaseButton(ownedColumn, index, r.characterRosterOwnedScroll, ownedRoster).HitTest(x, y) {
			r.confirmCharacterRelease(character)
			return InputAction{}, true
		}
		if characterRosterCouncilManageable(character) && characterRosterDutyRemovable(character) &&
			characterRosterDutyActionButton(ownedColumn, index, r.characterRosterOwnedScroll, character, ownedRoster).HitTest(x, y) {
			return InputAction{Kind: ActionClearCouncilRole, CharacterID: character.ID, CouncilRole: character.AssignedCouncilRole}, true
		}
		if characterRosterCouncilManageable(character) && character.AssignedCouncilRole == "" &&
			characterCanBeAssignedToAnyCouncilRole(r.gs, character.ID) &&
			characterCouncilAssignButton(ownedColumn, index, r.characterRosterOwnedScroll, ownedRoster).HitTest(x, y) {
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

func handleCharacterRosterOutsideClick(r *Renderer, x, y float64, leftPressed, leftWasPressed bool) (InputAction, bool) {
	if r == nil || !leftPressed || leftWasPressed || r.characterRosterAssignMenuID == "" {
		return InputAction{}, false
	}
	action, handled := handleCharacterRosterPanelClick(r, x, y)
	if handled {
		if r.prevMouse == nil {
			r.prevMouse = make(map[ebiten.MouseButton]bool)
		}
		r.prevMouse[ebiten.MouseButtonLeft] = true
		return action, true
	}
	r.characterRosterAssignMenuID = ""
	return InputAction{}, false
}

func drawCharacterRosterScrollbar(screen *ebiten.Image, viewport gameui.Rect, count, scroll int, ownedCharacters []*army.Commander) {
	maxScroll := characterRosterMaxScroll(count, viewport)
	visibleRows := characterRosterVisibleRows(viewport)
	if ownedCharacters != nil {
		maxScroll = characterRosterMaxOwnedScroll(viewport, ownedCharacters)
		visibleRows = characterRosterOwnedVisibleRows(viewport, ownedCharacters, scroll)
	}
	if maxScroll <= 0 || viewport.H <= 0 {
		return
	}
	track := characterRosterScrollbarTrack(viewport)
	thumbH := track.H * float64(visibleRows) / float64(count)
	if thumbH < 18 {
		thumbH = 18
	}
	if thumbH > track.H {
		thumbH = track.H
	}
	if ownedCharacters != nil {
		scroll = clampCharacterRosterOwnedScroll(scroll, viewport, ownedCharacters)
	} else {
		scroll = clampCharacterRosterScroll(scroll, count, viewport)
	}
	thumbY := track.Y
	if track.H > thumbH {
		thumbY += (track.H - thumbH) * float64(scroll) / float64(maxScroll)
	}
	drawRoundedRect(screen, float32(track.X), float32(track.Y), float32(track.W), float32(track.H), 2, color.RGBA{70, 65, 55, 180})
	drawRoundedRect(screen, float32(track.X), float32(thumbY), float32(track.W), float32(thumbH), 2, color.RGBA{210, 175, 85, 230})
}

func drawCharacterRosterRows(screen *ebiten.Image, gs *state.GameState, missionOptions map[string][]scenario.CharacterMissionDefinition, column, viewport gameui.Rect, characters []*army.Commander, scroll int, candidate bool, assignmentMenuCharacterID, selectedCharacterID string) {
	if len(characters) == 0 {
		message := "Karakter bulunmuyor."
		if candidate {
			message = "Tarafsız aday yok."
		}
		drawUILabel(screen, viewport, message, ColorGray, gameui.TextSmall, gameui.TextAlignCenter)
		return
	}

	if candidate {
		scroll = clampCharacterRosterScroll(scroll, len(characters), viewport)
	} else {
		scroll = clampCharacterRosterOwnedScroll(scroll, viewport, characters)
	}
	visibleRows := characterRosterVisibleRows(viewport)
	if !candidate {
		visibleRows = characterRosterOwnedVisibleRows(viewport, characters, scroll)
	}
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
		visibleIndex := index - scroll
		row := characterRosterRowRect(viewport, visibleIndex)
		if !candidate {
			row = characterRosterOwnedRowRect(viewport, index, scroll, characters)
			if dividerY, showDivider := characterRosterDutyGroupDividerY(viewport, characters, index, scroll); showDivider {
				dividerX := row.X + characterRosterCardPadding
				dividerEndX := row.X + row.W - characterRosterCardPadding
				vector.StrokeLine(body, float32(dividerX), float32(dividerY), float32(dividerEndX), float32(dividerY), 1, color.RGBA{174, 132, 58, 190}, false)
			}
		}
		rowBG := color.RGBA{35, 26, 14, 210}
		rowBorder := color.RGBA{100, 75, 30, 180}
		vector.FillRect(body, float32(row.X), float32(row.Y), float32(row.W), float32(row.H), rowBG, false)
		border := characterRosterRowBorderRect(row)
		vector.StrokeRect(body, float32(border.X), float32(border.Y), float32(border.W), float32(border.H), 1, rowBorder, false)
		if !candidate && character.ID == selectedCharacterID {
			vector.StrokeRect(body, float32(border.X), float32(border.Y), float32(border.W), float32(border.H), 2.5, ColorGold, false)
		}
		portrait := characterRosterPortraitRect(row)
		drawCommanderPortrait(body, character, portrait.X, portrait.Y, portrait.W, portrait.H)
		if character != nil && len(character.AdventureAbilities) > 0 {
			drawCharacterRosterAbilityBadge(body, portrait)
		}

		textX := portrait.X + portrait.W + 10
		textW := row.X + row.W - textX - characterRosterCardPadding
		nameW := textW
		var adopt gameui.Button
		var missionButton gameui.Button
		hasMissionButton := !candidate && character.ActiveMissionID == "" &&
			(len(character.PendingAdventureAbilities) > 0 || len(missionOptions[character.ID]) > 0)
		if candidate {
			adopt = characterAdoptButton(column, index, scroll)
			nameW = adopt.X - textX - 5
		} else if characterRosterCouncilManageable(character) && (character.AssignedCouncilRole != "" ||
			characterCanBeAssignedToAnyCouncilRole(gs, character.ID)) {
			button := characterRosterDutyActionButton(column, index, scroll, character, characters)
			nameW = button.X - textX - 5
		}
		if hasMissionButton {
			missionButton = characterRosterMissionButton(column, index, scroll, characters)
			if availableWidth := missionButton.X - textX - 5; availableWidth < nameW {
				nameW = availableWidth
			}
		}
		drawUILabel(body, gameui.Rect{X: textX, Y: row.Y + 8, W: nameW, H: 18}, trimTextToWidth(character.Name, FaceSmall, nameW), ColorWhite, gameui.TextSmall, gameui.TextAlignStart)
		kind := "Karakter"
		switch character.Type {
		case army.CharacterTypeAdministrator:
			kind = "Yönetici"
		case army.CharacterTypeCommander:
			kind = "Komutan"
		}
		if !candidate && character.AssignedArmyID != "" {
			kind += " · Orduya atanmış"
		}
		detail := kind
		var missionDetails [2]string
		var bonus []string
		upkeepLabel := characterRosterUpkeepLabel(character, scenario.CouncilRoleDefinition{})
		if candidate {
			detail = "Tarafsız aday"
		} else if character.Type == army.CharacterTypeCommander {
			if role, ok := gs.CouncilConfig.WithDefaults().Role(character.AssignedCouncilRole); ok {
				upkeepLabel = characterRosterUpkeepLabel(character, role)
				detail = role.NameTR + " · " + upkeepLabel
				bonus = characterCouncilEffectsLinesTR(gs.CouncilRoleEffects(gs.PlayerFactionID, character.AssignedCouncilRole))
			} else {
				detail = kind
			}
		} else if character.AssignedCouncilRole != "" {
			if role, ok := gs.CouncilConfig.WithDefaults().Role(character.AssignedCouncilRole); ok {
				upkeepLabel = characterRosterUpkeepLabel(character, role)
				detail = role.NameTR + " · " + upkeepLabel
				bonus = characterCouncilEffectsLinesTR(gs.CouncilRoleEffects(gs.PlayerFactionID, character.AssignedCouncilRole))
			} else {
				detail = character.AssignedCouncilRole + " · " + upkeepLabel
			}
		} else {
			detail = kind + " · " + upkeepLabel
		}
		if character.ActiveMissionID != "" {
			missionName := character.ActiveMissionID
			for _, mission := range gs.CouncilConfig.Missions {
				if mission.ID == character.ActiveMissionID {
					missionName = mission.NameTR
					break
				}
			}
			missionDetails = characterRosterMissionDetailLines(missionName, character.ActiveMissionTurnsRemaining)
			bonus = nil
		} else if len(character.PendingAdventureAbilities) > 0 {
			detail = "Yetenek seçimi bekliyor"
		}
		detailRect := characterRosterDetailRect(row, textX)
		if hasMissionButton {
			detailRect.W = missionButton.X - detailRect.X - 5
		}
		if candidate {
			costLines := characterRosterCandidateCostLines(character)
			drawUILabel(body, detailRect, trimTextToWidth(costLines[0], FaceSmall, detailRect.W), ColorGold, gameui.TextSmall, gameui.TextAlignStart)
			upkeepRect := detailRect
			upkeepRect.Y += 18
			drawUILabel(body, upkeepRect, trimTextToWidth(costLines[1], FaceSmall, upkeepRect.W), ColorGray, gameui.TextSmall, gameui.TextAlignStart)
		} else {
			if missionDetails[0] != "" {
				drawUILabel(body, detailRect, trimTextToWidth(missionDetails[0], FaceSmall, detailRect.W), ColorGray, gameui.TextSmall, gameui.TextAlignStart)
				missionTurnsRect := detailRect
				missionTurnsRect.Y += 18
				drawUILabel(body, missionTurnsRect, trimTextToWidth(missionDetails[1], FaceSmall, missionTurnsRect.W), ColorGray, gameui.TextSmall, gameui.TextAlignStart)
			} else {
				drawUILabel(body, detailRect, trimTextToWidth(detail, FaceSmall, detailRect.W), ColorGray, gameui.TextSmall, gameui.TextAlignStart)
			}
		}
		for bonusIndex, bonusLine := range bonus {
			bonusRect := gameui.Rect{
				X: textX,
				Y: detailRect.Y + 20 + float64(bonusIndex)*16,
				W: row.X + row.W - textX - characterRosterCardPadding,
				H: 16,
			}
			drawUILabel(body, bonusRect, trimTextToWidth(bonusLine, FaceSmall, bonusRect.W), color.RGBA{165, 205, 145, 255}, gameui.TextSmall, gameui.TextAlignStart)
		}
		abilityEffects := characterAdventureAbilityEffectsSummaryTR(character.AdventureAbilities)
		if abilityEffects != "" {
			abilityLineY := detailRect.Y + 20 + float64(len(bonus))*16
			if missionDetails[0] != "" && abilityLineY < detailRect.Y+36 {
				abilityLineY = detailRect.Y + 36
			}
			abilityRect := gameui.Rect{
				X: textX,
				Y: abilityLineY,
				W: row.X + row.W - textX - characterRosterCardPadding,
				H: 16,
			}
			if !candidate && characterRosterReleaseable(character) &&
				character.AssignedCouncilRole == "" && character.ID != assignmentMenuCharacterID {
				abilityRect = characterRosterLineRectBeforeButton(
					row, textX, abilityLineY, 16,
					characterReleaseButton(column, index, scroll, characters),
				)
			}
			drawUILabel(body, abilityRect, trimTextToWidth(abilityEffects, FaceSmall, abilityRect.W),
				color.RGBA{190, 220, 155, 255}, gameui.TextSmall, gameui.TextAlignStart)
		}
		traitRow := characterRosterTraitRowRect(row)
		if hasMissionButton {
			traitRow.W = missionButton.X - traitRow.X - characterRosterCardPadding
		}
		drawCharacterTraitBadges(body, character, traitRow.X, traitRow.Y, traitRow.W, characterTraitBadgeOptions{MaxRows: 1}, gs.CouncilConfig)
		if candidate {
			drawUIButtonWidget(body, adopt, applyTinyButtonStyle)
		} else {
			if characterRosterReleaseable(character) && character.AssignedCouncilRole == "" && character.ID != assignmentMenuCharacterID {
				drawUIButtonWidget(body, characterReleaseButton(column, index, scroll, characters), dangerTinyButtonStyle)
			}
			if characterRosterCouncilManageable(character) && characterRosterDutyRemovable(character) {
				drawUIButtonWidget(body, characterRosterDutyActionButton(column, index, scroll, character, characters), tinyButtonStyle)
			} else if characterRosterCouncilManageable(character) && characterCanBeAssignedToAnyCouncilRole(gs, character.ID) {
				drawUIButtonWidget(body, characterRosterDutyActionButton(column, index, scroll, character, characters), tinyButtonStyle)
			}
			if hasMissionButton {
				drawUIButtonWidget(body, missionButton, tinyButtonStyle)
			}
		}
	}
	ownedCharacters := characters
	if candidate {
		ownedCharacters = nil
	}
	drawCharacterRosterScrollbar(screen, viewport, len(characters), scroll, ownedCharacters)
}

func drawCharacterRosterAssignmentMenu(screen *ebiten.Image, gs *state.GameState, column gameui.Rect, rowIndex, scroll int, characters []*army.Commander, roles []scenario.CouncilRoleDefinition) {
	menu := characterRosterAssignmentMenuRectWithCount(column, rowIndex, scroll, len(roles), characters)
	drawUICardRect(screen, menu, color.RGBA{24, 20, 14, 250}, color.RGBA{174, 132, 58, 255}, 1)
	cursorX, cursorY := ebiten.CursorPosition()
	var character *army.Commander
	if rowIndex >= 0 && rowIndex < len(characters) {
		character = characters[rowIndex]
	}
	for optionIndex, role := range roles {
		option := characterRosterAssignmentOptionButtonForRole(column, rowIndex, scroll, optionIndex, len(roles), role.NameTR, characters)
		enabled := characterRosterCanAssignRole(gs, character, role)
		if enabled && option.HitTest(float64(cursorX), float64(cursorY)) {
			vector.FillRect(screen, float32(option.X), float32(option.Y), float32(option.W), float32(option.H), color.RGBA{86, 63, 30, 150}, false)
		}
		if optionIndex > 0 {
			separatorY := float32(option.Y)
			vector.StrokeLine(screen, float32(menu.X+8), separatorY, float32(menu.X+menu.W-8), separatorY, 1, color.RGBA{92, 72, 43, 170}, false)
		}
		labelRect := gameui.Rect{X: option.X + 8, Y: option.Y + 4, W: option.W - 16, H: option.H - 8}
		labelColor := color.RGBA{235, 220, 185, 255}
		if !enabled {
			labelColor = color.RGBA{128, 118, 101, 210}
		}
		drawUILabel(screen, labelRect, option.Label, labelColor, gameui.TextSmall, gameui.TextAlignStart)
	}
}

func (r *Renderer) drawCharacterRosterPanel(screen *ebiten.Image) {
	if r == nil || r.gs == nil || !r.showCharacterRoster {
		return
	}
	r.ensureCharacterRosterPanelCache()
	panel := characterRosterPanelRect()
	drawUIPanelFrame(screen, panel, color.RGBA{12, 10, 8, 244}, color.RGBA{154, 112, 54, 255}, 1.5, 5)
	drawUILabel(screen, gameui.Rect{X: panel.X + characterRosterPanelPad, Y: panel.Y + 9, W: panel.W - 60}, "Hane Halkı", color.RGBA{255, 220, 118, 255}, gameui.TextMedium, gameui.TextAlignStart)
	drawCloseButton(screen, characterRosterCloseButton())

	ownedColumn, candidateColumn := characterRosterColumns()
	ownedRoster := r.characterRosterOwned
	candidates := r.characterRosterCandidates
	roles := characterCouncilRoles(r.gs)
	ownedViewport := characterRosterOwnedListViewport(ownedColumn)
	candidateViewport := characterRosterListViewport(candidateColumn)
	r.characterRosterOwnedScroll = clampCharacterRosterOwnedScroll(r.characterRosterOwnedScroll, ownedViewport, ownedRoster)
	r.characterRosterNeutralScroll = clampCharacterRosterScroll(r.characterRosterNeutralScroll, len(candidates), candidateViewport)

	drawUICardRect(screen, ownedColumn, color.RGBA{20, 16, 11, 220}, color.RGBA{92, 72, 40, 190}, 1)
	drawUICardRect(screen, candidateColumn, color.RGBA{20, 16, 11, 220}, color.RGBA{92, 72, 40, 190}, 1)
	drawUILabel(screen, gameui.Rect{X: ownedColumn.X + 8, Y: ownedColumn.Y + 7, W: ownedColumn.W - 16}, "Hane Üyeleri ("+itoa(len(ownedRoster))+")", ColorGold, gameui.TextSmall, gameui.TextAlignStart)
	drawUILabel(screen, gameui.Rect{X: candidateColumn.X + 8, Y: candidateColumn.Y + 7, W: candidateColumn.W - 16}, "Adaylar ("+itoa(len(candidates))+")", ColorGold, gameui.TextSmall, gameui.TextAlignStart)
	drawUILabel(screen, gameui.Rect{X: candidateColumn.X + 8, Y: candidateColumn.Y + 34, W: candidateColumn.W - 16, H: 32}, "Alım ve tur başı bakım ücreti kartta", ColorGray, gameui.TextSmall, gameui.TextAlignStart)
	drawCharacterRosterRows(screen, r.gs, r.characterMissionsByCharacter, ownedColumn, ownedViewport, ownedRoster, r.characterRosterOwnedScroll, false, r.characterRosterAssignMenuID, r.characterRosterSelectedCharacterID)
	drawCharacterRosterRows(screen, r.gs, r.characterMissionsByCharacter, candidateColumn, candidateViewport, candidates, r.characterRosterNeutralScroll, true, "", "")
	menuIndex := characterRosterIndexByID(ownedRoster, r.characterRosterAssignMenuID)
	if menuIndex >= r.characterRosterOwnedScroll &&
		menuIndex < r.characterRosterOwnedScroll+characterRosterOwnedVisibleRows(ownedViewport, ownedRoster, r.characterRosterOwnedScroll) &&
		characterRosterCouncilManageable(ownedRoster[menuIndex]) {
		drawCharacterRosterAssignmentMenu(screen, r.gs, ownedColumn, menuIndex, r.characterRosterOwnedScroll, ownedRoster, roles)
	}
	r.drawCharacterMissionMenu(screen)
	drawCharacterRosterAbilityPopup(screen, r)
}

func (r *Renderer) handleCharacterRosterPanelInput() (InputAction, bool) {
	if r == nil || r.gs == nil || !r.showCharacterRoster {
		return InputAction{}, false
	}
	r.ensureCharacterRosterPanelCache()
	if r.keyJustPressed(ebiten.KeyEscape) {
		if r.characterRosterMissionMenuCharacterID != "" {
			if r.characterRosterSelectedMissionID != "" {
				r.characterRosterSelectedMissionID = ""
				r.characterRosterMissionTargets = nil
				r.characterRosterMissionScroll = 0
			} else {
				r.CloseCharacterMissionMenu()
			}
			return InputAction{}, true
		}
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
		if action, handled := handleCharacterRosterOutsideClick(
			r, x, y,
			ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft),
			r.prevMouse[ebiten.MouseButtonLeft],
		); handled {
			return action, true
		}
		return InputAction{}, false
	}
	ownedColumn, candidateColumn := characterRosterColumns()
	ownedRoster := r.characterRosterOwned
	candidates := r.characterRosterCandidates
	ownedViewport := characterRosterOwnedListViewport(ownedColumn)
	candidateViewport := characterRosterListViewport(candidateColumn)
	if _, wheelY := ebiten.Wheel(); wheelY != 0 {
		if r.scrollCharacterMissionMenu(x, y, wheelY) {
			return InputAction{}, true
		}
		r.characterRosterAssignMenuID = ""
		switch {
		case ownedViewport.Hit(x, y):
			r.characterRosterOwnedScroll = clampCharacterRosterOwnedScroll(r.characterRosterOwnedScroll-int(wheelY), ownedViewport, ownedRoster)
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
	r.openCharacterRosterPanel("")
}

func (r *Renderer) OpenCharacterRosterPanelForCharacter(characterID string) {
	r.openCharacterRosterPanel(characterID)
}

func (r *Renderer) openCharacterRosterPanel(characterID string) {
	if r == nil || r.gs == nil || r.gs.PlayerFactionID == "" {
		return
	}
	r.showCharacterRoster = true
	r.characterRosterSelectedCharacterID = characterID
	r.characterRosterAssignMenuID = ""
	r.CloseCharacterMissionMenu()
	r.characterRosterOwnedScroll = 0
	r.characterRosterNeutralScroll = 0
	r.RefreshCharacterRosterPanel()
	if index := characterRosterIndexByID(r.characterRosterOwned, characterID); index >= 0 {
		ownedColumn, _ := characterRosterColumns()
		viewport := characterRosterOwnedListViewport(ownedColumn)
		r.characterRosterOwnedScroll = clampCharacterRosterOwnedScroll(index, viewport, r.characterRosterOwned)
	}
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
	r.characterMissionsByCharacter = make(map[string][]scenario.CharacterMissionDefinition, len(r.characterRosterOwned))
	for _, character := range r.characterRosterOwned {
		if missions := r.gs.CharacterMissionOptions(r.gs.PlayerFactionID, character.ID); len(missions) > 0 {
			r.characterMissionsByCharacter[character.ID] = missions
		}
	}
	r.characterRosterCacheValid = true
}

func (r *Renderer) CloseCharacterRosterPanel() {
	if r == nil {
		return
	}
	r.showCharacterRoster = false
	r.characterRosterSelectedCharacterID = ""
	r.characterRosterAssignMenuID = ""
	r.characterRosterOwnedScroll = 0
	r.characterRosterNeutralScroll = 0
	r.characterRosterOwned = nil
	r.characterRosterCandidates = nil
	r.characterRosterCacheValid = false
	r.characterMissionsByCharacter = nil
	r.CloseCharacterMissionMenu()
}

func (r *Renderer) characterRosterPanelHovering(x, y float64) bool {
	return characterRosterPanelInteractiveHit(r, x, y)
}
