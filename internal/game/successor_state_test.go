package game

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/render"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestLiberateSuccessorRevivesFactionWithMilitiaAndAlliance(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "liberator",
		Factions: map[faction.FactionID]*faction.Faction{
			"liberator": {ID: "liberator", NameTR: "Özgürleştirici"},
			"successor": {ID: "successor", NameTR: "Ardıl", IsEliminated: true, Gold: 0, Grain: 0},
		},
		Regions: map[world.RegionID]*world.Region{
			"former_capital": {
				ID:                 "former_capital",
				OwnerID:            "liberator",
				SuccessorFactionID: "successor",
				NameTR:             "Eski Başkent",
				TradeCapacity:      3,
				WorldX:             120,
				WorldY:             100,
			},
			"liberator_home": {ID: "liberator_home", OwnerID: "liberator", TradeCapacity: 4, WorldX: 100, WorldY: 100},
		},
		Armies: map[army.ArmyID]*army.Army{
			"invader": {ID: "invader", OwnerID: "liberator", RegionID: "former_capital", Units: army.MakeUnits("militia", 3)},
		},
		Relations: map[string]*faction.Relation{},
		UnitTypes: map[string]*army.UnitType{"militia": {ID: "militia", MovementPoints: 2}},
		TradeCenters: world.TradeCenterConfig{Centers: []world.TradeCenterDef{{
			ID: "liberator_home", Tier: world.TradeCenterPrimary,
		}}},
		NextArmySeq: 4,
	}

	g := &Game{gs: gs, renderer: &render.Renderer{}}
	g.liberateSuccessor("former_capital")

	if got := gs.Regions["former_capital"].OwnerID; got != "successor" {
		t.Fatalf("özgürleştirilen bölge ardıl devlete geçmedi: %q", got)
	}
	if gs.Factions["successor"].IsEliminated {
		t.Fatal("ardıl devlet elenmiş kalmamalıydı")
	}
	if got := gs.Factions["successor"].Gold; got != liberatedFactionGold {
		t.Fatalf("düşük başlangıç altını atanmadı: %d", got)
	}
	if got := gs.Relations[faction.RelationKey("liberator", "successor")]; got == nil || got.Stance != faction.StanceAllied {
		t.Fatalf("özgürleştiriciyle ittifak kurulmadı: %+v", got)
	}

	var revivedArmy *army.Army
	for _, current := range gs.Armies {
		if current != nil && current.OwnerID == "successor" {
			revivedArmy = current
			break
		}
	}
	if revivedArmy == nil || revivedArmy.RegionID != "former_capital" || len(revivedArmy.Units) != liberatedMilitiaCount {
		t.Fatalf("5 milislik diriliş ordusu oluşmadı: %+v", revivedArmy)
	}
	if gs.Armies["invader"].RegionID == "former_capital" {
		t.Fatal("özgürleştirici ordusu yeni ardıl toprağında bırakılmamalıydı")
	}
	for _, center := range gs.TradeCenters.Centers {
		if center.ID == "former_capital" && center.NetworkOnly && len(center.Links) == 1 && center.Links[0] == "liberator_home" {
			return
		}
	}
	t.Fatalf("yeniden kurulan ardıl devlet ticaret ağına bağlanmadı: %+v", gs.TradeCenters.Centers)
}
