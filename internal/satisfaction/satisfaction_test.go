package satisfaction

import (
	"testing"

	"mapp-game-go/internal/city"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestCalculateUsesSharedTaxBuildingAndWarComponents(t *testing.T) {
	gs := &state.GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", Grain: 100},
			"enemy":  {ID: "enemy", Grain: 100},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("player", "enemy"): {
				FactionA: "player", FactionB: "enemy", Stance: faction.StanceWar,
			},
		},
		Regions: map[world.RegionID]*world.Region{
			"home": {
				ID: "home", OwnerID: "player", TaxRate: 50,
				Buildings: []string{"temple", "barracks"},
			},
		},
		BuildingTypes: map[string]*city.Building{
			"temple":   {ID: "temple", SatBonus: 5},
			"barracks": {ID: "barracks", SatBonus: -2},
		},
	}

	breakdown := Calculate(gs, gs.Regions["home"])
	if !breakdown.Valid {
		t.Fatal("geçerli bölge için memnuniyet breakdown'ı üretilmeliydi")
	}
	if breakdown.Tax != -20 || breakdown.Buildings != 3 || breakdown.WarFatigue != -3 {
		t.Fatalf("ortak memnuniyet bileşenleri yanlış: %+v", breakdown)
	}
	if breakdown.Total != -20 {
		t.Fatalf("toplam memnuniyet deltası -20 olmalıydı, got=%d (%+v)", breakdown.Total, breakdown)
	}
}

func TestCalculateShowsZeroGrainPenaltyWhenFactionStockpileIsEmpty(t *testing.T) {
	gs := &state.GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", Grain: 0},
		},
		Regions: map[world.RegionID]*world.Region{
			"home": {ID: "home", OwnerID: "player", TaxRate: 30},
		},
	}

	breakdown := Calculate(gs, gs.Regions["home"])
	if breakdown.Grain != -5 || breakdown.Total != -5 {
		t.Fatalf("boş tahıl stokunda -5 tahıl etkisi gösterilmeliydi: %+v", breakdown)
	}
}
