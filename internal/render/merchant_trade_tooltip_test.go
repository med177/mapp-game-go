package render

import (
	"strings"
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestMerchantTradeBonusTooltipIncludesRouteNames(t *testing.T) {
	route := &economy.TradeRoute{
		FromFactionID: "east_rome",
		ToFactionID:   "genoa",
		Good:          economy.GoodGrain,
		AmountPerTurn: 3,
		GoldPerUnit:   2,
	}
	gs := &state.GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"east_rome": {ID: "east_rome", NameTR: "Doğu Roma İmparatorluğu"},
			"genoa":     {ID: "genoa", NameTR: "Ceneviz"},
		},
		Regions: map[world.RegionID]*world.Region{
			"from_port": {ID: "from_port", OwnerID: "east_rome", Neighbors: []world.RegionID{"sea_from"}, Settlements: []world.Settlement{{ID: "from_harbor", Type: world.SettlementPort}}},
			"to_port":   {ID: "to_port", OwnerID: "genoa", Neighbors: []world.RegionID{"sea_to"}, Settlements: []world.Settlement{{ID: "to_harbor", Type: world.SettlementPort}}},
			"sea_from":  {ID: "sea_from", IsSea: true, Neighbors: []world.RegionID{"from_port", "sea_to"}},
			"sea_to":    {ID: "sea_to", IsSea: true, Neighbors: []world.RegionID{"sea_from", "to_port"}},
		},
		Armies: map[army.ArmyID]*army.Army{
			"merchant": {
				ID:            "merchant",
				OwnerID:       "east_rome",
				RegionID:      "sea_to",
				IsNaval:       true,
				TradeRouteKey: route.AssignmentKey(),
				Units:         []army.Unit{{TypeID: "merchant_ship", CurrentHP: army.MaxUnitHP}},
			},
		},
		TradeRoutes: []*economy.TradeRoute{route},
		UnitTypes:   map[string]*army.UnitType{"merchant_ship": {ID: "merchant_ship", Category: army.CategoryNavalTrade}},
	}

	_, detail, ok := merchantTradeBonusTooltipText(gs, gs.Armies["merchant"])
	if !ok {
		t.Fatal("ticaret rotası bonus popup metni üretilmedi")
	}
	for _, want := range []string{"Doğu Roma İmparatorluğu", "Ceneviz", "Tahıl"} {
		if !strings.Contains(detail, want) {
			t.Errorf("popup ayrıntısı %q içinde %q yok", detail, want)
		}
	}
	if lines := strings.Split(detail, "\n"); len(lines) != 2 {
		t.Fatalf("popup satır sayısı = %d, want 2", len(lines))
	}
}
