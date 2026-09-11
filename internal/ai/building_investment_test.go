package ai

import (
	"testing"

	"mapp-game-go/internal/city"
	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestAIScoreBuildingInvestmentPrefersSatisfactionBonusInDecliningRegion(t *testing.T) {
	gs := &state.GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"ai": {ID: "ai", Grain: 100},
		},
		Regions: map[world.RegionID]*world.Region{
			"region": {
				ID: "region", OwnerID: "ai", Satisfaction: 40, TaxRate: 50,
			},
		},
	}
	region := gs.Regions["region"]
	snapshot := aiBuildEconomySnapshot(gs, "ai")
	signals := aiRegionInvestmentSignals{}

	stabilityBuilding := aiScoreBuildingInvestment(
		gs, gs.Factions["ai"], region,
		&city.Building{ID: "temple", SatBonus: 5},
		economy.ResourceCost{}, 1, 0, 0, snapshot, signals,
	)
	productionBuilding := aiScoreBuildingInvestment(
		gs, gs.Factions["ai"], region,
		&city.Building{ID: "market", SatBonus: 0},
		economy.ResourceCost{}, 1, 0, 0, snapshot, signals,
	)

	if stabilityBuilding.StabilityScore <= productionBuilding.StabilityScore {
		t.Fatalf("memnuniyet bonuslu bina düşen memnuniyette önceliklenmeli: bonuslu=%d, diğer=%d", stabilityBuilding.StabilityScore, productionBuilding.StabilityScore)
	}
}
