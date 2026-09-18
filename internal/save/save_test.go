package save

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"mapp-game-go/internal/ai"
	"mapp-game-go/internal/diplomacy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/victory"
	"mapp-game-go/internal/world"
)

func TestLoadScenarioBaseStateLoadsTerrainAreas(t *testing.T) {
	gs, err := loadScenarioBaseState("1300_ottoman_rise", filepath.Join("..", "..", "assets", "scenarios", "1300_ottoman_rise"))
	if err != nil {
		t.Fatalf("loadScenarioBaseState() error = %v", err)
	}
	if len(gs.TerrainAreas) == 0 {
		t.Fatal("scenario terrain areas were not loaded")
	}
	for _, area := range gs.TerrainAreas {
		if gs.Regions[world.TerrainAreaRegionID(area.ID)] == nil {
			t.Fatalf("runtime terrain region for %q was not created", area.ID)
		}
	}
}

func Test1300TradeCenterRelationsSeedAIMerchantAssignments(t *testing.T) {
	gs, err := loadScenarioBaseState("1300_ottoman_rise", filepath.Join("..", "..", "assets", "scenarios", "1300_ottoman_rise"))
	if err != nil {
		t.Fatalf("loadScenarioBaseState() error = %v", err)
	}
	diplomacy.EnsureTradeRoutesForActiveRelations(gs)
	if got, want := gs.Regions["venice"].TradeCapacity, 16; got != want {
		t.Fatalf("Venedik merkez kapasitesi = %d, want %d", got, want)
	}
	if got, want := diplomacy.TradePartnerLimit(gs, faction.FactionID("genoa")), 8; got < want {
		t.Fatalf("Ceneviz merkez partner limiti = %d, want at least %d", got, want)
	}

	for _, fid := range []faction.FactionID{"venice", "genoa"} {
		foundRoute := false
		for _, route := range gs.TradeRoutes {
			if route != nil && route.FromFactionID == string(fid) && route.SuspendedTurns <= 0 {
				foundRoute = true
				break
			}
		}
		if !foundRoute {
			t.Fatalf("%s için başlangıç ticaret merkezi bağlantılı rota oluşturulmadı", fid)
		}

		ai.TakeTurn(gs, fid)
		assigned := false
		for _, fleet := range gs.Armies {
			if fleet == nil || fleet.OwnerID != string(fid) || !fleet.IsNaval || fleet.TradeRouteKey == "" {
				continue
			}
			for _, unit := range fleet.Units {
				if unit.TypeID == "merchant_ship" {
					assigned = true
					break
				}
			}
			if assigned {
				break
			}
		}
		if !assigned {
			t.Fatalf("%s AI merchant filosuna başlangıç ticaret rotası atamadı", fid)
		}
	}
}

func Test1300MaritimeStatesStartWithTradeWeightedNavalComposition(t *testing.T) {
	gs, err := loadScenarioBaseState("1300_ottoman_rise", filepath.Join("..", "..", "assets", "scenarios", "1300_ottoman_rise"))
	if err != nil {
		t.Fatalf("loadScenarioBaseState() error = %v", err)
	}

	type expectedComposition struct {
		warships  int
		merchants int
		deployed  int
	}
	expected := map[faction.FactionID]expectedComposition{
		"aragon":            {warships: 4, merchants: 2, deployed: 7},
		"aydin_bey":         {warships: 1, merchants: 1, deployed: 2},
		"candar_bey":        {warships: 1, merchants: 1, deployed: 2},
		"castile_kingdom":   {warships: 3, merchants: 2, deployed: 7},
		"cyprus_kingdom":    {warships: 2, merchants: 2, deployed: 4},
		"denmark_kingdom":   {warships: 1, merchants: 2, deployed: 3},
		"east_rome":         {warships: 1, merchants: 2, deployed: 8},
		"england":           {warships: 2, merchants: 3, deployed: 6},
		"flanders_county":   {warships: 1, merchants: 2, deployed: 3},
		"france":            {warships: 2, merchants: 2, deployed: 5},
		"florence_rep":      {warships: 0, merchants: 2, deployed: 2},
		"genoa":             {warships: 8, merchants: 6, deployed: 16},
		"granada_emirate":   {warships: 1, merchants: 2, deployed: 3},
		"hafsid_sultanate":  {warships: 1, merchants: 3, deployed: 4},
		"karesioglu_bey":    {warships: 2, merchants: 1, deployed: 3},
		"marinid_sultanate": {warships: 2, merchants: 3, deployed: 5},
		"mamluk":            {warships: 1, merchants: 3, deployed: 10},
		"mecca_sharifate":   {warships: 0, merchants: 1, deployed: 1},
		"mentese_bey":       {warships: 1, merchants: 1, deployed: 2},
		"naples_kingdom":    {warships: 2, merchants: 3, deployed: 6},
		"novgorod_rep":      {warships: 0, merchants: 1, deployed: 1},
		"portugal":          {warships: 2, merchants: 3, deployed: 6},
		"saruhan_bey":       {warships: 0, merchants: 1, deployed: 1},
		"trebizond_emp":     {warships: 1, merchants: 2, deployed: 3},
		"usfurid_emirate":   {warships: 0, merchants: 1, deployed: 1},
		"venice":            {warships: 16, merchants: 6, deployed: 24},
	}
	for fid, want := range expected {
		warships := 0
		merchants := 0
		for _, fleet := range gs.Armies {
			if fleet == nil || fleet.OwnerID != string(fid) || !fleet.IsNaval {
				continue
			}
			for _, unit := range fleet.Units {
				switch unit.TypeID {
				case "warship":
					warships++
				case "merchant_ship":
					merchants++
				}
			}
		}
		if warships != want.warships || merchants != want.merchants {
			t.Fatalf("%s başlangıç donanması = %d savaş, %d ticaret; want %d savaş, %d ticaret", fid, warships, merchants, want.warships, want.merchants)
		}
		if got := gs.DeployedNavalUnits(fid); got != want.deployed {
			t.Fatalf("%s başlangıç toplam gemi = %d, want %d", fid, got, want.deployed)
		}
		if got := gs.NavalCap(fid); got < want.deployed {
			t.Fatalf("%s başlangıç gemileri kapasiteyi aşıyor: %d/%d", fid, want.deployed, got)
		}
	}
}

func Test1300MaritimeMerchantFleetsHaveTradeRoutesForAI(t *testing.T) {
	gs, err := loadScenarioBaseState("1300_ottoman_rise", filepath.Join("..", "..", "assets", "scenarios", "1300_ottoman_rise"))
	if err != nil {
		t.Fatalf("loadScenarioBaseState() error = %v", err)
	}
	diplomacy.EnsureTradeRoutesForActiveRelations(gs)

	merchantStates := []faction.FactionID{
		"aragon", "aydin_bey", "candar_bey", "castile_kingdom", "cyprus_kingdom", "denmark_kingdom",
		"east_rome", "england", "flanders_county", "florence_rep", "genoa",
		"france", "granada_emirate", "hafsid_sultanate", "hormuz_sultanate", "karesioglu_bey",
		"marinid_sultanate", "mamluk", "mecca_sharifate", "mentese_bey", "naples_kingdom",
		"novgorod_rep", "portugal", "saruhan_bey", "trebizond_emp", "usfurid_emirate", "venice",
	}
	for _, fid := range merchantStates {
		hasRoute := false
		for _, route := range gs.TradeRoutes {
			if route == nil || route.FromFactionID != string(fid) || route.SuspendedTurns > 0 {
				continue
			}
			if len(gs.MerchantTradeRouteSeaRegions(route)) > 0 {
				hasRoute = true
				break
			}
		}
		if !hasRoute {
			t.Fatalf("%s merchant filosu için başlangıçta kullanılabilir deniz ticaret rotası yok", fid)
		}

		ai.TakeTurn(gs, fid)
		assigned := false
		for _, fleet := range gs.Armies {
			if fleet == nil || fleet.OwnerID != string(fid) || !fleet.IsNaval || fleet.TradeRouteKey == "" {
				continue
			}
			for _, unit := range fleet.Units {
				if unit.TypeID == "merchant_ship" {
					assigned = true
					break
				}
			}
			if assigned {
				break
			}
		}
		if !assigned {
			t.Fatalf("%s AI merchant filosunu başlangıç ticaret rotasına atamadı", fid)
		}
	}
}

func Test1300OpeningEconomyCoversUpkeepAfterMerchantAssignments(t *testing.T) {
	gs, err := loadScenarioBaseState("1300_ottoman_rise", filepath.Join("..", "..", "assets", "scenarios", "1300_ottoman_rise"))
	if err != nil {
		t.Fatalf("loadScenarioBaseState() error = %v", err)
	}
	diplomacy.EnsureTradeRoutesForActiveRelations(gs)
	for _, fid := range gs.FactionOrder {
		ai.TakeTurn(gs, fid)
	}

	for _, fid := range gs.FactionOrder {
		regions := gs.RegionsOwnedBy(fid)
		if len(regions) == 0 {
			continue
		}
		minimumNetChange := 100
		if len(regions) == 1 {
			minimumNetChange = 50
		}
		status := victory.GoldEconomyPreview(gs, fid)
		if status.Income <= 0 || status.NetChange >= minimumNetChange {
			continue
		}
		t.Fatalf("%s ilk tur net altın geliri yetersiz: %d (hedef=%d, bölge=%d, gelir=%d, ordu bakımı=%d, bina bakımı=%d)",
			fid, status.NetChange, minimumNetChange, len(regions), status.Income, status.Upkeep, status.BuildingUpkeep)
	}
}

func TestCampaignSaveStateRestoresTerrainAreasAndRuntimeRegions(t *testing.T) {
	areas := []world.TerrainArea{{
		ID:       "saved_area",
		Terrain:  world.TerrainDesert,
		MoveCost: -1,
		Polygons: [][][2]int{{{1, 1}, {4, 1}, {4, 4}, {1, 4}}},
	}}
	saved := campaignSaveState{ScenarioID: "1300_ottoman_rise", TerrainAreas: areas}
	payload, err := json.Marshal(saved)
	if err != nil {
		t.Fatalf("marshal campaign save state: %v", err)
	}
	decoded, err := decodeCampaignSaveState(payload)
	if err != nil {
		t.Fatalf("decode campaign save state: %v", err)
	}
	if len(decoded.TerrainAreas) != 1 || decoded.TerrainAreas[0].ID != "saved_area" {
		t.Fatalf("decoded terrain areas = %#v", decoded.TerrainAreas)
	}

	gs := &state.GameState{
		Regions: map[world.RegionID]*world.Region{
			"base": {ID: "base", Terrain: world.TerrainPlain},
		},
		TerrainAreas: []world.TerrainArea{{ID: "old_area", MoveCost: 0}},
	}
	applyCampaignSaveState(gs, decoded)
	if len(gs.TerrainAreas) != 1 || gs.TerrainAreas[0].ID != "saved_area" {
		t.Fatalf("restored terrain areas = %#v", gs.TerrainAreas)
	}
	if gs.Regions[world.TerrainAreaRegionID("saved_area")] == nil {
		t.Fatal("restored terrain runtime region was not recreated")
	}
}

func TestCampaignSaveStateIgnoresStaleTerrainRegionLock(t *testing.T) {
	areas := []world.TerrainArea{{
		ID:       "blocked_area",
		MoveCost: 0,
		Polygons: [][][2]int{{{1, 1}, {4, 1}, {4, 4}, {1, 4}}},
	}}
	wasUnlocked := false
	saved := campaignSaveState{
		ScenarioID:   "1300_ottoman_rise",
		TerrainAreas: areas,
		Regions: map[world.RegionID]regionSaveState{
			world.TerrainAreaRegionID("blocked_area"): {IsLocked: &wasUnlocked},
		},
	}

	gs := &state.GameState{
		Regions: map[world.RegionID]*world.Region{
			"base": {ID: "base", Terrain: world.TerrainPlain},
		},
	}
	applyCampaignSaveState(gs, saved)

	terrain := gs.Regions[world.TerrainAreaRegionID("blocked_area")]
	if terrain == nil {
		t.Fatal("terrain runtime region was not recreated")
	}
	if !terrain.IsLocked {
		t.Fatal("stale saved unlock state made a move_cost=0 terrain area passable")
	}
}

func TestCampaignSaveStateRestoresDismissedCommanderIDs(t *testing.T) {
	saved := campaignSaveState{
		ScenarioID:            "1300_ottoman_rise",
		DismissedCommanderIDs: map[string]bool{"commander_template": true},
	}
	payload, err := json.Marshal(saved)
	if err != nil {
		t.Fatalf("marshal campaign save state: %v", err)
	}
	decoded, err := decodeCampaignSaveState(payload)
	if err != nil {
		t.Fatalf("decode campaign save state: %v", err)
	}

	gs := &state.GameState{}
	applyCampaignSaveState(gs, decoded)
	if !gs.DismissedCommanderIDs["commander_template"] {
		t.Fatal("dismissed commander ID was not restored")
	}
}
