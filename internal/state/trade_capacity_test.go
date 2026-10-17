package state

import (
	"testing"

	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/world"
)

func TestTradeCenterBenefitsReuseVolumeGrowthWithoutChangingIncome(t *testing.T) {
	region := &world.Region{ID: "center", OwnerID: "owner", TradeCapacity: 50}
	gs := &GameState{
		Regions: map[world.RegionID]*world.Region{region.ID: region},
		TradeRoutes: []*economy.TradeRoute{{
			FromFactionID: "owner", ToFactionID: "partner", AmountPerTurn: 26,
		}},
		TradeCenters: world.TradeCenterConfig{
			PrimaryTradeCapacityBonus: 2,
			PrimaryTradeIncomeBonus:   4,
			Centers: []world.TradeCenterDef{{
				ID: region.ID, Tier: world.TradeCenterPrimary,
			}},
		},
	}

	if got, want := gs.TradeCenterVolume(region), 76; got != want {
		t.Fatalf("trade center volume = %d, want %d", got, want)
	}
	if capacity, income := gs.TradeCenterBenefits(region); capacity != 4 || income != 8 {
		t.Fatalf("trade center benefits = (%d, %d), want (4, 8)", capacity, income)
	}
	if got, want := gs.EffectiveRegionTradeCapacity(region), 54; got != want {
		t.Fatalf("effective trade capacity = %d, want %d", got, want)
	}
	if got, want := gs.BaseRegionTradeIncome(region), 116; got != want {
		t.Fatalf("base region trade income = %d, want %d", got, want)
	}
	if got, want := gs.RegionTradeCenterIncome(region), 16; got != want {
		t.Fatalf("region trade center income = %d, want %d", got, want)
	}

	region.TradeCapacity = 0
	if got := gs.EffectiveRegionTradeCapacity(region); got != 0 {
		t.Fatalf("zero-capacity region trade capacity = %d, want 0", got)
	}
	if got, want := gs.BaseRegionTradeIncome(region), 4; got != want {
		t.Fatalf("zero-capacity region trade income = %d, want center income bonus %d", got, want)
	}
}
