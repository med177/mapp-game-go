package state

import (
	"testing"

	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/world"
)

func TestApplyGrainAidImprovesSatisfactionAndIsLimitedPerTurn(t *testing.T) {
	gs := &GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", Grain: 30},
		},
		Regions: map[world.RegionID]*world.Region{
			"home": {ID: "home", OwnerID: "player", Satisfaction: 20},
		},
	}

	if reason := gs.GrainAidBlockReason("home"); reason != "" {
		t.Fatalf("yardım uygulanabilir olmalıydı, reason=%q", reason)
	}
	if !gs.ApplyGrainAid("home") {
		t.Fatal("tahıl yardımı uygulanmalıydı")
	}
	if gs.Factions["player"].Grain != 18 || gs.Regions["home"].Satisfaction != 30 {
		t.Fatalf("yardım tahıl ve memnuniyeti güncellemedi: faction=%+v region=%+v", gs.Factions["player"], gs.Regions["home"])
	}
	if gs.CanApplyGrainAid("home") {
		t.Fatal("aynı bölgeye aynı turda ikinci yardım yapılmamalıydı")
	}

	gs.AdvanceTurn()
	if !gs.CanApplyGrainAid("home") {
		t.Fatal("tur değişiminden sonra yardım hakkı sıfırlanmalıydı")
	}
}

func TestGrainAidRejectsForeignSiegedAndSatisfiedRegions(t *testing.T) {
	gs := &GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", Grain: 100},
			"other":  {ID: "other", Grain: 100},
		},
		Regions: map[world.RegionID]*world.Region{
			"foreign": {ID: "foreign", OwnerID: "other", Satisfaction: 10},
			"sieged":  {ID: "sieged", OwnerID: "player", Satisfaction: 20},
			"happy":   {ID: "happy", OwnerID: "player", Satisfaction: 90},
		},
		Sieges: map[world.RegionID]*SiegeState{
			"sieged": {RegionID: "sieged"},
		},
	}

	for _, rid := range []world.RegionID{"foreign", "sieged", "happy"} {
		if gs.CanApplyGrainAid(rid) {
			t.Fatalf("%s bölgesinde tahıl yardımı engellenmeliydi", rid)
		}
	}
}

func TestApplyEmergencyGrainSalePreservesStorageCapacity(t *testing.T) {
	gs := &GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", Grain: 160, Gold: 10},
		},
		GrainEconomy: map[faction.FactionID]GrainEconomyStatus{
			"player": {FactionID: "player", StorageCapacity: 100},
		},
		MarketPrices: economy.CurrentMarketPrice{economy.GoodGrain: 10},
	}

	sold, gold := gs.ApplyEmergencyGrainSale(80)
	if sold != 60 || gold != 420 {
		t.Fatalf("yalnız kapasite üstü tahıl %%30 indirimle satılmalıydı, sold=%d gold=%d", sold, gold)
	}
	if gs.Factions["player"].Grain != 100 || gs.Factions["player"].Gold != 430 {
		t.Fatalf("acil satış sonrası rezerv tabanı korunmadı: faction=%+v", gs.Factions["player"])
	}

	sold, gold = gs.ApplyEmergencyGrainSale(1)
	if sold != 0 || gold != 0 {
		t.Fatalf("kapasite altı stoktan acil satış yapılmamalıydı, sold=%d gold=%d", sold, gold)
	}
}

func TestApplyAutomaticGrainExportUsesTradePartnersDeterministically(t *testing.T) {
	gs := &GameState{
		PlayerFactionID: "player",
		AutoGrainExport: true,
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", Grain: 160},
			"alpha":  {ID: "alpha", Gold: 24},
			"beta":   {ID: "beta", Gold: 12},
			"enemy":  {ID: "enemy", Gold: 100},
		},
		GrainEconomy: map[faction.FactionID]GrainEconomyStatus{
			"player": {FactionID: "player", StorageCapacity: 100},
		},
		MarketPrices: economy.CurrentMarketPrice{economy.GoodGrain: 10},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("player", "enemy"): {
				FactionA: "player", FactionB: "enemy", Stance: faction.StanceWar,
			},
		},
		TradeRoutes: []*economy.TradeRoute{
			{FromFactionID: "player", ToFactionID: "beta", AmountPerTurn: 1},
			{FromFactionID: "player", ToFactionID: "alpha", AmountPerTurn: 1},
			{FromFactionID: "player", ToFactionID: "enemy", AmountPerTurn: 1},
		},
	}

	sold, gold := gs.ApplyAutomaticGrainExport()
	if sold != 6 || gold != 36 {
		t.Fatalf("otomatik ihracat partner altını ve kapasite üstünü kullanmalıydı, sold=%d gold=%d", sold, gold)
	}
	if gs.Factions["player"].Grain != 154 || gs.Factions["player"].Gold != 36 {
		t.Fatalf("oyuncu stok/altın otomatik ihracat sonrası hatalı: %+v", gs.Factions["player"])
	}
	if gs.Factions["alpha"].Grain != 4 || gs.Factions["alpha"].Gold != 0 {
		t.Fatalf("alpha ilk deterministik partner olarak 4 tahıl almalıydı: %+v", gs.Factions["alpha"])
	}
	if gs.Factions["beta"].Grain != 2 || gs.Factions["beta"].Gold != 0 {
		t.Fatalf("beta ikinci partner olarak 2 tahıl almalıydı: %+v", gs.Factions["beta"])
	}
	if gs.Factions["enemy"].Grain != 0 || gs.Factions["enemy"].Gold != 100 {
		t.Fatal("savaş halindeki partner otomatik ihracata dahil edilmemeliydi")
	}
}

func TestActiveGrainEventModifiersAffectProductionAndDemandUntilExpiry(t *testing.T) {
	gs := &GameState{
		ActiveRegionEvents: []RegionEventStatus{
			{RegionID: "farm", TurnsLeft: 2, GrainProductionPercent: -35, GrainDemandPercent: 20},
			{RegionID: "farm", TurnsLeft: 1, GrainProductionPercent: -10},
			{RegionID: "farm", TurnsLeft: 0, GrainProductionPercent: 80, GrainDemandPercent: 100},
		},
	}
	region := &world.Region{ID: "farm", Population: 100, BaseGrainOutput: 100}

	if got := gs.RegionGrainProductionModifier(region.ID); got != -45 {
		t.Fatalf("aktif üretim etkileri toplanmalı, got=%d", got)
	}
	if got := gs.CivilianGrainDemandForRegion(region); got != 6 {
		t.Fatalf("aktif tüketim etkisi 5 taban talebi 6 yapmalıydı, got=%d", got)
	}

	gs.ActiveRegionEvents[0].TurnsLeft = 0
	gs.ActiveRegionEvents[1].TurnsLeft = 0
	if got := gs.RegionGrainProductionModifier(region.ID); got != 0 {
		t.Fatalf("süresi biten üretim etkileri yok sayılmalı, got=%d", got)
	}
	if got := gs.CivilianGrainDemandForRegion(region); got != 5 {
		t.Fatalf("süresi biten tüketim etkisi taban talebi geri getirmeli, got=%d", got)
	}
}

func TestRegionProductionSummaryAppliesActiveGrainEventModifier(t *testing.T) {
	gs := &GameState{
		ActiveRegionEvents: []RegionEventStatus{{RegionID: "farm", TurnsLeft: 2, GrainProductionPercent: -35}},
	}
	region := &world.Region{ID: "farm", OwnerID: "player", BaseGrainOutput: 100}

	if got := gs.RegionProductionSummary(region).Grain; got != 65 {
		t.Fatalf("aktif olay üretimi %%35 azaltmalıydı, got=%d", got)
	}
}

func TestRegionMilitaryGrainProductionUsesSameProductionAndCivilianDemandSeams(t *testing.T) {
	gs := &GameState{
		ActiveRegionEvents: []RegionEventStatus{{RegionID: "farm", TurnsLeft: 2, GrainProductionPercent: -35, GrainDemandPercent: 100}},
	}
	region := &world.Region{ID: "farm", OwnerID: "player", Population: 200, BaseGrainOutput: 100}

	if got := gs.RegionMilitaryGrainProduction(region); got != 45 {
		t.Fatalf("askeri ikmal üretimi efektif üretimden aktif sivil talebi düşmeli, got=%d", got)
	}
}

func TestStrategicGrainDemandTargetsThreeMonthsAndSurplusPreservesReserve(t *testing.T) {
	gs := &GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", Grain: 140},
		},
		GrainEconomy: map[faction.FactionID]GrainEconomyStatus{
			"player": {FactionID: "player", TotalDemand: 20, StorageCapacity: 100},
		},
	}

	if got := gs.StrategicGrainDemand("player"); got != 0 {
		t.Fatalf("üç aylık hedef zaten karşılanıyken talep 0 olmalıydı, got=%d", got)
	}
	gs.Factions["player"].Grain = 10
	if got := gs.StrategicGrainDemand("player"); got != 50 {
		t.Fatalf("20 aylık talep ve 10 stok için 50 ithalat ihtiyacı bekleniyordu, got=%d", got)
	}
	gs.Factions["player"].Grain = 140
	if got := gs.StrategicGrainSurplus("player"); got != 40 {
		t.Fatalf("100 rezerv kapasitesi üzerindeki 140 stoktan 40 fazla dönmeli, got=%d", got)
	}
}
