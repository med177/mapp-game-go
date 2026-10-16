package render

import (
	"fmt"
	"strings"
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/scenario"
	"mapp-game-go/internal/state"
	gameui "mapp-game-go/internal/ui"
	"mapp-game-go/internal/world"
)

func TestCharacterRosterPortraitMatchesCommanderPortraitSize(t *testing.T) {
	row := characterRosterRowRect(gameui.Rect{X: 20, Y: 30, W: 300, H: 500}, 0)
	portrait := characterRosterPortraitRect(row)
	wantSize := float64(armyPanelCommanderPortrait)
	if portrait.W != wantSize || portrait.H != wantSize {
		t.Fatalf("character portrait = %.0fx%.0f, want %.0fx%.0f", portrait.W, portrait.H, wantSize, wantSize)
	}
	if portrait.Y < row.Y || portrait.Y+portrait.H > row.Y+row.H {
		t.Fatalf("character portrait %v does not fit row %v", portrait, row)
	}
	if portrait.X-row.X != characterRosterCardPadding ||
		row.X+row.W-(portrait.X+portrait.W) < characterRosterCardPadding {
		t.Fatalf("portrait %v does not respect the shared horizontal card padding %v", portrait, characterRosterCardPadding)
	}
	traits := characterRosterTraitRowRect(row)
	if portrait.Y-row.Y != characterRosterCardPadding ||
		traits.Y-(portrait.Y+portrait.H) != characterRosterCardPadding {
		t.Fatalf("portrait %v and trait row %v do not have equal top/inner spacing in row %v", portrait, traits, row)
	}
	if traits.X-row.X != characterRosterCardPadding ||
		row.X+row.W-(traits.X+traits.W) != characterRosterCardPadding ||
		row.Y+row.H-(traits.Y+traits.H) != characterRosterCardPadding {
		t.Fatalf("trait row %v does not use equal card-edge padding in row %v", traits, row)
	}
}

func TestCharacterRosterAbilityBadgeOverlaysPortraitBottomRight(t *testing.T) {
	row := characterRosterRowRect(gameui.Rect{X: 20, Y: 30, W: 300, H: 500}, 0)
	portrait := characterRosterPortraitRect(row)
	badge := characterRosterAbilityBadgeRect(portrait)
	if badge.X+badge.W != portrait.X+portrait.W-characterAbilityBadgePadding ||
		badge.Y+badge.H != portrait.Y+portrait.H-characterAbilityBadgePadding {
		t.Fatalf("ability badge %v does not keep %vpx inset from portrait bottom-right %v",
			badge, characterAbilityBadgePadding, portrait)
	}
	if badge.X < portrait.X || badge.Y < portrait.Y ||
		badge.X+badge.W > portrait.X+portrait.W ||
		badge.Y+badge.H > portrait.Y+portrait.H {
		t.Fatalf("ability badge %v falls outside portrait %v", badge, portrait)
	}
	if badge.H != characterAbilityBadgeHeight {
		t.Fatalf("ability badge height = %.2f, want %.2f", badge.H, characterAbilityBadgeHeight)
	}
	if badge.W != characterAbilityIconWidth {
		t.Fatalf("ability badge width = %.2f, want %.2f", badge.W, characterAbilityIconWidth)
	}
}

func TestCharacterRosterViewportsReserveScrollbarGutter(t *testing.T) {
	column := gameui.Rect{X: 20, Y: 30, W: 240, H: 400}
	candidateViewport := characterRosterListViewport(column)
	if candidateViewport.Y != column.Y+34+32+8 {
		t.Fatalf("candidate list starts at y=%.0f, want 8px below the description at y=%.0f", candidateViewport.Y, column.Y+34+32+8)
	}
	for name, viewport := range map[string]gameui.Rect{
		"owned":      characterRosterOwnedListViewport(column),
		"candidates": candidateViewport,
	} {
		track := characterRosterScrollbarTrack(viewport)
		row := characterRosterRowRect(viewport, 0)
		if row.X+row.W+characterRosterScrollbarGap > track.X {
			t.Errorf("%s card ends at %.0f and overlaps or touches scrollbar at %v", name, row.X+row.W, track)
		}
		if track.X < column.X || track.X+track.W+characterRosterScrollbarRight > column.X+column.W {
			t.Errorf("%s scrollbar %v does not fit inside column %v with a right margin", name, track, column)
		}
	}
}

func TestCharacterRosterAssignedAndUnassignedRowsHaveDividerGap(t *testing.T) {
	viewport := gameui.Rect{X: 20, Y: 40, W: 200, H: 500}
	assigned := &army.Commander{AssignedCouncilRole: state.CouncilRoleTreasurer}
	unassigned := &army.Commander{}
	rows := []*army.Commander{assigned, unassigned}
	previous := characterRosterOwnedRowRect(viewport, 0, 0, rows)
	current := characterRosterOwnedRowRect(viewport, 1, 0, rows)
	dividerY, showDivider := characterRosterDutyGroupDividerY(viewport, rows, 1, 0)
	if !showDivider {
		t.Fatal("assigned-to-unassigned transition should show a divider")
	}
	if above, below := dividerY-(previous.Y+previous.H), current.Y-dividerY; above < characterRosterDutyGroupPad || below < characterRosterDutyGroupPad {
		t.Fatalf("group divider spacing = %.0f above / %.0f below, want at least %.0f on both sides", above, below, characterRosterDutyGroupPad)
	}
	if current.Y-previous.Y != characterRosterRowH+characterRosterDutyGroupGap {
		t.Fatalf("row pitch across groups = %.0f, want base pitch %.0f plus group gap %.0f", current.Y-previous.Y, characterRosterRowH, characterRosterDutyGroupGap)
	}
	if gap := characterRosterRowH - previous.H; gap != 8 {
		t.Fatalf("regular card gap = %.0f, want 8px", gap)
	}
	if above, below := dividerY-(previous.Y+previous.H), current.Y-dividerY; above != 8 || below != 8 {
		t.Fatalf("divider spacing = %.0f above / %.0f below, want 8px each", above, below)
	}
	if _, showDivider := characterRosterDutyGroupDividerY(viewport, []*army.Commander{unassigned, unassigned}, 1, 0); showDivider {
		t.Fatal("unassigned-to-unassigned rows should not show a divider")
	}
	if assigned.AssignedCouncilRole == "" || unassigned.AssignedCouncilRole != "" {
		t.Fatal("test rows do not represent assigned and unassigned council duties")
	}
}

func TestCharacterRosterListsShowPartialNextCardAtViewportBottom(t *testing.T) {
	viewport := gameui.Rect{X: 20, Y: 40, W: 200, H: 5*characterRosterRowH + 64}
	if got, want := characterRosterVisibleRows(viewport), 6; got != want {
		t.Fatalf("candidate visible rows = %d, want %d to render the partial bottom card", got, want)
	}
	partial := characterRosterRowRect(viewport, 5)
	if partial.Y >= viewport.Y+viewport.H || partial.Y+partial.H <= viewport.Y+viewport.H {
		t.Fatalf("last visible candidate row %v should be clipped at viewport bottom %.0f", partial, viewport.Y+viewport.H)
	}
	characters := make([]*army.Commander, 6)
	for index := range characters {
		characters[index] = &army.Commander{ID: itoa(index)}
	}
	if got := characterRosterOwnedVisibleRows(viewport, characters, 0); got != 6 {
		t.Fatalf("owned visible rows = %d, want 6 including partial bottom card", got)
	}
	candidates := make([]*army.Commander, 12)
	for index := range candidates {
		candidates[index] = &army.Commander{ID: itoa(index)}
	}
	scroll := characterRosterMaxScroll(len(candidates), viewport)
	lastCandidate := characterRosterRowRect(viewport, len(candidates)-1-scroll)
	if lastCandidate.Y >= viewport.Y+viewport.H || lastCandidate.Y+lastCandidate.H <= viewport.Y+viewport.H {
		t.Fatalf("last candidate row %v should remain partially visible at max scroll in viewport %v", lastCandidate, viewport)
	}
}

func TestCharacterRosterOwnedRowsKeepSpacingAndScrollGeometryAligned(t *testing.T) {
	viewport := gameui.Rect{X: 20, Y: 40, W: 200, H: 450}
	characters := []*army.Commander{
		{ID: "assigned", AssignedCouncilRole: state.CouncilRoleTreasurer},
		{ID: "free1"},
		{ID: "free2"},
		{ID: "free3"},
	}
	firstFree := characterRosterOwnedRowRect(viewport, 1, 0, characters)
	secondFree := characterRosterOwnedRowRect(viewport, 2, 0, characters)
	if secondFree.Y-firstFree.Y != characterRosterRowH {
		t.Fatalf("rows in same group have pitch %.0f, want %.0f", secondFree.Y-firstFree.Y, characterRosterRowH)
	}
	scrolledFirst := characterRosterOwnedRowRect(viewport, 1, 1, characters)
	if scrolledFirst.Y != viewport.Y {
		t.Fatalf("first scrolled row y=%.0f, want viewport top %.0f", scrolledFirst.Y, viewport.Y)
	}
	if got := characterRosterOwnedRowAt(firstFree.X+1, firstFree.Y+1, characters, 0, viewport); got != 1 {
		t.Fatalf("hit on first unassigned row = %d, want 1", got)
	}
	maxScroll := characterRosterMaxOwnedScroll(viewport, characters)
	lastRow := characterRosterOwnedRowRect(viewport, len(characters)-1, maxScroll, characters)
	if lastRow.Y+lastRow.H > viewport.Y+viewport.H {
		t.Fatalf("last row %v is clipped at max owned scroll %d in viewport %v", lastRow, maxScroll, viewport)
	}
}

func TestCharacterRosterRowBorderStaysInsideClippedRow(t *testing.T) {
	viewport := gameui.Rect{X: 100, Y: 200, W: 300, H: 500}
	row := characterRosterRowRect(viewport, 0)
	border := characterRosterRowBorderRect(row)
	if border.X <= viewport.X || border.Y <= viewport.Y ||
		border.X+border.W >= viewport.X+viewport.W ||
		border.Y+border.H >= viewport.Y+viewport.H {
		t.Fatalf("row border %v touches or crosses clipped viewport %v", border, viewport)
	}
	if border.W != row.W-2 || border.H != row.H-2 {
		t.Fatalf("row border %v should inset the row %v by one pixel", border, row)
	}
}

func TestCharacterAdoptButtonIsIconOnlyAtBottomRight(t *testing.T) {
	column := gameui.Rect{X: 20, Y: 30, W: 240, H: 400}
	row := characterRosterRowRect(characterRosterListViewport(column), 0)
	button := characterAdoptButton(column, 0, 0)
	if button.Label != "" || button.Icon != gameui.IconPlus {
		t.Fatalf("adopt button content = label %q, icon %q; want icon only", button.Label, button.Icon)
	}
	traits := characterRosterTraitRowRect(row)
	portrait := characterRosterPortraitRect(row)
	if button.X+button.W > row.X+row.W || button.X < row.X+row.W-button.W-6 ||
		button.Y+button.H > portrait.Y+portrait.H || button.Y < row.Y {
		t.Fatalf("adopt button %v should sit at the lower right beside the portrait %v", button, portrait)
	}
	if traits.Y < portrait.Y+portrait.H || traits.Y < button.Y+button.H ||
		traits.W <= button.W || traits.X+traits.W > row.X+row.W {
		t.Fatalf("trait row %v should span below portrait %v and join button %v", traits, portrait, button)
	}
}

func TestCharacterReleaseButtonAndConfirmation(t *testing.T) {
	column := gameui.Rect{X: 20, Y: 30, W: 240, H: 400}
	row := characterRosterRowRect(characterRosterOwnedListViewport(column), 0)
	button := characterReleaseButton(column, 0, 0)
	if button.Label != "" || button.Icon != gameui.IconClose {
		t.Fatalf("release button = %+v, want icon-only close button", button)
	}
	if button.X+button.W > row.X+row.W || button.Y+button.H > row.Y+row.H {
		t.Fatalf("release button %v does not fit in row %v", button, row)
	}

	r := &Renderer{}
	character := &army.Commander{ID: "chr_bronn", Name: "Bronn"}
	r.confirmCharacterRelease(character)
	if !r.confirmDialog.show || r.confirmDialog.pendingAction.Kind != ActionReleaseCharacter ||
		r.confirmDialog.pendingAction.CharacterID != character.ID ||
		r.confirmDialog.acceptLabel != "Evet" || r.confirmDialog.declineLabel != "Hayır" {
		t.Fatalf("release confirmation = %+v", r.confirmDialog)
	}
	if !strings.Contains(r.confirmDialog.message, character.Name) {
		t.Fatalf("release confirmation message %q omits character name", r.confirmDialog.message)
	}
}

func TestCharacterAdoptionConfirmation(t *testing.T) {
	r := &Renderer{}
	character := &army.Commander{ID: "chr_bronn", Name: "Bronn", AdoptionCostGold: 500, GoldUpkeep: 6}
	r.confirmCharacterAdoption(character)
	if !r.confirmDialog.show || r.confirmDialog.pendingAction.Kind != ActionAdoptNeutralCharacter ||
		r.confirmDialog.pendingAction.CharacterID != character.ID ||
		r.confirmDialog.acceptLabel != "Evet" || r.confirmDialog.declineLabel != "Hayır" {
		t.Fatalf("adoption confirmation = %+v", r.confirmDialog)
	}
	if !strings.Contains(r.confirmDialog.message, character.Name) {
		t.Fatalf("adoption confirmation message %q omits character name", r.confirmDialog.message)
	}
	if !strings.Contains(r.confirmDialog.message, "500 altın") ||
		!strings.Contains(r.confirmDialog.message, "6 altın/tur") {
		t.Fatalf("adoption confirmation message %q omits costs", r.confirmDialog.message)
	}
}

func TestCharacterRosterCandidateShowsOneTimeAndPerTurnCostsSeparately(t *testing.T) {
	character := &army.Commander{
		AdoptionCostGold: 500,
		GoldUpkeep:       6,
	}
	got := characterRosterCandidateCostLines(character)
	want := [2]string{
		"Alım: 500 altın (tek seferlik)",
		"Bakım: 6 altın/tur",
	}
	if got != want {
		t.Fatalf("candidate cost lines = %q, want %q", got, want)
	}
}

func TestCharacterRosterMissionDetailsUseSeparateLines(t *testing.T) {
	got := characterRosterMissionDetailLines("Diplomatik Elçilik", 2)
	want := [2]string{"Görevde: Diplomatik Elçilik", "Kalan süre: 2 tur"}
	if got != want {
		t.Fatalf("mission detail lines = %q, want %q", got, want)
	}
}

func TestCharacterAdventureAbilityLinesShowNameAndEffect(t *testing.T) {
	abilities := []army.CharacterAbility{
		{ID: "duel_veteran", NameTR: "Düello Ustalığı", Scope: "army", Effect: "army_power", Modifier: 0.03},
		{ID: "fieldcraft", NameTR: "Saha Keşfi", Scope: "faction", Effect: "military_power", Modifier: 0.02},
	}
	got := characterAdventureAbilityLinesTR(abilities)
	want := []string{"Düello Ustalığı (+%3 ordu gücü)", "Saha Keşfi (+%2 askerî güç)"}
	if len(got) != len(want) {
		t.Fatalf("adventure ability lines = %q, want %q", got, want)
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("adventure ability line %d = %q, want %q", index, got[index], want[index])
		}
	}
	if got := characterAdventureAbilityLinesTR([]army.CharacterAbility{{ID: "unnamed"}}); len(got) != 0 {
		t.Fatalf("unnamed adventure ability lines = %q, want empty", got)
	}
}

func TestCharacterAdventureAbilityEffectsSummaryOmitsAbilityNames(t *testing.T) {
	abilities := []army.CharacterAbility{
		{ID: "duel_veteran", NameTR: "Düello Ustalığı", Scope: "army", Effect: "army_power", Modifier: 0.03},
		{ID: "fieldcraft", NameTR: "Saha Keşfi", Scope: "faction", Effect: "military_power", Modifier: 0.02},
	}
	if got, want := characterAdventureAbilityEffectsSummaryTR(abilities), "+%3 ordu gücü · +%2 askerî güç"; got != want {
		t.Fatalf("adventure ability effect summary = %q, want %q", got, want)
	}
}

func TestCharacterRosterCaptiveCannotBeManagedOrReleased(t *testing.T) {
	character := &army.Commander{
		Type: army.CharacterTypeCharacter, CaptiveAtSettlementID: "holding_main",
	}
	if characterRosterCouncilManageable(character) {
		t.Fatal("tutsak karakter konsey görevinde yönetilebilir")
	}
	if characterRosterReleaseable(character) {
		t.Fatal("tutsak karakter hane listesinden serbest bırakılabilir")
	}
	if characterRosterDutyRemovable(character) {
		t.Fatal("tutsak karakterin konsey görevi düzenlenebilir")
	}
}

func TestCharacterRosterAbilityEffectLineAvoidsReleaseButton(t *testing.T) {
	row := gameui.Rect{X: 100, Y: 200, W: 400, H: characterRosterCardH}
	button := gameui.NewButton(row.X+row.W-34, row.Y+59, 28, 28, "")
	line := characterRosterLineRectBeforeButton(row, row.X+150, row.Y+54, 16, button)
	if line.X+line.W+4 != button.X {
		t.Fatalf("ability line ends at %.0f; want 4px gap before release button at %.0f", line.X+line.W, button.X)
	}
	if line.Y >= button.Y+button.H || line.Y+line.H <= button.Y {
		t.Fatal("test ability line should vertically overlap release button")
	}
}

func TestCharacterRosterAbilityEffectLineDoesNotOverlapTraits(t *testing.T) {
	row := gameui.Rect{X: 100, Y: 200, W: 400, H: characterRosterCardH}
	detail := characterRosterDetailRect(row, row.X+150)

	line, fits := characterRosterAbilityEffectLineRect(row, detail.X, detail, 2, false, false)
	if !fits || line.Y+line.H > characterRosterTraitRowRect(row).Y {
		t.Fatalf("ability line %v should fit above trait row %v", line, characterRosterTraitRowRect(row))
	}

	line, fits = characterRosterAbilityEffectLineRect(row, detail.X, detail, 3, false, false)
	if fits || line.Y+line.H <= characterRosterTraitRowRect(row).Y {
		t.Fatalf("ability line %v should be omitted rather than overlap trait row %v", line, characterRosterTraitRowRect(row))
	}

	line, fits = characterRosterAbilityEffectLineRect(row, detail.X, detail, 0, false, true)
	if !fits || line.Y != detail.Y+36 || line.Y+line.H > characterRosterTraitRowRect(row).Y {
		t.Fatalf("candidate ability line %v should appear after both cost lines and before traits %v", line, characterRosterTraitRowRect(row))
	}
}

func TestCharacterAbilityPopupExpandsForEveryAbility(t *testing.T) {
	popup := characterAbilityPopupRect(100, 100, 10)
	wantHeight := characterAbilityPopupPad*2 + characterAbilityPopupTitleH + 10*characterAbilityPopupLineH
	if popup.H != wantHeight {
		t.Fatalf("10-ability popup height = %.0f, want %.0f", popup.H, wantHeight)
	}
	if popup.W != characterAbilityPopupWidth {
		t.Fatalf("ability popup width = %.0f, want %.0f", popup.W, characterAbilityPopupWidth)
	}
}

func TestCharacterRosterAbilityPortraitHoverShowsPointer(t *testing.T) {
	const owner faction.FactionID = "player"
	character := &army.Commander{
		ID: "chr_bronn", Type: army.CharacterTypeCharacter, OwnerID: string(owner), Name: "Bronn",
	}
	for index := 0; index < 10; index++ {
		character.AdventureAbilities = append(character.AdventureAbilities, army.CharacterAbility{
			ID: "ability_" + itoa(index), NameTR: "Yetenek " + itoa(index),
			Scope: "army", Effect: "army_power", Modifier: 0.01,
		})
	}
	r := &Renderer{
		gs: &state.GameState{
			Year: 300, PlayerFactionID: owner,
			Factions: map[faction.FactionID]*faction.Faction{owner: {ID: owner}},
		},
		showCharacterRoster:       true,
		characterRosterOwned:      []*army.Commander{character},
		characterRosterCacheValid: true,
	}
	ownedColumn, _ := characterRosterColumns()
	viewport := characterRosterOwnedListViewport(ownedColumn)
	row := characterRosterOwnedRowRect(viewport, 0, 0, r.characterRosterOwned)
	portrait := characterRosterPortraitRect(row)
	x, y := portrait.X+portrait.W/2, portrait.Y+portrait.H/2

	gotCharacter, lines := characterRosterAbilityPortraitAt(r, x, y)
	if gotCharacter != character || len(lines) != 10 {
		t.Fatalf("hovered ability portrait = %v with %d lines, want character and 10 lines", gotCharacter, len(lines))
	}
	if !r.characterRosterPanelHovering(x, y) {
		t.Fatal("character portrait with abilities should use the pointer cursor")
	}
}

func TestCharacterRosterMissionButtonMatchesDutyButtonAtBottomRight(t *testing.T) {
	ownedColumn, _ := characterRosterColumns()
	character := &army.Commander{}
	dutyButton := characterRosterDutyActionButton(ownedColumn, 0, 0, character)
	missionButton := characterRosterMissionButton(ownedColumn, 0, 0, []*army.Commander{character})
	row := characterRosterOwnedRowRect(characterRosterOwnedListViewport(ownedColumn), 0, 0, []*army.Commander{character})
	traits := characterRosterTraitRowRect(row)

	if missionButton.Label != "Görev" || missionButton.W != dutyButton.W {
		t.Fatalf("mission button = %+v, want Ata/Kaldır-sized button %+v", missionButton, dutyButton)
	}
	if missionButton.X+missionButton.W != row.X+row.W-characterRosterCardPadding ||
		missionButton.Y != traits.Y || missionButton.H != traits.H ||
		missionButton.Y+missionButton.H != row.Y+row.H-characterRosterCardPadding {
		t.Fatalf("mission button %v is not inset at the card's bottom-right corner: %v", missionButton, row)
	}
}

func TestMilitaryCommanderShowsCouncilAssignmentButNoReleaseControl(t *testing.T) {
	const owner faction.FactionID = "player"
	commander := &army.Commander{
		ID: "cmd_jaime", Type: army.CharacterTypeCommander, OwnerID: string(owner), Name: "Jaime",
		AssignedArmyID: "army_1",
	}
	if !characterRosterCouncilManageable(commander) {
		t.Fatal("military commander should be eligible for council controls")
	}
	if characterRosterReleaseable(commander) {
		t.Fatal("military commander should not receive household release control")
	}
	if !characterRosterReleaseable(&army.Commander{Type: army.CharacterTypeCharacter}) {
		t.Fatal("household character should retain council and release controls")
	}
	r := &Renderer{
		gs: &state.GameState{
			Year:            300,
			PlayerFactionID: owner,
			Factions:        map[faction.FactionID]*faction.Faction{owner: {ID: owner}},
			Commanders:      map[string]*army.Commander{commander.ID: commander},
			CouncilConfig: scenario.CharacterCouncilConfig{
				Roles: []scenario.CouncilRoleDefinition{
					{ID: state.CouncilRoleTreasurer, NameTR: "Hazinedar"},
					{ID: state.CouncilRoleMilitaryResponsible, NameTR: "Serasker", SetCommanderArmy: true},
				},
			},
		},
	}
	r.OpenCharacterRosterPanel()
	if len(r.characterRosterOwned) != 1 || r.characterRosterOwned[0] != commander {
		t.Fatalf("owned household panel members = %#v, want Jaime", r.characterRosterOwned)
	}
	ownedColumn, _ := characterRosterColumns()
	release := characterReleaseButton(ownedColumn, 0, r.characterRosterOwnedScroll, r.characterRosterOwned)
	if r.characterRosterPanelHovering(release.X+release.W/2, release.Y+release.H/2) {
		t.Fatal("military commander exposed the household release control")
	}
	assign := characterRosterDutyActionButton(ownedColumn, 0, r.characterRosterOwnedScroll, commander, r.characterRosterOwned)
	if assign.Label != "Ata" || !r.characterRosterPanelHovering(assign.X+assign.W/2, assign.Y+assign.H/2) {
		t.Fatal("eligible army commander does not show an interactive Ata button")
	}
	r.characterRosterAssignMenuID = commander.ID
	roles := r.gs.CouncilConfig.Roles
	disabledOption := characterRosterAssignmentOptionButtonForRole(ownedColumn, 0, r.characterRosterOwnedScroll, 0, len(roles), roles[0].NameTR, r.characterRosterOwned)
	if r.characterRosterPanelHovering(disabledOption.X+disabledOption.W/2, disabledOption.Y+disabledOption.H/2) {
		t.Fatal("army commander showed an unavailable council role as interactive")
	}
	if action, handled := handleCharacterRosterPanelClick(r, disabledOption.X+disabledOption.W/2, disabledOption.Y+disabledOption.H/2); !handled ||
		action.Kind != ActionNone || r.characterRosterAssignMenuID != commander.ID {
		t.Fatalf("disabled council role click = %+v, handled=%t, menu=%q; want inert option with menu open",
			action, handled, r.characterRosterAssignMenuID)
	}
	option := characterRosterAssignmentOptionButtonForRole(ownedColumn, 0, r.characterRosterOwnedScroll, 1, len(roles), roles[1].NameTR, r.characterRosterOwned)
	if !r.characterRosterPanelHovering(option.X+option.W/2, option.Y+option.H/2) {
		t.Fatal("military commander council role option is not interactive")
	}
}

func TestOwnedCharacterReleaseButtonIsInteractive(t *testing.T) {
	const owner faction.FactionID = "player"
	character := &army.Commander{
		ID: "chr_bronn", Type: army.CharacterTypeCharacter, OwnerID: string(owner), Name: "Bronn",
	}
	r := &Renderer{
		gs: &state.GameState{
			Year:            300,
			PlayerFactionID: owner,
			Factions:        map[faction.FactionID]*faction.Faction{owner: {ID: owner}},
			Commanders:      map[string]*army.Commander{character.ID: character},
		},
	}
	r.OpenCharacterRosterPanel()
	ownedColumn, _ := characterRosterColumns()
	release := characterReleaseButton(ownedColumn, 0, r.characterRosterOwnedScroll)
	if !r.characterRosterPanelHovering(release.X+release.W/2, release.Y+release.H/2) {
		t.Fatal("owned character release button is not interactive")
	}

	character.ActiveMissionID = "recon"
	character.ActiveMissionTurnsRemaining = 2
	r.RefreshCharacterRosterPanel()
	release = characterReleaseButton(ownedColumn, 0, r.characterRosterOwnedScroll, r.characterRosterOwned)
	if characterRosterReleaseable(character) || r.characterRosterPanelHovering(release.X+release.W/2, release.Y+release.H/2) {
		t.Fatal("character with an active mission exposed the household release control")
	}
	if _, handled := handleCharacterRosterPanelClick(r, release.X+release.W/2, release.Y+release.H/2); handled ||
		r.confirmDialog.show {
		t.Fatal("character with an active mission could be released")
	}

	character.ActiveMissionID = ""
	character.ActiveMissionTurnsRemaining = 0
	r.RefreshCharacterRosterPanel()
	if !characterRosterReleaseable(character) || !r.characterRosterPanelHovering(release.X+release.W/2, release.Y+release.H/2) {
		t.Fatal("character release control did not return after the mission ended")
	}
}

func TestCharacterRosterColumnsGiveCandidatesMoreCardWidth(t *testing.T) {
	owned, candidates := characterRosterColumns()
	previousCandidateWidth := (owned.W + candidates.W) * 0.42
	if candidates.W <= previousCandidateWidth {
		t.Fatalf("candidate column width = %.0f, want wider than the previous 42%% allocation %.0f", candidates.W, previousCandidateWidth)
	}
}

func TestCharacterRosterPanelFitsSixFullOwnedRows(t *testing.T) {
	originalWidth, originalHeight := ScreenWidth, ScreenHeight
	defer func() {
		ScreenWidth, ScreenHeight = originalWidth, originalHeight
	}()
	ScreenWidth, ScreenHeight = 1680, 1200

	base := activeWarsPanelRect()
	panel := characterRosterPanelRect()
	owned, _ := characterRosterColumns()
	baseOwnedWidth := (base.W - 2*characterRosterPanelPad - characterRosterColumnGap) * 0.5
	if panel.W-base.W != 40 || baseOwnedWidth+40 != owned.W {
		t.Fatalf("panel/card width increase = %.0f/%.0f, want 40/40", panel.W-base.W, owned.W-baseOwnedWidth)
	}
	if panel.H != characterRosterPanelMaxH {
		t.Fatalf("character roster panel height = %.0f, want max height %.0f", panel.H, characterRosterPanelMaxH)
	}
	viewport := characterRosterOwnedListViewport(owned)
	if viewport.H != 6*characterRosterRowH {
		t.Fatalf("owned roster viewport height = %.0f, want exactly six row pitches (%.0f)", viewport.H, 6*characterRosterRowH)
	}
}

func TestCharacterCouncilEffectLabelsUseOneLinePerBonus(t *testing.T) {
	effects := scenario.CouncilEffects{
		IncomeMod:          1.05,
		SatisfactionPoints: 1,
		MilitaryPowerMod:   1.03,
		CommanderPowerMod:  1.02,
	}
	got := characterCouncilEffectsLinesTR(effects)
	want := []string{"Bölgesel vergi +%5", "Memnuniyet +1", "Askerî güç +%3", "Komuta gücü +%2"}
	if len(got) != len(want) {
		t.Fatalf("council bonus lines = %v, want %v", got, want)
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("council bonus line %d = %q, want %q", index, got[index], want[index])
		}
	}
}

func TestCharacterCouncilEffectLabelsFormatPenaltiesAsNegative(t *testing.T) {
	effects := scenario.CouncilEffects{
		SatisfactionPoints:  -1,
		MilitaryPowerMod:    0.98,
		RelationBonusPoints: -2,
	}
	got := characterCouncilEffectsLinesTR(effects)
	want := []string{"Memnuniyet -1", "Askerî güç -%2", "Heyet -2"}
	if len(got) != len(want) {
		t.Fatalf("council penalty lines = %v, want %v", got, want)
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("council penalty line %d = %q, want %q", index, got[index], want[index])
		}
	}
}

func TestCharacterRosterPanelFitsThreeUnassignedRowsBelowAssignedGroup(t *testing.T) {
	originalHeight := ScreenHeight
	defer func() { ScreenHeight = originalHeight }()
	ScreenHeight = 1200

	panel := characterRosterPanelRect()
	if panel.H <= activeWarsPanelRect().H {
		t.Fatalf("character roster panel height %.0f should exceed active wars panel height %.0f", panel.H, activeWarsPanelRect().H)
	}
	ownedColumn, _ := characterRosterColumns()
	viewport := characterRosterOwnedListViewport(ownedColumn)
	characters := []*army.Commander{
		{AssignedCouncilRole: state.CouncilRoleHeadOfHouse},
		{AssignedCouncilRole: state.CouncilRoleTreasurer},
		{AssignedCouncilRole: state.CouncilRoleDiplomat},
		{},
		{},
		{},
	}
	visibleRows := characterRosterOwnedVisibleRows(viewport, characters, 0)
	unassignedVisible := 0
	for index := 0; index < visibleRows; index++ {
		if characters[index].AssignedCouncilRole == "" {
			unassignedVisible++
		}
	}
	if unassignedVisible < 3 {
		t.Fatalf("visible unassigned rows = %d, want at least 3 (panel=%v viewport=%v)", unassignedVisible, panel, viewport)
	}
}

func TestCharacterDutyActionsAndAssignmentMenuStayInRosterCard(t *testing.T) {
	ownedColumn, _ := characterRosterColumns()
	ownedViewport := characterRosterOwnedListViewport(ownedColumn)
	roles := scenario.DefaultCharacterCouncilConfig().Roles
	if ownedViewport.Y != ownedColumn.Y+34 {
		t.Fatalf("owned roster starts at y=%.0f, want directly below header at %.0f", ownedViewport.Y, ownedColumn.Y+34)
	}
	unassignedButton := characterRosterDutyActionButton(ownedColumn, 0, 0, &army.Commander{})
	assignedButton := characterRosterDutyActionButton(ownedColumn, 0, 0, &army.Commander{AssignedCouncilRole: state.CouncilRoleTreasurer})
	if unassignedButton.Label != "Ata" || unassignedButton.W != characterRosterDutyButtonW {
		t.Fatalf("assign button = %+v, want compact Ata label button", unassignedButton)
	}
	if assignedButton.Label != "Kaldır" || assignedButton.Icon != "" ||
		assignedButton.W != unassignedButton.W || assignedButton.H != unassignedButton.H {
		t.Fatalf("remove-duty button = %+v, want Ata-sized text-only Kaldır button", assignedButton)
	}
	menu := characterRosterAssignmentMenuRectWithCount(ownedColumn, 0, 0, len(roles))
	if menu.Y != unassignedButton.Y+unassignedButton.H+4 {
		t.Fatalf("assignment menu y=%.0f, want immediately below Ata button at %.0f", menu.Y, unassignedButton.Y+unassignedButton.H+4)
	}
	for optionIndex, role := range roles {
		option := characterRosterAssignmentOptionButtonForRole(ownedColumn, 0, 0, optionIndex, len(roles), role.NameTR)
		if option.X < menu.X || option.X+option.W > menu.X+menu.W ||
			option.Y < menu.Y || option.Y+option.H > menu.Y+menu.H {
			t.Fatalf("assignment option %d = %+v, outside menu %v", optionIndex, option, menu)
		}
		if option.W != menu.W-characterRosterAssignMenuPad*2 || option.H != characterRosterAssignOptionH ||
			option.Y != menu.Y+characterRosterAssignMenuPad+float64(optionIndex)*characterRosterAssignOptionH {
			t.Fatalf("assignment option %d = %+v, want padded full-width list row in menu %v", optionIndex, option, menu)
		}
	}
	row := characterRosterRowRect(ownedViewport, 0)
	detail := characterRosterDetailRect(row, row.X+float64(armyPanelCommanderPortrait)+16)
	if detail.Y <= assignedButton.Y+assignedButton.H || detail.X+detail.W > row.X+row.W {
		t.Fatalf("duty detail rect %v should use the full card width below the remove button %v", detail, assignedButton)
	}

	lastOption := characterRosterAssignmentOptionButtonForRole(ownedColumn, 0, 0, len(roles)-1, len(roles), roles[len(roles)-1].NameTR)
	if lastOption.Label != roles[len(roles)-1].NameTR || lastOption.Y+lastOption.H > menu.Y+menu.H {
		t.Fatalf("last council role option = %+v, outside expanded menu %v", lastOption, menu)
	}
}

func TestCharacterTraitBadgesUseLocalizedLabelsAndTraitColors(t *testing.T) {
	for trait, wantLabel := range map[army.CharacterTrait]string{
		"mercenary": "Paralı Asker",
		"scholar":   "Bilgin",
		"assassin":  "Suikastçı",
	} {
		style := characterTraitBadge(trait)
		if style.Label != wantLabel {
			t.Errorf("characterTraitBadge(%q).Label = %q, want %q", trait, style.Label, wantLabel)
		}
		if style.BG == (characterTraitBadgeStyle{}).BG {
			t.Errorf("characterTraitBadge(%q) has no category color", trait)
		}
	}
}

func TestCharacterRosterUpkeepLabelForCouncilCommander(t *testing.T) {
	commander := &army.Commander{
		ID: "cmd_tyrion", Type: army.CharacterTypeCommander, GoldUpkeep: 7,
	}
	paidRole := scenario.CouncilRoleDefinition{ID: state.CouncilRoleTreasurer, NameTR: "Hazinedar"}
	if got, want := characterRosterUpkeepLabel(commander, paidRole), "Bakım: 7 altın/tur"; got != want {
		t.Fatalf("assigned commander's upkeep label = %q, want %q", got, want)
	}
	freeRole := scenario.CouncilRoleDefinition{ID: state.CouncilRoleHeadOfHouse, NameTR: "Lider", NoCost: true}
	if got, want := characterRosterUpkeepLabel(commander, freeRole), "Tur başı bakım yok"; got != want {
		t.Fatalf("no-cost commander's upkeep label = %q, want %q", got, want)
	}
}

func TestCharacterRosterHUDButtonSitsBesideActiveWarsButton(t *testing.T) {
	active := buildActiveWarsHUDButton()
	characters := buildCharacterRosterHUDButton()
	if characters.X-active.X != activeWarsHudButtonStep || characters.Y != active.Y ||
		characters.W != active.W || characters.H != active.H {
		t.Fatalf("character button = %+v, active wars button = %+v", characters, active)
	}
	if !characterRosterHUDButtonHit(characters.X+characters.W/2, characters.Y+characters.H/2) {
		t.Fatal("character roster button center is not clickable")
	}
	if characterRosterHUDButtonHit(characters.X-1, characters.Y+characters.H/2) {
		t.Fatal("point outside character roster button was clickable")
	}
}

func TestCharacterRosterHUDBadgesCountOpenRolesAndActiveMissions(t *testing.T) {
	const player = faction.FactionID("player")
	roles := []scenario.CouncilRoleDefinition{
		{ID: "role_1"}, {ID: "role_2"}, {ID: "role_3"},
		{ID: "role_4"}, {ID: "role_5"}, {ID: "role_6"},
	}
	gs := &state.GameState{
		Year:            300,
		PlayerFactionID: player,
		CouncilConfig:   scenario.CharacterCouncilConfig{Roles: roles},
		Commanders: map[string]*army.Commander{
			"one":   {ID: "one", OwnerID: string(player), Type: army.CharacterTypeCharacter, AssignedCouncilRole: "role_1"},
			"two":   {ID: "two", OwnerID: string(player), Type: army.CharacterTypeCharacter, AssignedCouncilRole: "role_2", ActiveMissionID: "mission_a"},
			"three": {ID: "three", OwnerID: string(player), Type: army.CharacterTypeCommander, AssignedCouncilRole: "role_3"},
			"four":  {ID: "four", OwnerID: string(player), Type: army.CharacterTypeAdministrator, AssignedCouncilRole: "role_4", ActiveMissionID: "mission_b"},
			"five":  {ID: "five", OwnerID: string(player), Type: army.CharacterTypeCharacter},
			"six":   {ID: "six", OwnerID: string(player), Type: army.CharacterTypeCharacter},
			"enemy": {ID: "enemy", OwnerID: "enemy", Type: army.CharacterTypeCharacter, ActiveMissionID: "mission_enemy"},
			"dead":  {ID: "dead", OwnerID: string(player), Type: army.CharacterTypeCharacter, AssignedCouncilRole: "role_5", ActiveMissionID: "mission_dead"},
			"future": {
				ID: "future", OwnerID: string(player), Type: army.CharacterTypeCharacter,
				StartYear: 301, AssignedCouncilRole: "role_6", ActiveMissionID: "mission_future",
			},
		},
		DeadCommanderIDs: map[string]bool{"dead": true},
	}

	openRoles, activeMissions := characterRosterHUDCounts(gs)
	if openRoles != 2 {
		t.Fatalf("open council roles = %d, want 2", openRoles)
	}
	if activeMissions != 2 {
		t.Fatalf("active mission characters = %d, want 2", activeMissions)
	}

	gs.Commanders["five"].AssignedCouncilRole = "role_5"
	gs.Commanders["six"].AssignedCouncilRole = "role_6"
	openRoles, activeMissions = characterRosterHUDCounts(gs)
	if openRoles != 0 {
		t.Fatalf("open council roles after filling every role = %d, want 0", openRoles)
	}
}

func TestCharacterRosterButtonAndPanelUseTopInputLayers(t *testing.T) {
	originalWidth := ScreenWidth
	defer func() { ScreenWidth = originalWidth }()
	ScreenWidth = 1680

	character := &army.Commander{
		ID: "chr_bronn", Type: army.CharacterTypeCharacter,
		OwnerID: state.NeutralCharacterOwnerID, Name: "Bronn",
	}
	r := &Renderer{
		gs: &state.GameState{
			Phase:           state.PhasePlayerTurn,
			PlayerFactionID: "player",
			Factions:        map[faction.FactionID]*faction.Faction{"player": {ID: "player"}},
			Year:            300,
			Commanders:      map[string]*army.Commander{character.ID: character},
		},
		showCharacterRoster: true,
	}
	r.rebuildUILayers()

	button := buildCharacterRosterHUDButton()
	buttonX, buttonY := button.X+button.W/2, button.Y+button.H/2
	layer, ok := r.uiLayers.TopAt(buttonX, buttonY)
	if !ok || layer.ID != uiLayerTopStatus || !r.uiLayerPointerAt(buttonX, buttonY) {
		t.Fatalf("character roster HUD button layer/pointer = %+v, %v", layer, ok)
	}

	panelPoint := gameuiRectCenter(characterRosterPanelRect())
	layer, ok = r.uiLayers.TopAt(panelPoint.X, panelPoint.Y)
	if !ok || layer.ID != uiLayerCharacterRoster {
		t.Fatalf("character roster panel layer = %+v, %v", layer, ok)
	}
	_, candidateColumn := characterRosterColumns()
	join := characterAdoptButton(candidateColumn, 0, 0)
	if !r.characterRosterPanelHovering(join.X+join.W/2, join.Y+join.H/2) {
		t.Fatal("candidate join button does not produce a pointer")
	}
}

func TestCharacterRosterTreasurerSeatAndAssignmentButtonAreInteractive(t *testing.T) {
	const owner faction.FactionID = "player"
	character := &army.Commander{
		ID: "cmd_tywin", Type: army.CharacterTypeAdministrator, OwnerID: string(owner), Name: "Tywin",
	}
	r := &Renderer{
		gs: &state.GameState{
			Year:            300,
			PlayerFactionID: owner,
			Factions:        map[faction.FactionID]*faction.Faction{owner: {ID: owner}},
			Commanders:      map[string]*army.Commander{character.ID: character},
		},
	}
	r.OpenCharacterRosterPanel()
	ownedColumn, _ := characterRosterColumns()
	viewport := characterRosterOwnedListViewport(ownedColumn)
	roles := characterCouncilRoles(r.gs)
	assign := characterCouncilAssignButton(ownedColumn, 0, 0)
	if assign.Label != "Ata" {
		t.Fatalf("assignment button label = %q, want Ata", assign.Label)
	}
	if !r.characterRosterPanelHovering(assign.X+assign.W/2, assign.Y+assign.H/2) {
		t.Fatal("eligible administrator's council assignment button is not interactive")
	}
	r.characterRosterAssignMenuID = character.ID
	for optionIndex, role := range roles {
		option := characterRosterAssignmentOptionButtonForRole(ownedColumn, 0, 0, optionIndex, len(roles), role.NameTR)
		menu := characterRosterAssignmentMenuRectWithCount(ownedColumn, 0, 0, len(roles))
		if option.Label != role.NameTR ||
			option.X < menu.X || option.X+option.W > menu.X+menu.W ||
			option.Y < menu.Y || option.Y+option.H > menu.Y+menu.H {
			t.Fatalf("assignment option %d = %+v, outside menu %v", optionIndex, option, menu)
		}
		if got := characterRosterAssignmentOptionAtForRoles(option.X+option.W/2, option.Y+option.H/2, ownedColumn, 0, 0, roles); got != optionIndex {
			t.Fatalf("assignment option hit = %d, want %d", got, optionIndex)
		}
		hovering := r.characterRosterPanelHovering(option.X+option.W/2, option.Y+option.H/2)
		if enabled := r.gs.CanAssignCharacterToCouncilRole(owner, character.ID, role.ID); hovering != enabled {
			t.Fatalf("assignment option %q hover = %t, want enabled state %t", option.Label, hovering, enabled)
		}
		if !r.gs.CanAssignCharacterToCouncilRole(owner, character.ID, role.ID) {
			continue
		}
		if !hovering {
			t.Fatalf("assignment option %q is not interactive", option.Label)
		}
	}
	if !r.gs.AssignCharacterToCouncilRole(owner, character.ID, state.CouncilRoleTreasurer) {
		t.Fatal("treasurer assignment failed")
	}
	r.characterRosterAssignMenuID = ""
	r.RefreshCharacterRosterPanel()
	removeDuty := characterRosterDutyActionButton(ownedColumn, 0, 0, character)
	if removeDuty.Label != "Kaldır" || removeDuty.Icon != "" ||
		!r.characterRosterPanelHovering(removeDuty.X+removeDuty.W/2, removeDuty.Y+removeDuty.H/2) {
		t.Fatal("assigned treasurer's text-only Kaldır button is not interactive")
	}
	character.ActiveMissionID = "envoy"
	character.ActiveMissionTurnsRemaining = 2
	r.RefreshCharacterRosterPanel()
	removeDuty = characterRosterDutyActionButton(ownedColumn, 0, 0, character, r.characterRosterOwned)
	if characterRosterDutyRemovable(character) || r.characterRosterPanelHovering(removeDuty.X+removeDuty.W/2, removeDuty.Y+removeDuty.H/2) {
		t.Fatal("character on a mission exposed the council Kaldır button")
	}
	if action, handled := handleCharacterRosterPanelClick(r, removeDuty.X+removeDuty.W/2, removeDuty.Y+removeDuty.H/2); handled ||
		action.Kind != ActionNone || character.AssignedCouncilRole != state.CouncilRoleTreasurer {
		t.Fatalf("active mission allowed council role removal: action=%+v handled=%t role=%q",
			action, handled, character.AssignedCouncilRole)
	}
	if r.gs.UnassignCharacterFromCouncilRole(owner, character.ID) ||
		character.AssignedCouncilRole != state.CouncilRoleTreasurer {
		t.Fatal("active mission allowed council role removal through state API")
	}
	character.ActiveMissionID = ""
	character.ActiveMissionTurnsRemaining = 0
	r.RefreshCharacterRosterPanel()
	removeDuty = characterRosterDutyActionButton(ownedColumn, 0, 0, character, r.characterRosterOwned)
	if !characterRosterDutyRemovable(character) ||
		!r.characterRosterPanelHovering(removeDuty.X+removeDuty.W/2, removeDuty.Y+removeDuty.H/2) {
		t.Fatal("council Kaldır button did not return after mission ended")
	}
	release := characterReleaseButton(ownedColumn, 0, 0)
	if r.characterRosterPanelHovering(release.X+release.W/2, release.Y+release.H/2) {
		t.Fatal("assigned treasurer's haneden çıkarma button should be hidden and non-interactive")
	}
	if role, ok := r.gs.CouncilConfig.WithDefaults().Role(character.AssignedCouncilRole); !ok || role.NameTR != "Hazine Sorumlusu" {
		t.Fatalf("assigned character card role = %+v, found=%t", role, ok)
	}
	if got := characterRosterRowAt(removeDuty.X+removeDuty.W/2, removeDuty.Y+removeDuty.H/2, 1, 0, viewport); got != 0 {
		t.Fatalf("treasurer assignment row index = %d, want 0", got)
	}
	if !r.gs.UnassignCharacterFromCouncilRole(owner, character.ID) {
		t.Fatal("treasurer could not be unassigned")
	}
	r.RefreshCharacterRosterPanel()
	if !r.characterRosterPanelHovering(release.X+release.W/2, release.Y+release.H/2) {
		t.Fatal("unassigned treasurer's haneden çıkarma button should be interactive")
	}
	assign = characterRosterDutyActionButton(ownedColumn, 0, 0, character)
	if assign.Label != "Ata" || !r.characterRosterPanelHovering(assign.X+assign.W/2, assign.Y+assign.H/2) {
		t.Fatal("unassigned character's Ata button is not interactive")
	}
	if !r.gs.AssignCharacterToCouncilRole(owner, character.ID, state.CouncilRoleHeadOfHouse) {
		t.Fatal("head of house assignment failed")
	}
	r.RefreshCharacterRosterPanel()
	removeDuty = characterRosterDutyActionButton(ownedColumn, 0, 0, character)
	if removeDuty.Label != "Kaldır" || removeDuty.Icon != "" ||
		!r.characterRosterPanelHovering(removeDuty.X+removeDuty.W/2, removeDuty.Y+removeDuty.H/2) {
		t.Fatal("assigned head of house's text-only Kaldır button is not interactive")
	}
}

func TestCharacterRosterAssignmentMenuUsesScenarioRoles(t *testing.T) {
	const owner faction.FactionID = "player"
	character := &army.Commander{
		ID: "chr_courtier", Type: army.CharacterTypeCharacter, OwnerID: string(owner), Name: "Courtier",
	}
	roles := []scenario.CouncilRoleDefinition{
		{ID: "steward", NameTR: "Kâhya"},
		{ID: "envoy", NameTR: "Elçi"},
	}
	r := &Renderer{
		gs: &state.GameState{
			Year:            300,
			PlayerFactionID: owner,
			Factions:        map[faction.FactionID]*faction.Faction{owner: {ID: owner}},
			Commanders:      map[string]*army.Commander{character.ID: character},
			CouncilConfig:   scenario.CharacterCouncilConfig{Roles: roles},
		},
	}
	r.OpenCharacterRosterPanel()
	ownedColumn, _ := characterRosterColumns()
	assign := characterCouncilAssignButton(ownedColumn, 0, 0, r.characterRosterOwned)
	if _, handled := handleCharacterRosterPanelClick(r, assign.X+assign.W/2, assign.Y+assign.H/2); !handled ||
		r.characterRosterAssignMenuID != character.ID {
		t.Fatal("scenario-configured character assignment menu did not open")
	}
	option := characterRosterAssignmentOptionButtonForRole(ownedColumn, 0, 0, 1, len(roles), roles[1].NameTR, r.characterRosterOwned)
	if option.Label != "Elçi" {
		t.Fatalf("dynamic role option label = %q, want Elçi", option.Label)
	}
	action, handled := handleCharacterRosterPanelClick(r, option.X+option.W/2, option.Y+option.H/2)
	if !handled || action.Kind != ActionAssignCouncilRole || action.CouncilRole != "envoy" {
		t.Fatalf("scenario role selection = %+v, handled=%t", action, handled)
	}
}

func TestCharacterRosterAssignmentMenuCapturesOverflowAboveArmyPanel(t *testing.T) {
	const owner faction.FactionID = "player"
	gs := &state.GameState{
		Year:            300,
		PlayerFactionID: owner,
		Factions:        map[faction.FactionID]*faction.Faction{owner: {ID: owner}},
		Commanders:      make(map[string]*army.Commander),
		Armies:          map[army.ArmyID]*army.Army{"army_player": {ID: "army_player", OwnerID: string(owner)}},
	}
	ownedColumn, _ := characterRosterColumns()
	visibleRows := characterRosterVisibleRows(characterRosterOwnedListViewport(ownedColumn))
	for index := 0; index < visibleRows; index++ {
		id := fmt.Sprintf("chr_overflow_%d", index)
		gs.Commanders[id] = &army.Commander{
			ID: id, Type: army.CharacterTypeCharacter, OwnerID: string(owner), Name: id,
		}
	}
	r := &Renderer{
		gs:                          gs,
		showArmyDetailPanel:         true,
		SelectedArmy:                "army_player",
		showCharacterRoster:         true,
		characterRosterAssignMenuID: fmt.Sprintf("chr_overflow_%d", visibleRows-1),
	}
	r.OpenCharacterRosterPanelForCharacter(r.characterRosterAssignMenuID)
	r.characterRosterAssignMenuID = fmt.Sprintf("chr_overflow_%d", visibleRows-1)
	r.rebuildUILayers()

	ownedColumn, _ = characterRosterColumns()
	roles := characterCouncilRoles(gs)
	characterIndex := characterRosterIndexByID(r.characterRosterOwned, r.characterRosterAssignMenuID)
	menu := characterRosterAssignmentMenuRectWithCount(ownedColumn, characterIndex, r.characterRosterOwnedScroll, len(roles), r.characterRosterOwned)
	panel := characterRosterPanelRect()
	armyPanel, ok := armyDetailPanelRect(gs, "army_player")
	if !ok {
		t.Fatal("army detail panel has no geometry")
	}
	if menu.Y+menu.H <= panel.Y+panel.H {
		t.Fatalf("assignment menu %v does not overflow the roster panel %v", menu, panel)
	}
	characterID := r.characterRosterAssignMenuID
	var selectedRole scenario.CouncilRoleDefinition
	var x, y float64
	for optionIndex, role := range roles {
		option := characterRosterAssignmentOptionButtonForRole(ownedColumn, characterIndex, r.characterRosterOwnedScroll, optionIndex, len(roles), role.NameTR, r.characterRosterOwned)
		optionX, optionY := option.X+option.W/2, option.Y+option.H/2
		if panel.Hit(optionX, optionY) && armyPanel.Hit(optionX, optionY) &&
			gs.CanAssignCharacterToCouncilRole(owner, characterID, role.ID) {
			selectedRole, x, y = role, optionX, optionY
			break
		}
	}
	if selectedRole.ID == "" {
		t.Fatalf("no enabled assignment option overlaps both panels; roster=%v army=%v menu=%v", panel, armyPanel, menu)
	}
	if layer, ok := r.uiLayers.TopAt(x, y); !ok || layer.ID != uiLayerCharacterRoster {
		t.Fatalf("overflowing role option top layer = %+v, %v; want character roster", layer, ok)
	}
	if !characterRosterPanelInteractiveHit(r, x, y) {
		t.Fatal("overlapping assignment option did not receive the pointer hit-test")
	}
	action, handled := handleCharacterRosterPanelClick(r, x, y)
	if !handled || action.Kind != ActionAssignCouncilRole || action.CharacterID != characterID || action.CouncilRole != selectedRole.ID {
		t.Fatalf("overflow role selection = %+v, handled=%t", action, handled)
	}
	overflowX, overflowY := menu.X+menu.W/2, panel.Y+panel.H+1
	if !menu.Hit(overflowX, overflowY) {
		t.Fatalf("menu %v does not cover the expected overflow point (%v, %v)", menu, overflowX, overflowY)
	}
	if layer, ok := r.uiLayers.TopAt(overflowX, overflowY); !ok || layer.ID != uiLayerCharacterRoster {
		t.Fatalf("menu overflow top layer = %+v, %v; want character roster", layer, ok)
	}
}

func TestCharacterMissionMenuDispatchesRoleLinkedAndIndependentMissions(t *testing.T) {
	const owner faction.FactionID = "player"
	const target faction.FactionID = "target"
	diplomat := &army.Commander{
		ID: "chr_diplomat", Type: army.CharacterTypeCharacter, OwnerID: string(owner), Name: "Diplomat",
		AssignedCouncilRole: state.CouncilRoleDiplomat,
	}
	independent := &army.Commander{
		ID: "chr_independent", Type: army.CharacterTypeCharacter, OwnerID: string(owner), Name: "Independent",
	}
	r := &Renderer{
		gs: &state.GameState{
			Year:            300,
			PlayerFactionID: owner,
			Factions: map[faction.FactionID]*faction.Faction{
				owner:  {ID: owner, Gold: 1000},
				target: {ID: target, NameTR: "Hedef Hanesi"},
			},
			Commanders: map[string]*army.Commander{
				diplomat.ID:    diplomat,
				independent.ID: independent,
			},
			Regions: map[world.RegionID]*world.Region{
				"target_region": {ID: "target_region", OwnerID: string(target), NameTR: "Sınır Bölgesi"},
			},
			Armies: map[army.ArmyID]*army.Army{
				"army_player": {ID: "army_player", OwnerID: string(owner)},
			},
			CouncilConfig: scenario.CharacterCouncilConfig{Missions: []scenario.CharacterMissionDefinition{
				{ID: "envoy", NameTR: "Diplomatik Elçilik", RoleID: state.CouncilRoleDiplomat, TargetKind: scenario.CharacterMissionTargetFaction, DurationTurns: 2, GoldCost: 100},
				{ID: "archive", NameTR: "Arşiv Araştırması", DurationTurns: 2, GoldCost: 50},
				{ID: "recon", NameTR: "Bölgesel Keşif", TargetKind: scenario.CharacterMissionTargetRegion, DurationTurns: 2, GoldCost: 75},
			}},
		},
		showArmyDetailPanel: true,
		SelectedArmy:        "army_player",
	}
	r.OpenCharacterRosterPanel()
	ownedColumn, _ := characterRosterColumns()

	diplomatIndex := characterRosterIndexByID(r.characterRosterOwned, diplomat.ID)
	missionButton := characterRosterMissionButton(ownedColumn, diplomatIndex, r.characterRosterOwnedScroll, r.characterRosterOwned)
	if _, handled := handleCharacterRosterPanelClick(r, missionButton.X+missionButton.W/2, missionButton.Y+missionButton.H/2); !handled ||
		r.characterRosterMissionMenuCharacterID != diplomat.ID {
		t.Fatal("clicking the diplomat's Görev button did not open its mission menu")
	}
	menu, ok := characterMissionMenuRect(r)
	if !ok {
		t.Fatal("role-linked mission menu has no geometry")
	}
	missionButton = characterMissionMenuOptionButton(menu, 0, "Diplomatik Elçilik")
	r.rebuildUILayers()
	missionX, missionY := missionButton.X+missionButton.W/2, missionButton.Y+missionButton.H/2
	armyPanel, ok := armyDetailPanelRect(r.gs, r.SelectedArmy)
	if !ok || !armyPanel.Hit(missionX, missionY) {
		t.Fatalf("mission option (%v, %v) does not overlap the open army panel %v", missionX, missionY, armyPanel)
	}
	if layer, ok := r.uiLayers.TopAt(missionX, missionY); !ok || layer.ID != uiLayerCharacterRoster {
		t.Fatalf("overlapping mission option top layer = %+v, %v; want character roster", layer, ok)
	}
	action, handled := handleCharacterRosterPanelClick(r, missionButton.X+missionButton.W/2, missionButton.Y+missionButton.H/2)
	if !handled || action.Kind != ActionNone || r.characterRosterSelectedMissionID != "envoy" ||
		len(r.characterRosterMissionTargets) != 1 || r.characterRosterMissionTargets[0] != string(target) {
		t.Fatalf("role-linked task selection = %+v, handled=%t, selected=%q, targets=%v", action, handled,
			r.characterRosterSelectedMissionID, r.characterRosterMissionTargets)
	}
	menu, _ = characterMissionMenuRect(r)
	targetButton := characterMissionMenuOptionButton(menu, 0, "Hedef Hanesi")
	action, handled = handleCharacterRosterPanelClick(r, targetButton.X+targetButton.W/2, targetButton.Y+targetButton.H/2)
	if !handled || action.Kind != ActionNone || !r.confirmDialog.show ||
		r.confirmDialog.pendingAction.Kind != ActionStartCharacterMission ||
		r.confirmDialog.pendingAction.CharacterID != diplomat.ID ||
		r.confirmDialog.pendingAction.MissionID != "envoy" ||
		r.confirmDialog.pendingAction.TargetFaction != target ||
		r.confirmDialog.acceptLabel != "Evet" || r.confirmDialog.declineLabel != "Hayır" {
		t.Fatalf("role-linked task confirmation = %+v, handled=%t, dialog=%+v", action, handled, r.confirmDialog)
	}
	if !strings.Contains(r.confirmDialog.message, diplomat.Name) ||
		!strings.Contains(r.confirmDialog.message, "Diplomatik Elçilik") ||
		!strings.Contains(r.confirmDialog.message, "Hedef Hanesi") ||
		!strings.Contains(r.confirmDialog.message, "100 altın") {
		t.Fatalf("role-linked task confirmation message omits mission details: %q", r.confirmDialog.message)
	}
	if r.characterRosterMissionMenuCharacterID != "" {
		t.Fatal("mission menu remained open after selecting a target")
	}
	r.confirmDialog = confirmDialogState{}

	independentIndex := characterRosterIndexByID(r.characterRosterOwned, independent.ID)
	missionButton = characterRosterMissionButton(ownedColumn, independentIndex, r.characterRosterOwnedScroll, r.characterRosterOwned)
	if _, handled := handleCharacterRosterPanelClick(r, missionButton.X+missionButton.W/2, missionButton.Y+missionButton.H/2); !handled ||
		r.characterRosterMissionMenuCharacterID != independent.ID {
		t.Fatal("clicking the unassigned character's Görev button did not open its mission menu")
	}
	menu, _ = characterMissionMenuRect(r)
	missionButton = characterMissionMenuOptionButton(menu, 0, "Arşiv Araştırması")
	action, handled = handleCharacterRosterPanelClick(r, missionButton.X+missionButton.W/2, missionButton.Y+missionButton.H/2)
	if !handled || action.Kind != ActionNone || !r.confirmDialog.show ||
		r.confirmDialog.pendingAction.Kind != ActionStartCharacterMission ||
		r.confirmDialog.pendingAction.CharacterID != independent.ID ||
		r.confirmDialog.pendingAction.MissionID != "archive" ||
		r.confirmDialog.pendingAction.TargetFaction != "" || r.confirmDialog.pendingAction.TargetRegion != "" ||
		r.confirmDialog.acceptLabel != "Evet" || r.confirmDialog.declineLabel != "Hayır" {
		t.Fatalf("independent task confirmation = %+v, handled=%t, dialog=%+v", action, handled, r.confirmDialog)
	}
	if !strings.Contains(r.confirmDialog.message, independent.Name) ||
		!strings.Contains(r.confirmDialog.message, "Arşiv Araştırması") ||
		!strings.Contains(r.confirmDialog.message, "2 tur") ||
		!strings.Contains(r.confirmDialog.message, "50 altın") {
		t.Fatalf("independent task confirmation message omits mission details: %q", r.confirmDialog.message)
	}
	if r.characterRosterMissionMenuCharacterID != "" {
		t.Fatal("mission menu remained open after selecting an independent task")
	}
	r.confirmDialog = confirmDialogState{}

	missionButton = characterRosterMissionButton(ownedColumn, independentIndex, r.characterRosterOwnedScroll, r.characterRosterOwned)
	if _, handled := handleCharacterRosterPanelClick(r, missionButton.X+missionButton.W/2, missionButton.Y+missionButton.H/2); !handled {
		t.Fatal("clicking the scout's Görev button did not open its mission menu")
	}
	menu, _ = characterMissionMenuRect(r)
	missionButton = characterMissionMenuOptionButton(menu, 1, "Bölgesel Keşif")
	action, handled = handleCharacterRosterPanelClick(r, missionButton.X+missionButton.W/2, missionButton.Y+missionButton.H/2)
	if !handled || r.characterRosterSelectedMissionID != "recon" {
		t.Fatalf("region-targeted mission selection = %+v, handled=%t", action, handled)
	}
	menu, _ = characterMissionMenuRect(r)
	targetButton = characterMissionMenuOptionButton(menu, 0, "Sınır Bölgesi")
	action, handled = handleCharacterRosterPanelClick(r, targetButton.X+targetButton.W/2, targetButton.Y+targetButton.H/2)
	if !handled || action.Kind != ActionNone || !r.confirmDialog.show ||
		r.confirmDialog.pendingAction.Kind != ActionStartCharacterMission ||
		r.confirmDialog.pendingAction.CharacterID != independent.ID ||
		r.confirmDialog.pendingAction.MissionID != "recon" ||
		r.confirmDialog.pendingAction.TargetRegion != "target_region" ||
		r.confirmDialog.pendingAction.TargetFaction != "" ||
		!strings.Contains(r.confirmDialog.message, "Sınır Bölgesi") {
		t.Fatalf("region-targeted task confirmation = %+v, handled=%t, dialog=%+v", action, handled, r.confirmDialog)
	}
}

func TestCharacterMissionMenuHeaderUsesIconBackButtonAndBalancedPadding(t *testing.T) {
	menu := gameui.Rect{X: 10, Y: 20, W: characterMissionMenuW, H: characterMissionMenuHeaderH + 2*characterMissionMenuRowH + characterMissionMenuPad}
	back := characterMissionMenuBackButton(menu)
	closeButton := characterMissionMenuCloseButton(menu)
	title := characterMissionMenuTitleRect(menu)

	if back.Label != "" || back.Icon != gameui.IconBack {
		t.Fatalf("mission menu back button = %+v, want icon-only back button", back)
	}

	if back.Y-menu.Y != characterMissionMenuPad ||
		characterMissionMenuHeaderH-(back.Y-menu.Y)-back.H != characterMissionMenuPad ||
		closeButton.Y-menu.Y != characterMissionMenuPad ||
		characterMissionMenuHeaderH-(closeButton.Y-menu.Y)-closeButton.H != characterMissionMenuPad {
		t.Fatalf("mission menu header controls lack balanced vertical padding: back=%+v close=%+v header=%.0f pad=%.0f",
			back, closeButton, characterMissionMenuHeaderH, characterMissionMenuPad)
	}
	if title.X < back.X+back.W || title.X+title.W > closeButton.X ||
		title.Y < menu.Y+characterMissionMenuPad || title.Y+title.H > menu.Y+characterMissionMenuHeaderH-characterMissionMenuPad {
		t.Fatalf("mission menu title %v overlaps header controls or padding: back=%+v close=%+v", title, back, closeButton)
	}
	lastOption := characterMissionMenuOptionButton(menu, 1, "Hedef")
	if menu.Y+menu.H-(lastOption.Y+lastOption.H) < characterMissionMenuPad {
		t.Fatalf("mission menu bottom padding is less than %v px", characterMissionMenuPad)
	}
}

func TestCharacterRosterDispatchesAdventureAbilityChoice(t *testing.T) {
	const owner faction.FactionID = "player"
	character := &army.Commander{
		ID: "chr_arya", Type: army.CharacterTypeCharacter, OwnerID: string(owner), Name: "Arya",
		PendingAdventureAbilities: []army.CharacterAbility{
			{ID: "shadowcraft", NameTR: "Gölgede Ustalık"},
			{ID: "fieldcraft", NameTR: "Saha Keşfi"},
		},
	}
	r := &Renderer{gs: &state.GameState{
		Year: 300, PlayerFactionID: owner,
		Factions:   map[faction.FactionID]*faction.Faction{owner: {ID: owner}},
		Commanders: map[string]*army.Commander{character.ID: character},
	}}
	r.OpenCharacterRosterPanel()
	column, _ := characterRosterColumns()
	index := characterRosterIndexByID(r.characterRosterOwned, character.ID)
	button := characterRosterMissionButton(column, index, r.characterRosterOwnedScroll, r.characterRosterOwned)
	if button.Label != "Yetenek" {
		t.Fatalf("pending ability action label = %q, want Yetenek", button.Label)
	}
	if _, handled := handleCharacterRosterPanelClick(r, button.X+button.W/2, button.Y+button.H/2); !handled ||
		!r.characterRosterAbilityChoice {
		t.Fatal("clicking Yetenek did not open the ability selection menu")
	}
	menu, ok := characterMissionMenuRect(r)
	if !ok {
		t.Fatal("ability choice menu has no geometry")
	}
	option := characterMissionMenuOptionButton(menu, 1, "Saha Keşfi")
	action, handled := handleCharacterRosterPanelClick(r, option.X+option.W/2, option.Y+option.H/2)
	if !handled || action.Kind != ActionChooseAdventureAbility || action.CharacterID != character.ID ||
		action.AdventureAbilityID != "fieldcraft" || r.characterRosterMissionMenuCharacterID != "" {
		t.Fatalf("ability selection action = %+v handled=%t; menu character=%q", action, handled, r.characterRosterMissionMenuCharacterID)
	}
}

func TestAdventureMissionConfirmationShowsTraitAdjustedChancesAndRisk(t *testing.T) {
	const owner faction.FactionID = "player"
	character := &army.Commander{
		ID: "chr_arya", Type: army.CharacterTypeCharacter, OwnerID: string(owner), Name: "Arya",
		Traits: []army.CharacterTrait{"assassin"}, CouncilLevel: 2,
	}
	r := &Renderer{gs: &state.GameState{
		Year: 300, PlayerFactionID: owner,
		Factions:   map[faction.FactionID]*faction.Faction{owner: {ID: owner, Gold: 500}},
		Commanders: map[string]*army.Commander{character.ID: character},
		CouncilConfig: scenario.CharacterCouncilConfig{Missions: []scenario.CharacterMissionDefinition{{
			ID: "journey", NameTR: "Macera", Type: scenario.CharacterMissionTypeAdventure,
			TargetKind: scenario.CharacterMissionTargetNone, DurationTurns: 3, GoldCost: 100,
			Adventure: &scenario.CharacterAdventureDefinition{
				SuccessChancePercent: 30, PartialSuccessChancePercent: 20,
				RelevantTraits: []string{"assassin"}, TraitBonusPercent: 10,
				LevelBonusPercent: 5, FailureUnavailableTurns: 2,
			},
		}}},
	}}
	r.OpenCharacterRosterPanel()
	column, _ := characterRosterColumns()
	index := characterRosterIndexByID(r.characterRosterOwned, character.ID)
	button := characterRosterMissionButton(column, index, r.characterRosterOwnedScroll, r.characterRosterOwned)
	if _, handled := handleCharacterRosterPanelClick(r, button.X+button.W/2, button.Y+button.H/2); !handled {
		t.Fatal("adventure mission menu did not open")
	}
	menu, _ := characterMissionMenuRect(r)
	option := characterMissionMenuOptionButton(menu, 0, "Macera")
	if _, handled := handleCharacterRosterPanelClick(r, option.X+option.W/2, option.Y+option.H/2); !handled ||
		!r.confirmDialog.show {
		t.Fatal("adventure mission did not open its confirmation")
	}
	if !strings.Contains(r.confirmDialog.message, "Başarı: %50") ||
		!strings.Contains(r.confirmDialog.message, "kısmi başarı: %20") ||
		!strings.Contains(r.confirmDialog.message, "başarısızlık: %30") ||
		!strings.Contains(r.confirmDialog.message, "2 tur yaralanma riski") {
		t.Fatalf("adventure confirmation omits adjusted chances/risk: %q", r.confirmDialog.message)
	}
}

func TestCharacterRosterClicksUseRenderedButtonGeometry(t *testing.T) {
	const owner faction.FactionID = "player"
	owned := &army.Commander{
		ID: "chr_owned", Type: army.CharacterTypeAdministrator, OwnerID: string(owner), Name: "Owned",
	}
	candidate := &army.Commander{
		ID: "chr_candidate", Type: army.CharacterTypeCharacter, OwnerID: state.NeutralCharacterOwnerID, Name: "Candidate",
	}
	r := &Renderer{
		gs: &state.GameState{
			Year:            300,
			PlayerFactionID: owner,
			Factions:        map[faction.FactionID]*faction.Faction{owner: {ID: owner}},
			Commanders: map[string]*army.Commander{
				owned.ID:     owned,
				candidate.ID: candidate,
			},
		},
	}
	r.OpenCharacterRosterPanel()
	ownedColumn, candidateColumn := characterRosterColumns()
	assign := characterRosterDutyActionButton(ownedColumn, 0, 0, owned)
	if action, handled := handleCharacterRosterPanelClick(r, assign.X+assign.W/2, assign.Y+assign.H/2); !handled || action.Kind != ActionNone {
		t.Fatalf("Ata click = %+v, handled=%t; want menu toggle", action, handled)
	}
	if r.characterRosterAssignMenuID != owned.ID {
		t.Fatalf("assignment menu character = %q, want %q", r.characterRosterAssignMenuID, owned.ID)
	}
	adopt := characterAdoptButton(candidateColumn, 0, 0)
	_, handled := handleCharacterRosterPanelClick(r, adopt.X+adopt.W/2, adopt.Y+adopt.H/2)
	if !handled || !r.confirmDialog.show || r.characterRosterAssignMenuID != "" {
		t.Fatalf("clicking another control with assignment menu open: handled=%t, confirmation=%t, menu=%q",
			handled, r.confirmDialog.show, r.characterRosterAssignMenuID)
	}
	r.confirmDialog.show = false

	if _, handled := handleCharacterRosterPanelClick(r, assign.X+assign.W/2, assign.Y+assign.H/2); !handled {
		t.Fatal("Ata click did not reopen assignment menu")
	}

	roles := characterCouncilRoles(r.gs)
	option := characterRosterAssignmentOptionButtonForRole(ownedColumn, 0, 0, 1, len(roles), roles[1].NameTR)
	action, handled := handleCharacterRosterPanelClick(r, option.X+option.W/2, option.Y+option.H/2)
	if !handled || action.Kind != ActionAssignCouncilRole || action.CharacterID != owned.ID || action.CouncilRole != state.CouncilRoleTreasurer {
		t.Fatalf("assignment option click = %+v, handled=%t", action, handled)
	}

	owned.AssignedCouncilRole = state.CouncilRoleTreasurer
	r.RefreshCharacterRosterPanel()
	removeDuty := characterRosterDutyActionButton(ownedColumn, 0, 0, owned)
	action, handled = handleCharacterRosterPanelClick(r, removeDuty.X+removeDuty.W/2, removeDuty.Y+removeDuty.H/2)
	if !handled || action.Kind != ActionClearCouncilRole || action.CharacterID != owned.ID {
		t.Fatalf("Kaldır click = %+v, handled=%t", action, handled)
	}
	release := characterReleaseButton(ownedColumn, 0, 0)
	_, handled = handleCharacterRosterPanelClick(r, release.X+release.W/2, release.Y+release.H/2)
	if handled || r.confirmDialog.show {
		t.Fatal("assigned character should not be releasable before its council duty is cleared")
	}

	_, handled = handleCharacterRosterPanelClick(r, adopt.X+adopt.W/2, adopt.Y+adopt.H/2)
	if !handled || !r.confirmDialog.show {
		t.Fatalf("candidate join click handled=%t, confirmation visible=%t", handled, r.confirmDialog.show)
	}
}

func TestCharacterRosterAssignmentMenuCanExtendOutsidePanel(t *testing.T) {
	const owner faction.FactionID = "player"
	r := &Renderer{
		gs: &state.GameState{
			Year:            300,
			PlayerFactionID: owner,
			Factions:        map[faction.FactionID]*faction.Faction{owner: {ID: owner}},
			Commanders:      make(map[string]*army.Commander),
		},
	}
	ownedColumn, _ := characterRosterColumns()
	viewport := characterRosterOwnedListViewport(ownedColumn)
	visibleRows := characterRosterVisibleRows(viewport)
	for index := 0; index < visibleRows; index++ {
		id := fmt.Sprintf("chr_%d", index)
		r.gs.Commanders[id] = &army.Commander{
			ID: id, Type: army.CharacterTypeCharacter, OwnerID: string(owner), Name: id,
		}
	}
	r.OpenCharacterRosterPanel()
	if len(r.characterRosterOwned) != visibleRows {
		t.Fatalf("owned roster has %d rows, want %d", len(r.characterRosterOwned), visibleRows)
	}
	lastIndex := len(r.characterRosterOwned) - 1
	character := r.characterRosterOwned[lastIndex]
	r.characterRosterAssignMenuID = character.ID

	roles := characterCouncilRoles(r.gs)
	menu := characterRosterAssignmentMenuRectWithCount(ownedColumn, lastIndex, r.characterRosterOwnedScroll, len(roles))
	if menu.Y+menu.H <= ownedColumn.Y+ownedColumn.H {
		t.Fatalf("menu %v should extend beyond the owned-character panel ending at y=%.0f", menu, ownedColumn.Y+ownedColumn.H)
	}
	option := characterRosterAssignmentOptionButtonForRole(ownedColumn, lastIndex, r.characterRosterOwnedScroll, len(roles)-1, len(roles), roles[len(roles)-1].NameTR)
	action, handled := handleCharacterRosterOutsideClick(r, option.X+option.W/2, option.Y+option.H/2, true, false)
	if !handled || action.Kind != ActionAssignCouncilRole ||
		action.CharacterID != character.ID || action.CouncilRole != roles[len(roles)-1].ID {
		t.Fatalf("outside-panel menu click = %+v, handled=%t", action, handled)
	}
	if len(r.prevMouse) == 0 {
		t.Fatal("handled menu click was not marked as consumed")
	}
}

func TestCharacterRosterOutsideMapClickPassesThrough(t *testing.T) {
	r := &Renderer{showCharacterRoster: true}
	panel := characterRosterPanelRect()
	x, y := panel.X-20, panel.Y+panel.H/2

	if _, handled := handleCharacterRosterOutsideClick(r, x, y, true, false); handled {
		t.Fatal("map click outside the character roster panel was handled by the panel")
	}
	if len(r.prevMouse) != 0 {
		t.Fatal("unhandled map click was consumed before map selection")
	}
}

func gameuiRectCenter(rect gameui.Rect) struct{ X, Y float64 } {
	return struct{ X, Y float64 }{X: rect.X + rect.W/2, Y: rect.Y + rect.H/2}
}
