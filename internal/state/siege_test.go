package state

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/world"
)

func TestNormalizeSiegeDefenderReferencesClearsOnlyStaleLinks(t *testing.T) {
	defender := &army.Army{
		ID:       "epirus_army",
		OwnerID:  "epir",
		RegionID: "yanya",
		Units:    []army.Unit{{TypeID: "militia", CurrentHP: army.MaxUnitHP}},
	}
	gs := &GameState{
		Regions: map[world.RegionID]*world.Region{
			"epirus": {ID: "epirus", OwnerID: "epir"},
			"yanya":  {ID: "yanya", OwnerID: "epir"},
		},
		Armies: map[army.ArmyID]*army.Army{
			defender.ID:          defender,
			"venice_epirus_army": {ID: "venice_epirus_army", OwnerID: "venice", RegionID: "epirus"},
			"venice_yanya_army":  {ID: "venice_yanya_army", OwnerID: "venice", RegionID: "yanya"},
		},
		Sieges: map[world.RegionID]*SiegeState{
			"epirus": {RegionID: "epirus", AttackerArmyID: "venice_epirus_army", DefenderArmyID: defender.ID},
			"yanya":  {RegionID: "yanya", AttackerArmyID: "venice_yanya_army", DefenderArmyID: defender.ID},
		},
	}

	if got := gs.NormalizeSiegeDefenderReferences(); got != 1 {
		t.Fatalf("yalnızca eski bağlantı temizlenmeli, got=%d", got)
	}
	if got := gs.Sieges["epirus"].DefenderArmyID; got != "" {
		t.Fatalf("Epirus eski savunmacı bağlantısı temizlenmedi: %q", got)
	}
	if got := gs.Sieges["yanya"].DefenderArmyID; got != defender.ID {
		t.Fatalf("Yanya geçerli savunmacı bağlantısı korunmalı: %q", got)
	}
}
