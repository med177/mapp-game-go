package render

import (
	"strings"
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	gameui "mapp-game-go/internal/ui"
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

func TestCharacterRosterViewportsReserveScrollbarGutter(t *testing.T) {
	column := gameui.Rect{X: 20, Y: 30, W: 240, H: 400}
	for name, viewport := range map[string]gameui.Rect{
		"owned":      characterRosterOwnedListViewport(column),
		"candidates": characterRosterListViewport(column),
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
	current := characterRosterRowRect(viewport, 1)
	dividerY, showDivider := characterRosterDutyGroupDividerY(rows, 1, current)
	if !showDivider {
		t.Fatal("assigned-to-unassigned transition should show a divider")
	}
	previous := characterRosterRowRect(viewport, 0)
	if dividerY <= previous.Y+previous.H || dividerY >= current.Y {
		t.Fatalf("group divider y=%.0f should fit in the gap between assigned row %v and unassigned row %v", dividerY, previous, current)
	}
	if _, showDivider := characterRosterDutyGroupDividerY([]*army.Commander{unassigned, unassigned}, 1, current); showDivider {
		t.Fatal("unassigned-to-unassigned rows should not show a divider")
	}
	if assigned.AssignedCouncilRole == "" || unassigned.AssignedCouncilRole != "" {
		t.Fatal("test rows do not represent assigned and unassigned council duties")
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
	character := &army.Commander{ID: "chr_bronn", Name: "Bronn"}
	r.confirmCharacterAdoption(character)
	if !r.confirmDialog.show || r.confirmDialog.pendingAction.Kind != ActionAdoptNeutralCharacter ||
		r.confirmDialog.pendingAction.CharacterID != character.ID ||
		r.confirmDialog.acceptLabel != "Evet" || r.confirmDialog.declineLabel != "Hayır" {
		t.Fatalf("adoption confirmation = %+v", r.confirmDialog)
	}
	if !strings.Contains(r.confirmDialog.message, character.Name) {
		t.Fatalf("adoption confirmation message %q omits character name", r.confirmDialog.message)
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
}

func TestCharacterRosterColumnsGiveCandidatesMoreCardWidth(t *testing.T) {
	owned, candidates := characterRosterColumns()
	previousCandidateWidth := (owned.W + candidates.W) * 0.42
	if candidates.W <= previousCandidateWidth {
		t.Fatalf("candidate column width = %.0f, want wider than the previous 42%% allocation %.0f", candidates.W, previousCandidateWidth)
	}
}

func TestCharacterDutyActionsAndAssignmentMenuStayInRosterCard(t *testing.T) {
	ownedColumn, _ := characterRosterColumns()
	ownedViewport := characterRosterOwnedListViewport(ownedColumn)
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
	menu := characterRosterAssignmentMenuRect(ownedColumn, 0, 0)
	if menu.Y != unassignedButton.Y+unassignedButton.H+4 {
		t.Fatalf("assignment menu y=%.0f, want immediately below Ata button at %.0f", menu.Y, unassignedButton.Y+unassignedButton.H+4)
	}
	for optionIndex := range characterCouncilRoleIDs {
		option := characterRosterAssignmentOptionButton(ownedColumn, 0, 0, optionIndex)
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

	wantRoleLabels := []string{"Hane Lideri", "Hazine Sorumlusu", "Diplomat", "Veliaht", "Askeri Sorumlu"}
	if len(characterCouncilRoleIDs) != len(wantRoleLabels) {
		t.Fatalf("council role count = %d, want %d", len(characterCouncilRoleIDs), len(wantRoleLabels))
	}
	for index, wantLabel := range wantRoleLabels {
		if got := characterCouncilRoleNameTR(characterCouncilRoleIDs[index]); got != wantLabel {
			t.Errorf("council role %q label = %q, want %q", characterCouncilRoleIDs[index], got, wantLabel)
		}
	}
	lastOption := characterRosterAssignmentOptionButton(ownedColumn, 0, 0, len(characterCouncilRoleIDs)-1)
	if lastOption.Label != "Askeri Sorumlu" || lastOption.Y+lastOption.H > menu.Y+menu.H {
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
	assign := characterCouncilAssignButton(ownedColumn, 0, 0)
	if assign.Label != "Ata" {
		t.Fatalf("assignment button label = %q, want Ata", assign.Label)
	}
	if !r.characterRosterPanelHovering(assign.X+assign.W/2, assign.Y+assign.H/2) {
		t.Fatal("eligible administrator's council assignment button is not interactive")
	}
	r.characterRosterAssignMenuID = character.ID
	for optionIndex, role := range characterCouncilRoleIDs {
		option := characterRosterAssignmentOptionButton(ownedColumn, 0, 0, optionIndex)
		menu := characterRosterAssignmentMenuRect(ownedColumn, 0, 0)
		if option.Label != characterCouncilRoleNameTR(role) ||
			option.X < menu.X || option.X+option.W > menu.X+menu.W ||
			option.Y < menu.Y || option.Y+option.H > menu.Y+menu.H {
			t.Fatalf("assignment option %d = %+v, outside menu %v", optionIndex, option, menu)
		}
		if got := characterRosterAssignmentOptionAt(option.X+option.W/2, option.Y+option.H/2, ownedColumn, 0, 0); got != optionIndex {
			t.Fatalf("assignment option hit = %d, want %d", got, optionIndex)
		}
		if !r.characterRosterPanelHovering(option.X+option.W/2, option.Y+option.H/2) {
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
	release := characterReleaseButton(ownedColumn, 0, 0)
	if r.characterRosterPanelHovering(release.X+release.W/2, release.Y+release.H/2) {
		t.Fatal("assigned treasurer's haneden çıkarma button should be hidden and non-interactive")
	}
	if got := characterCouncilRoleNameTR(character.AssignedCouncilRole); got != "Hazine Sorumlusu" ||
		characterCouncilRoleBenefitTR(character.AssignedCouncilRole) != "Vergi geliri +%5" {
		t.Fatalf("assigned character card role/benefit = %q / %q", got, characterCouncilRoleBenefitTR(character.AssignedCouncilRole))
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

	option := characterRosterAssignmentOptionButton(ownedColumn, 0, 0, 1)
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

func gameuiRectCenter(rect gameui.Rect) struct{ X, Y float64 } {
	return struct{ X, Y float64 }{X: rect.X + rect.W/2, Y: rect.Y + rect.H/2}
}
