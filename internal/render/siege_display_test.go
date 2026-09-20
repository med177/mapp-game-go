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
