package render

import (
	"reflect"
	"testing"

	"mapp-game-go/internal/army"
)

func TestArmyPanelUnitIndexGroupsTypesInDefinitionOrder(t *testing.T) {
	units := []army.Unit{
		{TypeID: "cavalry"},
		{TypeID: "infantry"},
		{TypeID: "cavalry"},
		{TypeID: "archer"},
		{TypeID: "infantry"},
	}
	unitTypeOrder := []string{"infantry", "cavalry", "archer"}

	got := make([]int, len(units))
	for displayIndex := range got {
		got[displayIndex] = armyPanelUnitIndex(units, unitTypeOrder, displayIndex)
	}
	want := []int{1, 4, 0, 2, 3}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("panel birim sırası = %v, want %v", got, want)
	}

	if got := armyPanelUnitIndex(units, unitTypeOrder, len(units)); got != -1 {
		t.Fatalf("taşan panel index'i = %d, want -1", got)
	}
}

func TestArmyPanelUnitIndexGroupsUnlistedTypesAfterDefinitions(t *testing.T) {
	units := []army.Unit{
		{TypeID: "unknown-b"},
		{TypeID: "known"},
		{TypeID: "unknown-a"},
		{TypeID: "unknown-b"},
	}

	got := make([]int, len(units))
	for displayIndex := range got {
		got[displayIndex] = armyPanelUnitIndex(units, []string{"known"}, displayIndex)
	}
	want := []int{1, 0, 3, 2}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("tanımsız tiplerle panel birim sırası = %v, want %v", got, want)
	}
}
