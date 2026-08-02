package state

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/world"
)

func blockadeEconomyTestState() *GameState {
	return &GameState{
		Month: 6,
		Factions: map[faction.FactionID]*faction.Faction{
			"target": {ID: "target"},
			"raider": {ID: "raider"},
		},
		Regions: map[world.RegionID]*world.Region{
			"port": {
				ID: "port", OwnerID: "target", Terrain: world.TerrainPlain,
				Neighbors:      []world.RegionID{"sea"},
				Settlements:    []world.Settlement{{ID: "port-town", Type: world.SettlementPort}},
				BaseGoldIncome: 100, TaxRate: 100, Satisfaction: 100,
				BaseGrainOutput: 100, BaseIronOutput: 20, BaseTimberOutput: 30,
				BaseStoneOutput: 40, BaseSpiceOutput: 50, BaseClothOutput: 60,
				TradeCapacity: 10,
			},
			"sea": {ID: "sea", IsSea: true, Neighbors: []world.RegionID{"port"}},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("raider", "target"): {
				FactionA: "raider", FactionB: "target", Stance: faction.StanceWar,
			},
		},
		UnitTypes: map[string]*army.UnitType{
			"warship": {ID: "warship", Category: army.CategoryNavalWar},
		},
		Armies: map[army.ArmyID]*army.Army{
			"raider-fleet": {
				ID: "raider-fleet", OwnerID: "raider", RegionID: "sea", IsNaval: true,
				NavalMission: &army.NavalMission{Kind: army.NavalMissionBlockade, TargetRegionID: "sea"},
				Units:        []army.Unit{{TypeID: "warship", CurrentHP: army.MaxUnitHP}},
			},
		},
	}
}

func TestRegionBlockadeEconomicEffectUsesApprovedRetentionAndLootRates(t *testing.T) {
	gs := blockadeEconomyTestState()
	region := gs.Regions["port"]

	effect := gs.RegionBlockadeEconomicEffect(region)
	if effect.BlockadePercent != 50 || effect.OutputRetentionPercent != 75 || effect.BlockaderLootPercent != 5 {
		t.Fatalf("tek gemi oranları yanlış: %+v", effect)
	}

	gs.Armies["raider-fleet"].Units = append(gs.Armies["raider-fleet"].Units, army.Unit{TypeID: "warship", CurrentHP: army.MaxUnitHP})
	effect = gs.RegionBlockadeEconomicEffect(region)
	if effect.BlockadePercent != 100 || effect.OutputRetentionPercent != 50 || effect.BlockaderLootPercent != 10 {
		t.Fatalf("iki gemi oranları yanlış: %+v", effect)
	}
}

func TestRegionProductionAndBlockadeLootFollowRetentionRates(t *testing.T) {
	gs := blockadeEconomyTestState()
	region := gs.Regions["port"]
	base := gs.UnblockedRegionProductionSummary(region)

	blocked := gs.RegionProductionSummary(region)
	for name, values := range map[string][2]int{
		"gold":   {blocked.Gold, scaleBlockadeOutput(base.Gold, 75)},
		"grain":  {blocked.Grain, scaleBlockadeOutput(base.Grain, 75)},
		"iron":   {blocked.Iron, scaleBlockadeOutput(base.Iron, 75)},
		"timber": {blocked.Timber, scaleBlockadeOutput(base.Timber, 75)},
		"stone":  {blocked.Stone, scaleBlockadeOutput(base.Stone, 75)},
		"spice":  {blocked.Spice, scaleBlockadeOutput(base.Spice, 75)},
		"cloth":  {blocked.Cloth, scaleBlockadeOutput(base.Cloth, 75)},
	} {
		if values[0] != values[1] {
			t.Errorf("%%50 abluka %s çıktısı %d değil %d olmalıydı", name, values[0], values[1])
		}
	}

	loot := gs.BlockadeLootForFaction("raider")
	if loot.Gold != scaleBlockadeOutput(base.Gold, 5) || loot.Grain != scaleBlockadeOutput(base.Grain, 5) {
		t.Fatalf("tek geminin loot'u yanlış: loot=%+v base=%+v", loot, base)
	}
	if got, want := gs.BlockadeLootGoldForFleet(gs.Armies["raider-fleet"]), scaleBlockadeOutput(base.Gold, 5); got != want {
		t.Fatalf("tek abluka filosunun altın loot'u %d değil %d olmalıydı", got, want)
	}

	gs.Armies["raider-fleet"].Units = append(gs.Armies["raider-fleet"].Units, army.Unit{TypeID: "warship", CurrentHP: army.MaxUnitHP})
	blocked = gs.RegionProductionSummary(region)
	if blocked.Gold != scaleBlockadeOutput(base.Gold, 50) || blocked.Grain != scaleBlockadeOutput(base.Grain, 50) {
		t.Fatalf("%%100 abluka çıktısı yarıya inmeli: blocked=%+v base=%+v", blocked, base)
	}
	loot = gs.BlockadeLootForFaction("raider")
	if loot.Gold != scaleBlockadeOutput(base.Gold, 10) || loot.Grain != scaleBlockadeOutput(base.Grain, 10) {
		t.Fatalf("iki geminin loot'u yanlış: loot=%+v base=%+v", loot, base)
	}
	if got, want := gs.BlockadeLootGoldForFleet(gs.Armies["raider-fleet"]), scaleBlockadeOutput(base.Gold, 10); got != want {
		t.Fatalf("iki gemili abluka filosunun altın loot'u %d değil %d olmalıydı", got, want)
	}
}
