package state

import (
	"path/filepath"
	"runtime"
	"testing"

	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/world"
)

func TestEnsureTradeNetworkCoverageCreatesBonusFreeGatewayForNewFaction(t *testing.T) {
	gs := &GameState{
		Year: 1300,
		Factions: map[faction.FactionID]*faction.Faction{
			"historic":  {ID: "historic"},
			"new_state": {ID: "new_state"},
		},
		Regions: map[world.RegionID]*world.Region{
			"historic_center": {ID: "historic_center", OwnerID: "historic", TradeCapacity: 5, WorldX: 100, WorldY: 100},
			"new_coast": {
				ID: "new_coast", OwnerID: "new_state", BaseGoldIncome: 8,
				WorldX: 130, WorldY: 100, Neighbors: []world.RegionID{"sea"},
			},
			"sea": {ID: "sea", IsSea: true},
		},
		TradeCenters: world.TradeCenterConfig{Centers: []world.TradeCenterDef{{
			ID: "historic_center", Tier: world.TradeCenterPrimary,
		}}},
	}

	gs.EnsureTradeNetworkCoverage()
	var gateway *world.TradeCenterDef
	for i := range gs.TradeCenters.Centers {
		center := &gs.TradeCenters.Centers[i]
		if center.ID == "new_coast" {
			gateway = center
			break
		}
	}
	if gateway == nil || !gateway.NetworkOnly {
		t.Fatalf("yeni devlet için runtime ağ geçidi oluşturulmadı: %+v", gs.TradeCenters.Centers)
	}
	if !containsTradeNetworkRegionID(gateway.Links, "historic_center") {
		t.Fatalf("ağ geçidi tarihsel merkeze bağlanmadı: %+v", gateway.Links)
	}
	if capacity, income := gs.TradeCenterBenefits(gs.Regions["new_coast"]); capacity != 0 || income != 0 {
		t.Fatalf("runtime ağ geçidi merkez bonusu vermemeli: capacity=%d income=%d", capacity, income)
	}

	gs.EnsureTradeNetworkCoverage()
	gatewayCount := 0
	for _, center := range gs.TradeCenters.Centers {
		if center.NetworkOnly {
			gatewayCount++
		}
	}
	if gatewayCount != 1 {
		t.Fatalf("tekrar kurulum ağ geçidini çoğaltmamalı: %d", gatewayCount)
	}
}

func TestEnsureTradeNetworkCoverageCoversEveryActive1300Faction(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("test dosya yolu çözümlenemedi")
	}
	scenarioPath := filepath.Join(filepath.Dir(file), "..", "..", "assets", "scenarios", "1300_ottoman_rise", "data")
	regions, err := world.LoadRegions(filepath.Join(scenarioPath, "regions.json"))
	if err != nil {
		t.Fatalf("1300 bölgeleri yüklenemedi: %v", err)
	}
	factions, err := faction.LoadFactions(filepath.Join(scenarioPath, "factions.json"))
	if err != nil {
		t.Fatalf("1300 fraksiyonları yüklenemedi: %v", err)
	}
	centers, err := world.LoadTradeCenters(filepath.Join(scenarioPath, "trade_centers.json"), regions)
	if err != nil {
		t.Fatalf("1300 ticaret merkezleri yüklenemedi: %v", err)
	}

	gs := &GameState{Year: 1300, Regions: regions, Factions: factions, TradeCenters: centers}
	gs.EnsureTradeNetworkCoverage()
	staticCenters := make(map[world.RegionID]bool, len(centers.Centers))
	for _, center := range centers.Centers {
		staticCenters[center.ID] = true
	}
	for factionID, f := range factions {
		if f == nil || f.IsEliminated || len(gs.LandRegionsOwnedBy(factionID)) == 0 {
			continue
		}
		hasNetworkNode := false
		for _, center := range gs.TradeCenters.Centers {
			region := regions[center.ID]
			if center.ActiveInYear(gs.Year) && !center.OffMap && region != nil && region.OwnerID == string(factionID) {
				hasNetworkNode = true
				break
			}
		}
		if !hasNetworkNode {
			t.Errorf("aktif kara devleti ticaret ağına bağlı değil: faction=%s", factionID)
		}
	}
	for _, center := range gs.TradeCenters.Centers {
		if !center.NetworkOnly {
			continue
		}
		if len(center.Links) == 0 || len(center.Links) > generatedTradeNetworkLinkCount {
			t.Errorf("runtime ağ geçidinin link sayısı geçersiz: center=%s links=%v", center.ID, center.Links)
		}
		for _, linkID := range center.Links {
			if !staticCenters[linkID] {
				t.Errorf("runtime ağ geçidi tarihsel olmayan düğüme bağlandı: center=%s link=%s", center.ID, linkID)
			}
		}
	}
}

func containsTradeNetworkRegionID(ids []world.RegionID, want world.RegionID) bool {
	for _, id := range ids {
		if id == want {
			return true
		}
	}
	return false
}
