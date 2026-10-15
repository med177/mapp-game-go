package render

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/scenario"
	"mapp-game-go/internal/state"
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
