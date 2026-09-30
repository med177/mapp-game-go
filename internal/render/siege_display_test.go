package render

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestSiegeForArmyDisplayIgnoresStaleDefenderReference(t *testing.T) {
	defender := &army.Army{
		ID:       "epirus_army",
		OwnerID:  "epir",
		RegionID: "yanya",
		Units:    []army.Unit{{TypeID: "militia", CurrentHP: army.MaxUnitHP}},
	}
	attacker := &army.Army{
		ID:       "venice_army",
		OwnerID:  "venice",
		RegionID: "yanya",
		Units:    []army.Unit{{TypeID: "militia", CurrentHP: army.MaxUnitHP}},
	}
	epirusSiege := &state.SiegeState{
		RegionID:       "epirus",
		AttackerArmyID: "venice_epirus_army",
		DefenderArmyID: defender.ID,
	}
	yanyaSiege := &state.SiegeState{
		RegionID:       "yanya",
		AttackerArmyID: attacker.ID,
		DefenderArmyID: defender.ID,
	}
	gs := &state.GameState{
		Regions: map[world.RegionID]*world.Region{
			"epirus": {ID: "epirus", OwnerID: "epir"},
			"yanya":  {ID: "yanya", OwnerID: "epir"},
		},
		Armies: map[army.ArmyID]*army.Army{
			defender.ID:          defender,
			attacker.ID:          attacker,
			"venice_epirus_army": {ID: "venice_epirus_army", OwnerID: "venice", RegionID: "epirus"},
		},
		Factions: map[faction.FactionID]*faction.Faction{
			"epir":   {ID: "epir"},
			"venice": {ID: "venice"},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("venice", "epir"): {
				FactionA: "venice",
				FactionB: "epir",
				Stance:   faction.StanceWar,
			},
		},
		Sieges: map[world.RegionID]*state.SiegeState{
			epirusSiege.RegionID: epirusSiege,
			yanyaSiege.RegionID:  yanyaSiege,
		},
	}
	r := &Renderer{gs: gs}

	for i := 0; i < 20; i++ {
		if got := r.siegeForArmyDisplay(defender); got != yanyaSiege {
			t.Fatalf("geçerli kuşatma yerine eski bağlantı seçildi: got=%#v", got)
		}
	}
}

func TestSiegeSupportMarkersStayBesideTheirSiegeSide(t *testing.T) {
	region := &world.Region{
		ID:      "yanya",
		OwnerID: "epirus",
		Settlements: []world.Settlement{{
			ID:       "yanya_fortress",
			Type:     world.SettlementFortress,
			IsCenter: true,
		}},
	}
	attacker := &army.Army{ID: "venice_army", OwnerID: "venice", RegionID: region.ID}
	attackerSupport := &army.Army{ID: "venice_support", OwnerID: "venice_ally", RegionID: region.ID}
	defender := &army.Army{ID: "epirus_army", OwnerID: "epirus", RegionID: region.ID}
	defenderSupport := &army.Army{ID: "epirus_support", OwnerID: "epirus_ally", RegionID: region.ID}
	siege := &state.SiegeState{
		RegionID:       region.ID,
		AttackerArmyID: attacker.ID,
		DefenderArmyID: defender.ID,
	}
	gs := &state.GameState{
		Phase:   state.PhaseEditMode,
		Regions: map[world.RegionID]*world.Region{region.ID: region},
		Armies: map[army.ArmyID]*army.Army{
			attacker.ID:        attacker,
			attackerSupport.ID: attackerSupport,
			defender.ID:        defender,
			defenderSupport.ID: defenderSupport,
		},
		Factions: map[faction.FactionID]*faction.Faction{
			"venice":      {ID: "venice"},
			"venice_ally": {ID: "venice_ally"},
			"epirus":      {ID: "epirus"},
			"epirus_ally": {ID: "epirus_ally"},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("venice", "venice_ally"): {
				FactionA: "venice", FactionB: "venice_ally", Stance: faction.StanceAllied,
			},
			faction.RelationKey("epirus", "epirus_ally"): {
				FactionA: "epirus", FactionB: "epirus_ally", Stance: faction.StanceAllied,
			},
		},
		Sieges: map[world.RegionID]*state.SiegeState{region.ID: siege},
	}
	r := &Renderer{
		gs:       gs,
		camScale: 1,
		worldMap: &WorldMap{
			settlementAnchor: map[settlementAnchorKey][2]int{{Region: region.ID, Index: 0}: {100, 100}},
			primarySettlement: map[world.RegionID][2]int{
				region.ID: {100, 100},
			},
		},
	}

	if got := r.siegeForArmyDisplay(attackerSupport); got != siege {
		t.Fatalf("kuşatan desteği kuşatma grubuna alınmadı: got=%#v", got)
	}
	if got := r.siegeForArmyDisplay(defenderSupport); got != siege {
		t.Fatalf("kuşatılan desteği kuşatma grubuna alınmadı: got=%#v", got)
	}

	positions := r.armyIconPositions()
	if len(positions) != 4 {
		t.Fatalf("kuşatma marker sayısı = %d, want 4", len(positions))
	}
	wantOrder := []army.ArmyID{attackerSupport.ID, attacker.ID, defender.ID, defenderSupport.ID}
	for i, want := range wantOrder {
		if positions[i].ArmyID != want {
			t.Fatalf("soldan sağa marker %d = %q, want %q", i, positions[i].ArmyID, want)
		}
	}
	for i := 1; i < len(positions); i++ {
		if positions[i-1].X >= positions[i].X {
			t.Fatalf("markerler yatayda sıralı değil: %+v", positions)
		}
	}
}
