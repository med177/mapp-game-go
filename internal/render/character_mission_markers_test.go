package render

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/scenario"
	"mapp-game-go/internal/state"
	gameui "mapp-game-go/internal/ui"
	"mapp-game-go/internal/world"
)

func TestCharacterMissionMapMarkersGroupTargetsAndResolveFactionCapital(t *testing.T) {
	const player faction.FactionID = "player"
	const target faction.FactionID = "target"
	gs := &state.GameState{
		PlayerFactionID: player,
		Factions: map[faction.FactionID]*faction.Faction{
			player: {ID: player},
			target: {ID: target, CapitalSettlementID: "target_capital"},
		},
		Regions: map[world.RegionID]*world.Region{
			"capital":  {ID: "capital", OwnerID: string(target), Settlements: []world.Settlement{{ID: "target_capital"}}},
			"region_a": {ID: "region_a", OwnerID: string(target)},
			"region_b": {ID: "region_b", OwnerID: string(target)},
		},
		Commanders: map[string]*army.Commander{
			"chr_z":         {ID: "chr_z", OwnerID: string(player), Name: "Zed", ActiveMissionID: "recon", ActiveMissionTargetID: "region_b", ActiveMissionTurnsRemaining: 2},
			"chr_a":         {ID: "chr_a", OwnerID: string(player), Name: "Arya", ActiveMissionID: "recon", ActiveMissionTargetID: "region_b", ActiveMissionTurnsRemaining: 1},
			"chr_envoy":     {ID: "chr_envoy", OwnerID: string(player), ActiveMissionID: "envoy", ActiveMissionTargetID: string(target), ActiveMissionTurnsRemaining: 3},
			"chr_adventure": {ID: "chr_adventure", OwnerID: string(player), ActiveMissionID: "adventure", ActiveMissionTurnsRemaining: 3},
			"chr_enemy":     {ID: "chr_enemy", OwnerID: string(target), ActiveMissionID: "recon", ActiveMissionTargetID: "region_a", ActiveMissionTurnsRemaining: 3},
		},
		CouncilConfig: scenario.CharacterCouncilConfig{Missions: []scenario.CharacterMissionDefinition{
			{ID: "recon", TargetKind: scenario.CharacterMissionTargetRegion},
			{ID: "envoy", TargetKind: scenario.CharacterMissionTargetFaction},
			{ID: "adventure", Type: scenario.CharacterMissionTypeAdventure, TargetKind: scenario.CharacterMissionTargetNone},
		}},
	}
	renderer := &Renderer{gs: gs}

	markers := renderer.characterMissionMapMarkers()
	if len(markers) != 2 {
		t.Fatalf("mission map markers = %+v, want one region marker and one faction marker", markers)
	}
	if markers[0].regionID != "capital" || markers[0].count != 1 ||
		markers[0].characterID != "chr_envoy" || markers[0].missionID != "envoy" {
		t.Errorf("faction target marker = %+v, want target faction capital", markers[0])
	}
	if markers[1].regionID != "region_b" || markers[1].count != 2 ||
		markers[1].characterID != "chr_a" || markers[1].turnsRemaining != 1 {
		t.Errorf("grouped region target marker = %+v, want deterministic representative and count 2", markers[1])
	}
}

func TestCharacterMissionMapTargetFactionFallsBackToOwnedRegion(t *testing.T) {
	const target faction.FactionID = "target"
	gs := &state.GameState{
		Factions: map[faction.FactionID]*faction.Faction{target: {ID: target}},
		Regions: map[world.RegionID]*world.Region{
			"z_region": {ID: "z_region", OwnerID: string(target)},
			"a_region": {ID: "a_region", OwnerID: string(target)},
			"sea":      {ID: "sea", OwnerID: string(target), IsSea: true},
			"terrain":  {ID: "terrain", OwnerID: string(target), IsTerrainArea: true},
		},
	}

	mission := scenario.CharacterMissionDefinition{TargetKind: scenario.CharacterMissionTargetFaction}
	if got := characterMissionMapTargetRegion(gs, mission, string(target)); got == nil || got.ID != "a_region" {
		t.Fatalf("faction target fallback = %+v, want first eligible owned region a_region", got)
	}
}

func TestCharacterMissionMapMarkerRectContainsItsCenter(t *testing.T) {
	rect := characterMissionMapMarkerRect(100, 120)
	if rect.W != characterMissionMapMarkerSize || rect.H != characterMissionMapMarkerSize ||
		!rect.Hit(100, 120) || rect.Hit(rect.X-1, 120) {
		t.Fatalf("mission marker geometry = %+v, expected centered hit rectangle", rect)
	}
}

func TestCharacterMissionMapMarkerCentersAfterSettlementLabel(t *testing.T) {
	labelX, labelY, labelW, labelH := 100.0, 150.0, 92.0, 16.0
	markerX, markerY := characterMissionMapMarkerCenterAfterLabel(labelX, labelY, labelW, labelH)
	labelRight := labelX + labelW
	markerLeft := markerX - characterMissionMapMarkerSize/2

	if markerLeft-labelRight != characterMissionMapMarkerLabelGap ||
		markerY != labelY+labelH/2 {
		t.Fatalf("mission marker center = (%.0f, %.0f), label right=%.0f; want %gpx gap and vertically centered",
			markerX, markerY, labelRight, characterMissionMapMarkerLabelGap)
	}
}

func TestCharacterMissionMapMarkerOpensAndHighlightsTargetCharacter(t *testing.T) {
	const player faction.FactionID = "player"
	const selectedID = "chr_14"
	characters := make(map[string]*army.Commander, 15)
	for index := 0; index < 15; index++ {
		id := "chr_" + itoa(index)
		character := &army.Commander{
			ID: id, Type: army.CharacterTypeCharacter, OwnerID: string(player), Name: "Character " + itoa(index),
		}
		characters[id] = character
	}
	characters[selectedID].Name = "Zed"
	characters[selectedID].ActiveMissionID = "recon"
	characters[selectedID].ActiveMissionTargetID = "target_region"
	characters[selectedID].ActiveMissionTurnsRemaining = 2
	r := &Renderer{
		gs: &state.GameState{
			PlayerFactionID: player,
			Commanders:      characters,
			Regions: map[world.RegionID]*world.Region{
				"target_region": {ID: "target_region", IsSea: true},
			},
			CouncilConfig: scenario.CharacterCouncilConfig{Missions: []scenario.CharacterMissionDefinition{{
				ID: "recon", TargetKind: scenario.CharacterMissionTargetRegion,
			}}},
		},
		mapMode:  MapModeNormal,
		camScale: 1,
	}

	markerX, markerY, ok := r.characterMissionMapMarkerCenter("target_region")
	if !ok {
		t.Fatal("mission marker has no screen position")
	}
	if _, ok := r.characterMissionMapMarkerAt(markerX, markerY); !ok {
		t.Fatal("marker center did not hit its shared clickable geometry")
	}
	if !r.openCharacterRosterForMissionMarkerAt(markerX, markerY) {
		t.Fatal("clicking the marker did not open the character roster")
	}
	if !r.showCharacterRoster || r.characterRosterSelectedCharacterID != selectedID {
		t.Fatalf("roster open/selection = %t/%q, want selected character %q", r.showCharacterRoster,
			r.characterRosterSelectedCharacterID, selectedID)
	}
	index := characterRosterIndexByID(r.characterRosterOwned, selectedID)
	if index < 0 {
		t.Fatal("selected mission character is missing from the household roster")
	}
	ownedColumn, _ := characterRosterColumns()
	viewport := characterRosterOwnedListViewport(ownedColumn)
	row := characterRosterOwnedRowRect(viewport, index, r.characterRosterOwnedScroll, r.characterRosterOwned)
	if got := characterRosterOwnedRowAt(row.X+row.W/2, row.Y+row.H/2, r.characterRosterOwned,
		r.characterRosterOwnedScroll, viewport); got != index {
		t.Fatalf("selected character row hit = %d, want visible row %d at scroll %d", got, index, r.characterRosterOwnedScroll)
	}
	r.mapMode = MapModeTrade
	if _, hit := r.characterMissionMapMarkerAt(markerX, markerY); hit {
		t.Fatal("mission marker remained clickable in trade map mode")
	}
}

func TestCharacterMissionMarkerTooltipFitsCommanderPortrait(t *testing.T) {
	for _, cursor := range [][2]float64{{0, 0}, {ScreenWidth / 2, ScreenHeight / 2}, {ScreenWidth, ScreenHeight}} {
		layout := characterMissionMarkerTooltipGeometry(cursor[0], cursor[1])
		if layout.panel.W != characterMissionMarkerTooltipWidth || layout.panel.W >= 400 {
			t.Fatalf("tooltip width = %.0f, want narrowed width %.0f", layout.panel.W, characterMissionMarkerTooltipWidth)
		}
		if layout.portrait.W != float64(armyPanelCommanderPortrait) ||
			layout.portrait.H != float64(armyPanelCommanderPortrait) {
			t.Fatalf("tooltip portrait = %v, want commander panel dimensions", layout.portrait)
		}
		if layout.panel.X > layout.portrait.X || layout.panel.Y > layout.portrait.Y ||
			layout.portrait.X+layout.portrait.W > layout.panel.X+layout.panel.W ||
			layout.portrait.Y+layout.portrait.H > layout.panel.Y+layout.panel.H {
			t.Fatalf("tooltip portrait %v does not fit panel %v", layout.portrait, layout.panel)
		}
		textRows := []gameui.Rect{layout.name, layout.mission, layout.remaining}
		for index, row := range textRows {
			if row.X < layout.portrait.X+layout.portrait.W || row.X+row.W > layout.panel.X+layout.panel.W ||
				row.Y < layout.panel.Y || row.Y+row.H > layout.panel.Y+layout.panel.H {
				t.Fatalf("tooltip text row %d (%v) overlaps portrait or escapes panel %v", index, row, layout.panel)
			}
			if index > 0 && row.Y <= textRows[index-1].Y {
				t.Fatalf("tooltip text rows are not ordered vertically: %+v", textRows)
			}
		}
	}
}
