package state

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/world"
)

func TestSiegeSurrenderTurnsAreShorterByFortLevel(t *testing.T) {
	tests := []struct {
		fortLevel int
		want      int
	}{
		{fortLevel: 1, want: 6},
		{fortLevel: 2, want: 8},
		{fortLevel: 3, want: 10},
	}
	for _, tt := range tests {
		if got := SiegeSurrenderTurns(tt.fortLevel); got != tt.want {
			t.Fatalf("tahkimat %d için teslim süresi %d olmalıydı, got=%d", tt.fortLevel, tt.want, got)
		}
	}
}

func TestEffectiveArmyGrainUpkeepUsesBorderSupplyAndCapitalDistance(t *testing.T) {
	gs := &GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"p1":     {ID: "p1", CapitalSettlementID: "capital_city"},
			"ally":   {ID: "ally", Grain: 30},
			"vassal": {ID: "vassal", OverlordID: "p1", Grain: 30},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("p1", "ally"): {FactionA: "p1", FactionB: "ally", Stance: faction.StanceAllied},
		},
		UnitTypes: map[string]*army.UnitType{
			"inf": {ID: "inf", GrainUpkeep: 10},
		},
		Regions: map[world.RegionID]*world.Region{
			"capital":       {ID: "capital", OwnerID: "p1", Neighbors: []world.RegionID{"frontier", "supply_1"}, Settlements: []world.Settlement{{ID: "capital_city", IsCenter: true}}},
			"frontier":      {ID: "frontier", OwnerID: "p1", Neighbors: []world.RegionID{"capital", "fort"}},
			"fort":          {ID: "fort", OwnerID: "p2", Neighbors: []world.RegionID{"frontier"}},
			"ally_fort":     {ID: "ally_fort", OwnerID: "p2", Neighbors: []world.RegionID{"ally_border"}},
			"ally_border":   {ID: "ally_border", OwnerID: "ally", Neighbors: []world.RegionID{"ally_fort"}},
			"vassal_fort":   {ID: "vassal_fort", OwnerID: "p2", Neighbors: []world.RegionID{"vassal_border"}},
			"vassal_border": {ID: "vassal_border", OwnerID: "vassal", Neighbors: []world.RegionID{"vassal_fort"}},
			"supply_1":      {ID: "supply_1", OwnerID: "p1", Neighbors: []world.RegionID{"capital", "supply_2"}},
			"supply_2":      {ID: "supply_2", OwnerID: "p1", Neighbors: []world.RegionID{"supply_1", "supply_3"}},
			"supply_3":      {ID: "supply_3", OwnerID: "p1", Neighbors: []world.RegionID{"supply_2", "supply_4"}},
			"supply_4":      {ID: "supply_4", OwnerID: "p1", Neighbors: []world.RegionID{"supply_3", "supply_5"}},
			"supply_5":      {ID: "supply_5", OwnerID: "p1", Neighbors: []world.RegionID{"supply_4"}},
			"disconnected":  {ID: "disconnected", OwnerID: "p2"},
		},
		Armies: map[army.ArmyID]*army.Army{
			"besieger":        {ID: "besieger", OwnerID: "p1", RegionID: "fort", Units: []army.Unit{{TypeID: "inf", CurrentHP: 100}}},
			"ally_besieger":   {ID: "ally_besieger", OwnerID: "p1", RegionID: "ally_fort", Units: []army.Unit{{TypeID: "inf", CurrentHP: 100}}},
			"vassal_besieger": {ID: "vassal_besieger", OwnerID: "p1", RegionID: "vassal_fort", Units: []army.Unit{{TypeID: "inf", CurrentHP: 100}}},
			"distant":         {ID: "distant", OwnerID: "p1", RegionID: "supply_5", Units: []army.Unit{{TypeID: "inf", CurrentHP: 100}}},
			"disconnected":    {ID: "disconnected", OwnerID: "p1", RegionID: "disconnected", Units: []army.Unit{{TypeID: "inf", CurrentHP: 100}}},
		},
		Sieges: map[world.RegionID]*SiegeState{
			"fort":        {RegionID: "fort", AttackerArmyID: "besieger"},
			"ally_fort":   {RegionID: "ally_fort", AttackerArmyID: "ally_besieger"},
			"vassal_fort": {RegionID: "vassal_fort", AttackerArmyID: "vassal_besieger"},
		},
	}

	if got := gs.RegionalArmyGrainDemand(gs.Armies["besieger"]); got != 15 {
		t.Fatalf("kendi sınırından ikmal alan kuşatma 150%% bakım kullanmalıydı, got=%d", got)
	}
	for _, id := range []army.ArmyID{"ally_besieger", "vassal_besieger"} {
		if got := gs.RegionalArmyGrainDemand(gs.Armies[id]); got != 16 {
			t.Fatalf("%s dost sınır ikmaliyle düşük kuşatma yükü almalıydı, got=%d", id, got)
		}
	}
	if got := gs.CapitalSupplyPenaltyPercent(gs.Armies["distant"]); got != 20 {
		t.Fatalf("bağlı başkent hattındaki uzak ordu kademeli mesafe cezası almalıydı, got=%d", got)
	}
	if got := gs.RegionalArmyGrainDemand(gs.Armies["disconnected"]); got != 15 {
		t.Fatalf("başkentten kopuk ordu ek ikmal yükü almalıydı, got=%d", got)
	}
}
