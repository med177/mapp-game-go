package render

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
)

func TestArmyIconGridUsesFiveColumnsAndAddsRowsBelow(t *testing.T) {
	base := [2]float32{100, 200}
	want := [][2]float32{
		{36, 200}, {68, 200}, {100, 200}, {132, 200}, {164, 200},
		{36, 232}, {68, 232}, {100, 232}, {132, 232}, {164, 232},
		{84, 264}, {116, 264},
	}
	for index, expected := range want {
		x, y := armyIconGridPosition(base, index, len(want), armyIconStep)
		if x != expected[0] || y != expected[1] {
			t.Fatalf("ikon %d konumu = (%v, %v), want (%v, %v)", index, x, y, expected[0], expected[1])
		}
	}
}

func TestCommanderMarkersUsePriorityAndWiderSpacing(t *testing.T) {
	withoutCommander := &army.Army{ID: "without"}
	withCommander := &army.Army{ID: "with", Commander: &army.Commander{ID: "commander"}}
	if armyHasDisplayedCommander(withoutCommander) {
		t.Fatal("komutansız ordu komutanlı görünmemeli")
	}
	if !armyHasDisplayedCommander(withCommander) {
		t.Fatal("komutanlı ordu komutanlı görünmeli")
	}
	if armyCommanderIconStep <= armyIconStep {
		t.Fatalf("komutanlı marker adımı = %v, normal adım = %v", armyCommanderIconStep, armyIconStep)
	}
	left, _ := armyIconGridPosition([2]float32{100, 200}, 0, 2, armyCommanderIconStep)
	right, _ := armyIconGridPosition([2]float32{100, 200}, 1, 2, armyCommanderIconStep)
	if got := right - left; got != armyCommanderIconStep {
		t.Fatalf("komutanlı marker aralığı = %v, want %v", got, armyCommanderIconStep)
	}
}

func TestMixedCommanderMarkersUseNarrowSpacingForNonCommanders(t *testing.T) {
	steps := []float32{armyCommanderIconStep, armyIconStep, armyIconStep}
	positions := make([][2]float32, len(steps))
	for index := range steps {
		positions[index][0], positions[index][1] = armyIconGridPositionVariable(
			[2]float32{100, 200}, index, len(steps), func(i int) float32 { return steps[i] },
		)
	}
	if got := positions[1][0] - positions[0][0]; got != armyCommanderIconStep {
		t.Fatalf("komutanlı-komutansız marker aralığı = %v, want %v", got, armyCommanderIconStep)
	}
	if got := positions[2][0] - positions[1][0]; got != armyIconStep {
		t.Fatalf("komutansız marker aralığı = %v, want %v", got, armyIconStep)
	}
}

func TestArmyMarkerZoomVisibilityUsesPlayerDiplomacy(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player":  {ID: "player"},
			"vassal":  {ID: "vassal", OverlordID: "player"},
			"ally":    {ID: "ally"},
			"enemy":   {ID: "enemy"},
			"neutral": {ID: "neutral"},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("player", "ally"):  {FactionA: "player", FactionB: "ally", Stance: faction.StanceAllied},
			faction.RelationKey("player", "enemy"): {FactionA: "player", FactionB: "enemy", Stance: faction.StanceWar},
		},
	}
	r := &Renderer{gs: gs, mapMode: MapModeNormal, camScale: armyAllFactionsZoomScale - 0.01}
	for _, ownerID := range []string{"player", "vassal", "ally", "enemy"} {
		if !r.armyVisibleAtCurrentZoom(&army.Army{OwnerID: ownerID}) {
			t.Fatalf("uzak görünümde ilişkili %s marker'ı gizlendi", ownerID)
		}
	}
	if r.armyVisibleAtCurrentZoom(&army.Army{OwnerID: "neutral"}) {
		t.Fatal("uzak görünümde tarafsız devlet marker'ı gösterildi")
	}

	r.camScale = armyAllFactionsZoomScale
	if !r.armyVisibleAtCurrentZoom(&army.Army{OwnerID: "neutral"}) {
		t.Fatal("yakın görünümde tarafsız devlet marker'ı gizlendi")
	}

	tradeFleet := &army.Army{OwnerID: "player", IsNaval: true, TradeRouteKey: "route"}
	if !r.armyVisibleAtCurrentZoom(tradeFleet) {
		t.Fatal("yakın görünümde ticaret filosu marker'ı gizlendi")
	}

	tradeMap := &Renderer{mapMode: MapModeTrade, camScale: 0}
	if !tradeMap.armyVisibleAtCurrentZoom(tradeFleet) {
		t.Fatal("ticaret haritasında filo marker'ı filtrelendi")
	}

	editMode := &Renderer{mapMode: MapModeNormal, camScale: 0, gs: &state.GameState{Phase: state.PhaseEditMode}}
	if !editMode.armyVisibleAtCurrentZoom(&army.Army{OwnerID: "neutral"}) {
		t.Fatal("Edit Mode'da devlet marker'ı filtrelendi")
	}
}
