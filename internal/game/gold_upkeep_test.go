package game

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func goldUpkeepFixture(gold, income, upkeep int) *state.GameState {
	return &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", Gold: gold, Grain: 100},
		},
		Regions: map[world.RegionID]*world.Region{
			"home": {
				ID: "home", OwnerID: "player", BaseGoldIncome: income,
				TaxRate: 100, Satisfaction: 50,
			},
		},
		UnitTypes: map[string]*army.UnitType{
			"infantry": {ID: "infantry", GoldUpkeep: upkeep},
		},
		Armies: map[army.ArmyID]*army.Army{
			"army": {ID: "army", OwnerID: "player", RegionID: "home", Units: []army.Unit{{TypeID: "infantry", CurrentHP: army.MaxUnitHP}}},
		},
	}
}

func TestApplyEconomyTickDeductsFixedGoldArmyUpkeep(t *testing.T) {
	gs := goldUpkeepFixture(10, 20, 8)
	report := applyEconomyTick(gs)

	if got, want := gs.Factions["player"].Gold, 22; got != want {
		t.Fatalf("gelirden sonra sabit altın bakımı düşülmedi: got=%d want=%d", got, want)
	}
	status := report.PlayerGoldStatus
	if status.Income != 20 || status.Upkeep != 8 || status.PaidUpkeep != 8 || status.Shortage != 0 {
		t.Fatalf("altın ekonomi raporu hatalı: %+v", status)
	}
}

func TestApplyEconomyTickClampsUnpaidGoldUpkeepAtZero(t *testing.T) {
	gs := goldUpkeepFixture(2, 0, 5)
	report := applyEconomyTick(gs)

	if got := gs.Factions["player"].Gold; got != 0 {
		t.Fatalf("ödenemeyen bakım sonrası hazine sıfırlanmalı: got=%d", got)
	}
	if got, want := report.PlayerGoldStatus.Shortage, 3; got != want {
		t.Fatalf("ödenemeyen bakım miktarı hatalı: got=%d want=%d", got, want)
	}
}

func TestGoldUpkeepShortageCausesAttritionAndDesertion(t *testing.T) {
	gs := goldUpkeepFixture(0, 0, 10)
	gs.Armies["army"].Units = make([]army.Unit, 10)
	for i := range gs.Armies["army"].Units {
		gs.Armies["army"].Units[i] = army.Unit{TypeID: "infantry", CurrentHP: army.MaxUnitHP}
	}

	report := applyEconomyTick(gs)
	status := report.PlayerGoldStatus
	if status.Shortage != 100 {
		t.Fatalf("tam maaş açığı bekleniyordu: got=%d", status.Shortage)
	}
	if status.AttritionHPDamage != 100 {
		t.Fatalf("tam maaş açığında 10 HP/tur yıpranma bekleniyordu: got=%d", status.AttritionHPDamage)
	}
	if status.DesertedUnits != 1 || len(gs.Armies["army"].Units) != 9 {
		t.Fatalf("tam maaş açığında bir asker kaçmalıydı: status=%+v units=%d", status, len(gs.Armies["army"].Units))
	}
	if status.ArmyMoraleDelta != -15 {
		t.Fatalf("tam maaş açığında moral kaybı hatalı: got=%d", status.ArmyMoraleDelta)
	}
}
