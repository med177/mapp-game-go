package game

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
)

func TestDisbandArmyRefundsTwentyPercentAndKeepsUnselectedUnits(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", Gold: 10, Grain: 10, Iron: 10, Timber: 10, Stone: 10, Spice: 10, Cloth: 10},
		},
		UnitTypes: map[string]*army.UnitType{
			"expensive": {ID: "expensive", GoldCost: 101, GrainCost: 25, IronCost: 5},
			"cheap":     {ID: "cheap", GoldCost: 10},
		},
		Armies: map[army.ArmyID]*army.Army{
			"army": {ID: "army", OwnerID: "player", Units: []army.Unit{{TypeID: "expensive"}, {TypeID: "cheap"}, {TypeID: "expensive"}}},
		},
	}

	(&Game{gs: gs}).disbandArmy("army", 0, 2)

	if got := len(gs.Armies["army"].Units); got != 1 || gs.Armies["army"].Units[0].TypeID != "cheap" {
		t.Fatalf("seçilmeyen birim korunmalıydı: %+v", gs.Armies["army"].Units)
	}
	player := gs.Factions["player"]
	if player.Gold != 50 || player.Grain != 20 || player.Iron != 12 {
		t.Fatalf("%s kaynak iadesi yanlış: %+v", "%%20", player)
	}
}

func TestDisbandArmyRemovesEmptyArmyAndLogistics(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions:        map[faction.FactionID]*faction.Faction{"player": {ID: "player"}},
		UnitTypes:       map[string]*army.UnitType{"unit": {ID: "unit", GoldCost: 20}},
		Armies:          map[army.ArmyID]*army.Army{"army": {ID: "army", OwnerID: "player", Units: []army.Unit{{TypeID: "unit"}}}},
		ArmyLogistics:   map[army.ArmyID]state.ArmyLogisticsStatus{"army": {}},
	}

	(&Game{gs: gs}).disbandArmy("army", 0)

	if _, ok := gs.Armies["army"]; ok {
		t.Fatal("son birim terhis edilince boş ordu kalmamalıydı")
	}
	if _, ok := gs.ArmyLogistics["army"]; ok {
		t.Fatal("silinen ordunun lojistik kaydı kalmamalıydı")
	}
	if gs.Factions["player"].Gold != 4 {
		t.Fatalf("son birim için %%20 altın iadesi yanlış: %d", gs.Factions["player"].Gold)
	}
}
