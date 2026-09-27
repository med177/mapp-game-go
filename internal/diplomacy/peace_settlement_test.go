package diplomacy

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func newPeaceSettlementTestState() *state.GameState {
	return &state.GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"hre":    {ID: "hre"},
			"naples": {ID: "naples"},
			"venice": {ID: "venice"},
		},
		Regions: map[world.RegionID]*world.Region{
			"hre_land":    {ID: "hre_land", OwnerID: "hre"},
			"naples_land": {ID: "naples_land", OwnerID: "naples"},
			"venice_land": {ID: "venice_land", OwnerID: "venice"},
		},
		Armies: map[army.ArmyID]*army.Army{
			"venice_siege": {ID: "venice_siege", OwnerID: "venice", RegionID: "naples_land"},
		},
	}
}

func TestCanImposeVassalageBlocksThirdPartyBesiegedLoser(t *testing.T) {
	gs := newPeaceSettlementTestState()
	gs.Sieges = map[world.RegionID]*state.SiegeState{
		"naples_land": {
			RegionID:          "naples_land",
			AttackerArmyID:    "venice_siege",
			AttackerFactionID: "venice",
		},
	}

	assessment := PeaceAssessment{WarScore: 70}
	if canImposeVassalage(gs, "hre", "naples", assessment) {
		t.Fatal("aktif üçüncü taraf kuşatması sürerken Napoli HRE vassalı yapılmamalı")
	}
	if !canImposeVassalage(gs, "venice", "naples", assessment) {
		t.Fatal("kuşatmayı yürüten Venedik için savaş sonrası vassallık korunmalı")
	}

	delete(gs.Sieges, "naples_land")
	if !canImposeVassalage(gs, "hre", "naples", assessment) {
		t.Fatal("üçüncü taraf kuşatması yokken uygun vassallık sonucu engellenmemeli")
	}
}
