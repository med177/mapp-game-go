package game

import (
	"strings"
	"testing"

	"mapp-game-go/internal/army"
)

func TestBattleReportSideCarriesCommanderEffectSummaries(t *testing.T) {
	cmd := &army.Commander{
		ID:            "cmd_1",
		Name:          "Osman Bey",
		PortraitAsset: "ottoman_osman.png",
		Level:         4,
		Traits:        []army.CommanderTrait{army.CommanderTraitVeteran, army.CommanderTraitTactician, army.CommanderTraitAggressor},
	}
	a := &army.Army{
		Commander: cmd,
		Units: []army.Unit{
			{TypeID: "inf", CurrentHP: 100},
			{TypeID: "inf", CurrentHP: 90},
		},
	}
	unitTypes := map[string]*army.UnitType{
		"inf": {ID: "inf", Category: army.CategoryInfantry, Attack: 12, Defense: 10, Morale: 50},
	}

	before := snapshotBattleArmy(a, unitTypes)
	after := before
	side := buildBattleReportSide("Saldıran", "Osmanlı", before, after)

	if side.CommanderName != "Osman Bey" {
		t.Fatalf("komutan adı battle report side'a taşınmalıydı, got=%q", side.CommanderName)
	}
	if side.CommanderPortraitAsset != "ottoman_osman.png" {
		t.Fatalf("komutan portresi battle report side'a taşınmalıydı, got=%q", side.CommanderPortraitAsset)
	}
	for _, want := range []string{"Saldırı +12%", "Savunma +4%", "Moral +8%"} {
		if !strings.Contains(side.CommanderBattleEffects, want) {
			t.Fatalf("muharebe özeti %q içinde %q yok", side.CommanderBattleEffects, want)
		}
	}
	for _, want := range []string{"Hareket +1", "Kuşatma +1/+1"} {
		if !strings.Contains(side.CommanderOperationalEffects, want) {
			t.Fatalf("operasyon özeti %q içinde %q yok", side.CommanderOperationalEffects, want)
		}
	}
}
