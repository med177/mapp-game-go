package render

import (
	"testing"

	"mapp-game-go/internal/army"
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

func TestTradeFleetMarkerZoomVisibility(t *testing.T) {
	fleet := &army.Army{IsNaval: true, TradeRouteKey: "route"}
	normal := &Renderer{mapMode: MapModeNormal, camScale: merchantFleetMarkerZoomScale - 0.01}
	if normal.tradeFleetMarkerVisibleAtCurrentZoom(fleet) {
		t.Fatal("normal haritada düşük zoom ticaret filosu marker'ını gösterdi")
	}

	normal.camScale = merchantFleetMarkerZoomScale
	if !normal.tradeFleetMarkerVisibleAtCurrentZoom(fleet) {
		t.Fatal("eşik zoom'da ticaret filosu marker'ı gizlendi")
	}

	tradeMap := &Renderer{mapMode: MapModeTrade, camScale: 0}
	if !tradeMap.tradeFleetMarkerVisibleAtCurrentZoom(fleet) {
		t.Fatal("ticaret haritasında ticaret filosu marker'ı filtrelendi")
	}

	editMode := &Renderer{mapMode: MapModeNormal, camScale: 0, gs: &state.GameState{Phase: state.PhaseEditMode}}
	if !editMode.tradeFleetMarkerVisibleAtCurrentZoom(fleet) {
		t.Fatal("Edit Mode'da ticaret filosu marker'ı filtrelendi")
	}

	nonTradeFleet := &army.Army{IsNaval: true}
	if normal.tradeFleetMarkerVisibleAtCurrentZoom(nonTradeFleet) != true {
		t.Fatal("rotasız filo marker'ı düşük zoom'da filtrelendi")
	}
}
