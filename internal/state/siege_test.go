package state

import (
	"math"
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
