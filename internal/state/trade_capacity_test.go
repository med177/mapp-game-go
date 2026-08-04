package state

import (
	"testing"

	"mapp-game-go/internal/city"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/world"
)

func TestEffectiveRegionTradeCapacityUsesBuildingsAndTradeCenter(t *testing.T) {
	region := &world.Region{
		ID:            "center",
		OwnerID:       "owner",
		TradeCapacity: 4,
		Buildings:     []string{"market", "granary", "temple"},
	}
	gs := &GameState{
		Year: 1300,
		Regions: map[world.RegionID]*world.Region{
			region.ID: region,
		},
		BuildingTypes: map[string]*city.Building{
			"market":  {ID: "market", TradeCapacityMod: 1.45},
			"granary": {ID: "granary", TradeCapacityMod: 1.05},
			"temple":  {ID: "temple", TradeCapacityMod: 1.03},
		},
		TradeCenters: world.TradeCenterConfig{
			PrimaryTradeCapacityBonus: 2,
			PrimaryTradeIncomeBonus:   4,
			Centers: []world.TradeCenterDef{{
				ID: region.ID, Tier: world.TradeCenterPrimary,
			}},
		},
	}

	// 4 × 1.45 × 1.05 × 1.03 = 6.27 -> 6; primary center +2 = 8.
	if got := gs.EffectiveRegionTradeCapacity(region); got != 8 {
		t.Fatalf("efektif bölge ticaret kapasitesi 8 olmalı, got=%d", got)
	}
	capacityBonus, incomeBonus := gs.TradeCenterBenefits(region)
	if capacityBonus != 2 || incomeBonus != 4 {
		t.Fatalf("ticaret merkezi faydaları beklenen 2 kapasite / 4 gümrük olmalı, got=%d/%d", capacityBonus, incomeBonus)
	}
	if got := gs.BaseRegionTradeIncome(region); got != 20 {
		t.Fatalf("merkez gümrüğü pasif ticaret gelirine eklenmeli, got=%d", got)
	}
}

func TestEffectiveRegionTradeCapacityIgnoresBuildingsWithoutTradeModifier(t *testing.T) {
	region := &world.Region{
		ID:            "mixed_buildings",
		OwnerID:       "owner",
		TradeCapacity: 4,
		Buildings:     []string{"market", "walls"},
	}
	gs := &GameState{
		BuildingTypes: map[string]*city.Building{
			"market": {ID: "market", TradeCapacityMod: 1.45},
			"walls":  {ID: "walls"},
		},
	}

	if got := gs.EffectiveRegionTradeCapacity(region); got != 6 {
		t.Fatalf("ticaret etkisi olmayan bina kapasiteyi sıfırlamamalı, got=%d", got)
	}
}

func TestTradeCenterBenefitsGrowWithVolumeWithoutUpperLimit(t *testing.T) {
	region := &world.Region{ID: "large_center", OwnerID: "owner", TradeCapacity: 133}
	gs := &GameState{
		Regions: map[world.RegionID]*world.Region{region.ID: region},
		TradeCenters: world.TradeCenterConfig{
			PrimaryTradeCapacityBonus: 2,
			PrimaryTradeIncomeBonus:   4,
			Centers:                   []world.TradeCenterDef{{ID: region.ID, Tier: world.TradeCenterPrimary}},
		},
	}

	if got := gs.TradeCenterVolume(region); got != 133 {
		t.Fatalf("merkez hacmi 133 olmalı: got=%d", got)
	}
	capacity, income := gs.TradeCenterBenefits(region)
	if capacity != 6 || income != 12 {
		t.Fatalf("hacim büyümesi sınırsız kademeli çalışmalı: got=%d/%d want=6/12", capacity, income)
	}
}

func TestEffectiveFactionTradeCapacityFollowsConquest(t *testing.T) {
	region := &world.Region{ID: "port", OwnerID: "old", TradeCapacity: 3, Buildings: []string{"market"}}
	gs := &GameState{
		Regions: map[world.RegionID]*world.Region{region.ID: region},
		BuildingTypes: map[string]*city.Building{
			"market": {ID: "market", TradeCapacityMod: 1.45},
		},
		TradeCenters: world.TradeCenterConfig{
			SecondaryTradeCapacityBonus: 1,
			Centers:                     []world.TradeCenterDef{{ID: region.ID, Tier: world.TradeCenterSecondary}},
		},
	}

	if got := gs.EffectiveFactionTradeCapacity(faction.FactionID("old")); got != 5 {
		t.Fatalf("eski sahip merkez bonusunu almalı, got=%d", got)
	}
	region.OwnerID = "new"
	if got := gs.EffectiveFactionTradeCapacity(faction.FactionID("old")); got != 0 {
		t.Fatalf("fetih sonrası eski sahibi kapasiteyi kaybetmeli, got=%d", got)
	}
	if got := gs.EffectiveFactionTradeCapacity(faction.FactionID("new")); got != 5 {
		t.Fatalf("fetih sonrası yeni sahibi merkez bonusunu almalı, got=%d", got)
	}
}
