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

func TestCalculateIncludesVassalTributePressure(t *testing.T) {
	gs := &state.GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"lord":   {ID: "lord"},
			"vassal": {ID: "vassal", Grain: 100, OverlordID: "lord", TributeRate: 50, TributeRateConfigured: true},
		},
		Regions: map[world.RegionID]*world.Region{
			"vassal-land": {ID: "vassal-land", OwnerID: "vassal", TaxRate: 30},
		},
	}

	breakdown := Calculate(gs, gs.Regions["vassal-land"])
	if breakdown.Tribute != -6 || breakdown.Total != -6 {
		t.Fatalf("%%50 haraç vassal sadakatine -6 yansıtmalıydı: %+v", breakdown)
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

func TestHistoricalTransitionPressureApproachesStartAndEndDates(t *testing.T) {
	gs := &state.GameState{
		Year: 1395,
		Factions: map[faction.FactionID]*faction.Faction{
			"successor": {ID: "successor", HistoricalStartYear: 1400, HistoricalEndYear: 1405},
		},
	}
	region := &world.Region{ID: "core", OwnerID: "old_owner", SuccessorFactionID: "successor"}

	if got := HistoricalTransitionDelta(gs, region); got != -2 {
		t.Fatalf("kuruluş tarihine yaklaşırken beklenen baskı -2, got=%d", got)
	}
	gs.Year = 1402
	if got := HistoricalTransitionDelta(gs, region); got != -3 {
		t.Fatalf("yıkılış tarihine yaklaşırken beklenen baskı -3, got=%d", got)
	}
	gs.Year = 1405
	if got := HistoricalTransitionDelta(gs, region); got != -4 {
		t.Fatalf("yıkılış tarihinde beklenen baskı -4, got=%d", got)
	}
}

func TestHistoricalTransitionPressureIsIncludedInBreakdown(t *testing.T) {
	gs := &state.GameState{
		Year: 1405,
		Factions: map[faction.FactionID]*faction.Faction{
			"successor": {ID: "successor", Grain: 100, HistoricalEndYear: 1405},
		},
		Regions: map[world.RegionID]*world.Region{
			"core": {ID: "core", OwnerID: "successor", SuccessorFactionID: "successor", TaxRate: 30},
		},
	}
	breakdown := Calculate(gs, gs.Regions["core"])
	if breakdown.Historical != -4 {
		t.Fatalf("tarihsel çözülme baskısı breakdown'a eklenmedi: %+v", breakdown)
	}
}
