package render

import (
	"testing"

	"mapp-game-go/internal/army"
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
