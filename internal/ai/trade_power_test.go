package ai

import (
	"testing"

	"mapp-game-go/internal/city"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestAITradePowerBuildingScorePrioritizesLowShareAndOwnedCenter(t *testing.T) {
	gs := &state.GameState{
		Year: 1300,
		Factions: map[faction.FactionID]*faction.Faction{
			"ai":    {ID: "ai"},
			"other": {ID: "other"},
		},
		Regions: map[world.RegionID]*world.Region{
			"center": {ID: "center", OwnerID: "ai", TradeCapacity: 4, Buildings: []string{"market"}},
			"other":  {ID: "other", OwnerID: "other", TradeCapacity: 40},
		},
		BuildingTypes: map[string]*city.Building{
			"market": {ID: "market", TradeCapacityMod: 1.45, MaxPerRegion: 3},
			"port":   {ID: "port", TradeCapacityMod: 1.2, MaxPerRegion: 3},
		},
		TradeCenters: world.TradeCenterConfig{
			Centers: []world.TradeCenterDef{{ID: "center", Tier: world.TradeCenterPrimary}},
		},
	}

	centerScore := aiTradePowerBuildingScore(gs, "ai", gs.Regions["center"], "market")
	otherRegion := &world.Region{ID: "other", OwnerID: "ai", TradeCapacity: 4, Buildings: []string{"market"}}
	otherScore := aiTradePowerBuildingScore(gs, "ai", otherRegion, "market")
	if centerScore <= otherScore {
		t.Fatalf("AI sahibi olduğu ticaret merkezindeki kapasite artışını önceliklendirmeliydi: center=%d other=%d", centerScore, otherScore)
	}
}
