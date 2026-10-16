package state

import (
	"math"
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/world"
)

func TestNormalizeSiegeReferencesRemovesStaleSiegesAndKeepsValidOnes(t *testing.T) {
	gs := &GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"attacker": {ID: "attacker"},
			"defender": {ID: "defender"},
			"owner":    {ID: "owner"},
			"peaceful": {ID: "peaceful"},
		},
		Regions: map[world.RegionID]*world.Region{
			"valid":       {ID: "valid", OwnerID: "defender", Neighbors: []world.RegionID{"home"}},
			"home":        {ID: "home", OwnerID: "attacker", Neighbors: []world.RegionID{"valid"}},
			"captured":    {ID: "captured", OwnerID: "attacker"},
			"transferred": {ID: "transferred", OwnerID: "owner"},
			"peace":       {ID: "peace", OwnerID: "peaceful"},
			"no_army":     {ID: "no_army", OwnerID: "defender"},
		},
		Armies: map[army.ArmyID]*army.Army{
			"active":               {ID: "active", OwnerID: "attacker", RegionID: "home", Units: []army.Unit{{TypeID: "unit"}}},
			"captured_attacker":    {ID: "captured_attacker", OwnerID: "attacker", RegionID: "captured", Units: []army.Unit{{TypeID: "unit"}}},
			"transferred_attacker": {ID: "transferred_attacker", OwnerID: "attacker", RegionID: "transferred", Units: []army.Unit{{TypeID: "unit"}}},
			"peace_attacker":       {ID: "peace_attacker", OwnerID: "attacker", RegionID: "peace", Units: []army.Unit{{TypeID: "unit"}}},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("attacker", "defender"): {
				FactionA: "attacker", FactionB: "defender", Stance: faction.StanceWar,
			},
			faction.RelationKey("attacker", "owner"): {
				FactionA: "attacker", FactionB: "owner", Stance: faction.StanceWar,
			},
		},
		Sieges: map[world.RegionID]*SiegeState{
			"valid":       {RegionID: "valid", AttackerArmyID: "active", AttackerFactionID: "attacker"},
			"captured":    {RegionID: "captured", AttackerArmyID: "captured_attacker", AttackerFactionID: "attacker"},
			"transferred": {RegionID: "transferred", AttackerArmyID: "transferred_attacker", AttackerFactionID: "defender"},
			"peace":       {RegionID: "peace", AttackerArmyID: "peace_attacker", AttackerFactionID: "attacker"},
			"no_army":     {RegionID: "no_army", AttackerArmyID: "removed_army", AttackerFactionID: "attacker"},
		},
	}

	if got := gs.NormalizeSiegeReferences(); got != 4 {
		t.Fatalf("temizlenen geçersiz kuşatma sayısı = %d, want 4", got)
	}
	if gs.SiegeAt("valid") == nil {
		t.Fatal("geçerli komşu bölge kuşatması temizlendi")
	}
	for _, regionID := range []world.RegionID{"captured", "transferred", "peace", "no_army"} {
		if gs.SiegeAt(regionID) != nil {
			t.Fatalf("%s için geçersiz kuşatma kaydı kaldı", regionID)
		}
	}
}

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

func TestSiegeDefensePressureUsesCompatibleUnitsAndCurrentHP(t *testing.T) {
	types := map[string]*army.UnitType{
		"bombard": {
			ID:                      "bombard",
			Category:                army.CategorySiege,
			Tier:                    2,
			SiegeBreachMaxFortLevel: 6,
			SiegeDefensePressure:    4,
		},
		"catapult": {
			ID:                   "catapult",
			Category:             army.CategorySiege,
			Tier:                 1,
			SiegeDefensePressure: 8,
		},
	}
	attacker := &army.Army{
		Units: []army.Unit{
			{TypeID: "bombard", CurrentHP: army.MaxUnitHP},
			{TypeID: "bombard", CurrentHP: army.MaxUnitHP / 2},
			{TypeID: "catapult", CurrentHP: army.MaxUnitHP},
		},
	}

	if got := SiegeDefensePressureForArmy(types, attacker, 6); got != 0.06 {
		t.Fatalf("T6 baskısı = %v, 0.06 bekleniyordu", got)
	}
	if got := SiegeDefensePressureForArmy(types, attacker, 7); got != 0 {
		t.Fatalf("T7 baskısı = %v, uyumsuz kuşatma birimleri katkı vermemeli", got)
	}
}

func TestSiegeDefenseBonusAppliesPressureWithCap(t *testing.T) {
	withoutPressure := SiegeDefenseBonus(6, 0, 0)
	withPressure := SiegeDefenseBonus(6, 0, 0.04)
	if withoutPressure != 1.02 {
		t.Fatalf("baskısız T6 savunma bonusu = %v, 1.02 bekleniyordu", withoutPressure)
	}
	if withPressure != 0.98 {
		t.Fatalf("%%4 baskılı T6 savunma bonusu = %v, 0.98 bekleniyordu", withPressure)
	}
	if got := SiegeDefenseBonus(6, 2, 0.50); math.Abs(got-0.01) > 1e-9 {
		t.Fatalf("baskı kapaklı büyük gedik bonusu = %v, 0.01 bekleniyordu", got)
	}
}
