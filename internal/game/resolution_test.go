package game

import (
	"fmt"
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/city"
	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/religion"
	"mapp-game-go/internal/render"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/tech"
	"mapp-game-go/internal/world"
)

func repeatedUnits(typeID string, count int, hp int) []army.Unit {
	out := make([]army.Unit, 0, count)
	for i := 0; i < count; i++ {
		out = append(out, army.Unit{TypeID: typeID, CurrentHP: hp})
	}
	return out
}

func TestCheckRegionUnlocksUnlocksTimedRegionAtTurn(t *testing.T) {
	gs := &state.GameState{
		Turn: 5,
		Regions: map[world.RegionID]*world.Region{
			"locked": {ID: "locked", IsLocked: true, UnlockTurn: 5},
		},
		Armies: map[army.ArmyID]*army.Army{},
	}

	unlocked := checkRegionUnlocks(gs)

	if gs.Regions["locked"].IsLocked {
		t.Fatal("timed region açılmadı")
	}
	if len(unlocked) != 1 || unlocked[0] != "locked" {
		t.Fatalf("beklenen unlock listesi [locked], got=%v", unlocked)
	}
}

func TestCheckRegionUnlocksDoesNotUnlockTimedRegionEarlyByAdjacency(t *testing.T) {
	gs := &state.GameState{
		Turn: 4,
		Regions: map[world.RegionID]*world.Region{
			"src":    {ID: "src", Neighbors: []world.RegionID{"locked"}},
			"locked": {ID: "locked", IsLocked: true, UnlockTurn: 5},
		},
		Armies: map[army.ArmyID]*army.Army{
			"a1": {ID: "a1", RegionID: "src"},
		},
	}

	unlocked := checkRegionUnlocks(gs)

	if !gs.Regions["locked"].IsLocked {
		t.Fatal("timed region erken açıldı")
	}
	if len(unlocked) != 0 {
		t.Fatalf("erken unlock listesi boş olmalı, got=%v", unlocked)
	}
}

func TestCheckRegionUnlocksUnlocksDiscoveryRegionByAdjacency(t *testing.T) {
	gs := &state.GameState{
		Turn: 4,
		Regions: map[world.RegionID]*world.Region{
			"src":    {ID: "src", Neighbors: []world.RegionID{"locked"}},
			"locked": {ID: "locked", IsLocked: true, UnlockTurn: 0},
		},
		Armies: map[army.ArmyID]*army.Army{
			"a1": {ID: "a1", RegionID: "src"},
		},
	}

	unlocked := checkRegionUnlocks(gs)

	if gs.Regions["locked"].IsLocked {
		t.Fatal("discovery region komşulukla açılmadı")
	}
	if len(unlocked) != 1 || unlocked[0] != "locked" {
		t.Fatalf("beklenen unlock listesi [locked], got=%v", unlocked)
	}
}

func TestCheckRebellionsClearsProductionOrdersForLostRegion(t *testing.T) {
	gs := &state.GameState{
		Regions: map[world.RegionID]*world.Region{
			"rebel": {ID: "rebel", OwnerID: "p1", Satisfaction: 10},
			"other": {ID: "other", OwnerID: "p1", Satisfaction: 60},
		},
		ProductionQueue: []state.ProductionOrder{
			{ID: "prod_1", Kind: productionKindBuilding, FactionID: "p1", RegionID: "rebel", TypeID: "walls", TurnsLeft: 2},
			{ID: "prod_2", Kind: productionKindUnit, FactionID: "p1", RegionID: "other", TypeID: "infantry", TurnsLeft: 2},
		},
	}

	checkRebellions(gs)

	if gs.Regions["rebel"].OwnerID != "" {
		t.Fatalf("isyanda bölge sahipsiz kalmalıydı, got=%s", gs.Regions["rebel"].OwnerID)
	}
	if gs.Regions["rebel"].Satisfaction != 50 {
		t.Fatalf("isyandan sonra memnuniyet resetlenmeliydi, got=%d", gs.Regions["rebel"].Satisfaction)
	}
	if got := len(gs.ProductionQueue); got != 1 {
		t.Fatalf("kaybedilen bolgedeki uretimler silinmeliydi, got=%d queue=%+v", got, gs.ProductionQueue)
	}
	if gs.ProductionQueue[0].RegionID != "other" {
		t.Fatalf("yalnız diğer bölgenin üretimi kalmalıydı, queue=%+v", gs.ProductionQueue)
	}
}

func TestCivilianGrainDemandRoundsUpAndIgnoresEmptyPopulation(t *testing.T) {
	if got := civilianGrainDemand(&world.Region{Population: 40}); got != 2 {
		t.Fatalf("40 nüfus için 2 tahıl bekleniyordu, got=%d", got)
	}
	if got := civilianGrainDemand(&world.Region{Population: 41}); got != 3 {
		t.Fatalf("41 nüfus için yukarı yuvarlanan 3 tahıl bekleniyordu, got=%d", got)
	}
	if got := civilianGrainDemand(&world.Region{Population: 0}); got != 0 {
		t.Fatalf("sıfır nüfus tüketim üretmemeli, got=%d", got)
	}
}

func TestEffectiveArmyGrainUpkeepUsesMovementAndSiegeCoefficients(t *testing.T) {
	gs := &state.GameState{
		UnitTypes: map[string]*army.UnitType{
			"inf": {ID: "inf", GrainUpkeep: 4},
		},
		Armies: map[army.ArmyID]*army.Army{
			"stationary": {ID: "stationary", Units: repeatedUnits("inf", 1, 100), MovePoints: 2, MaxMovePoints: 2},
			"moving":     {ID: "moving", Units: repeatedUnits("inf", 1, 100), MovePoints: 1, MaxMovePoints: 2},
			"garrison":   {ID: "garrison", IsGarrison: true, Units: repeatedUnits("inf", 1, 100), MovePoints: 2, MaxMovePoints: 2},
			"attacker":   {ID: "attacker", RegionID: "fort", Units: repeatedUnits("inf", 1, 100), MovePoints: 0, MaxMovePoints: 2},
			"defender":   {ID: "defender", RegionID: "fort", Units: repeatedUnits("inf", 1, 100), MovePoints: 2, MaxMovePoints: 2},
		},
		Sieges: map[world.RegionID]*state.SiegeState{
			"fort": {RegionID: "fort", AttackerArmyID: "attacker", DefenderArmyID: "defender"},
		},
	}

	want := map[army.ArmyID]int{
		"stationary": 4,
		"moving":     6,
		"garrison":   3,
		"attacker":   8,
		"defender":   5,
	}
	for id, expected := range want {
		if got := gs.EffectiveArmyGrainUpkeep(gs.Armies[id]); got != expected {
			t.Fatalf("%s için efektif tahıl bakımı %d olmalıydı, got=%d", id, expected, got)
		}
	}
}

func TestApplySeasonEffectsPreservesMovementUsageForGrainEconomy(t *testing.T) {
	gs := &state.GameState{
		Month: 4,
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", Grain: 100},
		},
		Regions: map[world.RegionID]*world.Region{
			"home": {ID: "home", OwnerID: "player"},
		},
		Armies: map[army.ArmyID]*army.Army{
			"field": {
				ID: "field", OwnerID: "player", RegionID: "home",
				Units: repeatedUnits("inf", 1, 100), MovePoints: 1, MaxMovePoints: 2,
			},
		},
		UnitTypes: map[string]*army.UnitType{
			"inf": {ID: "inf", GrainUpkeep: 4, MovementPoints: 2},
		},
	}

	applySeasonEffects(gs)
	if gs.Armies["field"].MovePoints != gs.Armies["field"].MaxMovePoints {
		t.Fatalf("sezon başlangıcında hareket puanı yenilenmeliydi, got=%d/%d", gs.Armies["field"].MovePoints, gs.Armies["field"].MaxMovePoints)
	}
	applyEconomyTick(gs)

	if got := gs.GrainEconomy["player"].ArmyUpkeep; got != 6 {
		t.Fatalf("reset öncesi hareket kullanımı ekonomi hesabına taşınmalıydı, got=%d", got)
	}
}

func TestGrainEconomyStatusUsesStockpileMonthsAndShortageLevels(t *testing.T) {
	warning := grainEconomyStatus("player", 30, 0, 10, 0, 0)
	if warning.MonthsOfSupply != 2 || warning.SupplyLevel != state.GrainSupplyWarning || warning.Shortage != 0 {
		t.Fatalf("2 aylık stok uyarı seviyesinde olmalıydı, got=%+v", warning)
	}

	critical := grainEconomyStatus("player", 10, 0, 10, 0, 0)
	if critical.MonthsOfSupply != 0 || critical.SupplyLevel != state.GrainSupplyCritical || critical.Shortage != 0 {
		t.Fatalf("1 aydan az stok kritik seviyede olmalıydı, got=%+v", critical)
	}

	famine := grainEconomyStatus("player", 5, 0, 10, 0, 0)
	if famine.Stockpile != 0 || famine.Shortage != 5 || famine.SupplyLevel != state.GrainSupplyFamine {
		t.Fatalf("talep karşılanamadığında kıtlık durumu hatalı, got=%+v", famine)
	}
}

func TestApplyEconomyTickBindsGrainSupplyToArmyMorale(t *testing.T) {
	tests := []struct {
		name       string
		grain      int
		morale     int
		wantMorale int
		wantDelta  int
	}{
		{name: "istikrar toparlar", grain: 100, morale: 40, wantMorale: 41, wantDelta: 1},
		{name: "kıtlık düşürür", grain: 0, morale: 40, wantMorale: 34, wantDelta: -6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gs := &state.GameState{
				Month: 4,
				Factions: map[faction.FactionID]*faction.Faction{
					"player": {ID: "player", Grain: tt.grain},
				},
				Regions: map[world.RegionID]*world.Region{
					"home": {ID: "home", OwnerID: "player", Population: 200},
				},
				Armies: map[army.ArmyID]*army.Army{
					"field": {ID: "field", OwnerID: "player", RegionID: "home", Morale: tt.morale, Units: []army.Unit{{TypeID: "inf", CurrentHP: 100}}},
				},
				UnitTypes: map[string]*army.UnitType{
					"inf": {ID: "inf", GrainUpkeep: 0},
				},
			}

			applyEconomyTick(gs)
			status := gs.GrainEconomy["player"]
			if got := gs.Armies["field"].Morale; got != tt.wantMorale {
				t.Fatalf("ordu morali arz seviyesine göre güncellenmeli, got=%d want=%d status=%+v", got, tt.wantMorale, status)
			}
			if status.ArmyMoraleDelta != tt.wantDelta {
				t.Fatalf("moral değişimi ekonomi raporuna yazılmalı, got=%d want=%d", status.ArmyMoraleDelta, tt.wantDelta)
			}
		})
	}
}

func TestApplyEconomyTickUsesActiveGrainEventModifiers(t *testing.T) {
	gs := &state.GameState{
		Month: 4,
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", Grain: 0},
		},
		Regions: map[world.RegionID]*world.Region{
			"farm": {
				ID:              "farm",
				OwnerID:         "player",
				Population:      200,
				BaseGrainOutput: 100,
			},
		},
		ActiveRegionEvents: []state.RegionEventStatus{{
			RegionID:               "farm",
			TurnsLeft:              2,
			GrainProductionPercent: -50,
			GrainDemandPercent:     100,
		}},
	}

	applyEconomyTick(gs)
	status := gs.GrainEconomy["player"]
	if status.Production != 50 || status.CivilianDemand != 20 || status.Stockpile != 30 {
		t.Fatalf("ekonomi aktif olay üretim ve tüketim etkisini kullanmalıydı, got=%+v", status)
	}
}

func TestGrainStorageCapacityUsesDemandAndMinimumReserve(t *testing.T) {
	if got := grainStorageCapacity(20, 10, 0); got != 150 {
		t.Fatalf("20 sivil + 10 ordu talebi için 150 kapasite bekleniyordu, got=%d", got)
	}
	if got := grainStorageCapacity(20, 10, 100); got != 250 {
		t.Fatalf("100 kapasiteli ambar bonusu toplam kapasiteye eklenmeliydi, got=%d", got)
	}
	if got := grainStorageCapacity(1, 0, 0); got != grainMinimumStorageCapacity {
		t.Fatalf("küçük devletlerde minimum depo kapasitesi korunmalıydı, got=%d", got)
	}
	if got := grainStorageCapacity(0, 0, 0); got != 0 {
		t.Fatalf("talep yoksa kapasite hesabı sıfır olmalıydı, got=%d", got)
	}
	if got := grainSpoilage(100, 100); got != 0 {
		t.Fatalf("kapasite sınırında bozulma olmamalıydı, got=%d", got)
	}
	if got := grainSpoilage(150, 100); got != 1 {
		t.Fatalf("küçük stok fazlası en az 1 bozulma üretmeliydi, got=%d", got)
	}
}

func TestApplyEconomyTickSpoilsExcessGrainSoftly(t *testing.T) {
	gs := &state.GameState{
		Month: 4,
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", Grain: 500},
		},
		Regions: map[world.RegionID]*world.Region{
			"home": {ID: "home", OwnerID: "player", Population: 200},
		},
		Armies:    map[army.ArmyID]*army.Army{},
		UnitTypes: map[string]*army.UnitType{},
	}

	applyEconomyTick(gs)

	status := gs.GrainEconomy["player"]
	if status.StorageCapacity != 100 || status.Spoiled != 7 || status.Stockpile != 483 {
		t.Fatalf("fazla tahıl yumuşak bozulma ile azalmalıydı, got=%+v", status)
	}
	if gs.Factions["player"].Grain != 483 {
		t.Fatalf("fraksiyon stoku bozulma sonrası 483 olmalıydı, got=%d", gs.Factions["player"].Grain)
	}
}

func TestApplyEconomyTickUsesGranaryStorageBonus(t *testing.T) {
	gs := &state.GameState{
		Month: 4,
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", Grain: 500},
		},
		Regions: map[world.RegionID]*world.Region{
			"home": {ID: "home", OwnerID: "player", Population: 200, Buildings: []string{"granary"}},
		},
		BuildingTypes: map[string]*city.Building{
			"granary": {ID: "granary", StorageCapacity: 100, MaxPerRegion: 3},
		},
		Armies:    map[army.ArmyID]*army.Army{},
		UnitTypes: map[string]*army.UnitType{},
	}

	applyEconomyTick(gs)

	status := gs.GrainEconomy["player"]
	if status.StorageCapacity != 160 || status.Spoiled != 6 {
		t.Fatalf("ambar kapasite bonusu bozulmayı azaltmalıydı, got=%+v", status)
	}
	if gs.Factions["player"].Grain != 484 {
		t.Fatalf("ambar varken stok bozulma sonrası 484 olmalıydı, got=%d", gs.Factions["player"].Grain)
	}
}

func TestApplyEconomyTickFundsAnnualPopulationGrowthFromGrainSurplus(t *testing.T) {
	gs := &state.GameState{
		Month: 11,
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", Grain: 160},
		},
		Regions: map[world.RegionID]*world.Region{
			"home": {
				ID:           "home",
				OwnerID:      "player",
				Population:   200,
				Satisfaction: 70,
			},
		},
		Armies:    map[army.ArmyID]*army.Army{},
		UnitTypes: map[string]*army.UnitType{},
	}

	applyEconomyTick(gs)

	status := gs.GrainEconomy["player"]
	if got := gs.Regions["home"].Population; got != 202 {
		t.Fatalf("stabil rezerv fazlası nüfusu 2 artırmalıydı, got=%d", got)
	}
	if got := gs.Factions["player"].Grain; got != 145 {
		t.Fatalf("bozulma ve nüfus yatırımı sonrası tahıl 145 olmalıydı, got=%d", got)
	}
	if status.PopulationGrowth != 2 || status.GrowthGrainSpent != 4 {
		t.Fatalf("nüfus büyümesi tahıl raporuna yazılmalıydı, got=%+v", status)
	}
	if status.NetChange != -14 || status.Stockpile != 145 {
		t.Fatalf("nüfus yatırımı net değişim ve stoku güncellemeli, got=%+v", status)
	}
}

func TestApplyEconomyTickFundsArmyReplenishmentOnlyFromCapacitySurplus(t *testing.T) {
	gs := &state.GameState{
		Month: 4,
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", Grain: 116},
		},
		Regions: map[world.RegionID]*world.Region{
			"home": {ID: "home", OwnerID: "player", Population: 20},
		},
		Armies: map[army.ArmyID]*army.Army{
			"a-first": {ID: "a-first", OwnerID: "player", RegionID: "home", Units: []army.Unit{{TypeID: "inf", CurrentHP: 60}}},
			"z-last":  {ID: "z-last", OwnerID: "player", RegionID: "home", Units: []army.Unit{{TypeID: "inf", CurrentHP: 60}}},
		},
		UnitTypes: map[string]*army.UnitType{
			"inf": {ID: "inf", GrainUpkeep: 0},
		},
	}

	applyEconomyTick(gs)

	if got := gs.Armies["a-first"].Units[0].CurrentHP; got != 70 {
		t.Fatalf("ilk ordu faction/army ID sırasına göre 10 HP almalıydı, got=%d", got)
	}
	if got := gs.Armies["z-last"].Units[0].CurrentHP; got != 64 {
		t.Fatalf("kalan kapasite fazlası ikinci orduya gitmeli, got=%d", got)
	}
	if got := gs.Factions["player"].Grain; got != 100 {
		t.Fatalf("yenileme rezerv kapasite tabanının altına inmemeli, got=%d", got)
	}
	status := gs.GrainEconomy["player"]
	if status.ReplenishmentHP != 14 || status.ReplenishmentGrainSpent != 14 {
		t.Fatalf("yenileme miktarı ve tahıl harcaması raporlanmalıydı, got=%+v", status)
	}
	if status.NetChange != -15 || status.Stockpile != 100 {
		t.Fatalf("yenileme net değişim ve stoku güncellemeli, got=%+v", status)
	}
}

func TestApplyEconomyTickDoesNotFundArmyReplenishmentBelowReserveCapacity(t *testing.T) {
	gs := &state.GameState{
		Month: 4,
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", Grain: 100},
		},
		Regions: map[world.RegionID]*world.Region{
			"home": {ID: "home", OwnerID: "player", Population: 20},
		},
		Armies: map[army.ArmyID]*army.Army{
			"army": {ID: "army", OwnerID: "player", RegionID: "home", Units: []army.Unit{{TypeID: "inf", CurrentHP: 60}}},
		},
		UnitTypes: map[string]*army.UnitType{
			"inf": {ID: "inf", GrainUpkeep: 0},
		},
	}

	applyEconomyTick(gs)

	if got := gs.Armies["army"].Units[0].CurrentHP; got != 60 {
		t.Fatalf("rezerv kapasitesi altındaki ordu ücretsiz ek yenileme almamalıydı, got=%d", got)
	}
	if status := gs.GrainEconomy["player"]; status.ReplenishmentGrainSpent != 0 {
		t.Fatalf("rezerv kapasitesi altındayken tahıl harcanmamalıydı, got=%+v", status)
	}
}

func TestApplyEconomyTickDoesNotFundPopulationGrowthDuringShortageOrOutsideAnnualMonth(t *testing.T) {
	tests := []struct {
		name         string
		month        int
		grain        int
		satisfaction int
		wantPop      int
	}{
		{name: "kıtlık", month: 11, grain: 5, satisfaction: 70, wantPop: 200},
		{name: "yıllık ay değil", month: 10, grain: 160, satisfaction: 70, wantPop: 200},
		{name: "isyan riski", month: 11, grain: 160, satisfaction: 20, wantPop: 200},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gs := &state.GameState{
				Month: tt.month,
				Factions: map[faction.FactionID]*faction.Faction{
					"player": {ID: "player", Grain: tt.grain},
				},
				Regions: map[world.RegionID]*world.Region{
					"home": {
						ID:           "home",
						OwnerID:      "player",
						Population:   200,
						Satisfaction: tt.satisfaction,
					},
				},
				Armies:    map[army.ArmyID]*army.Army{},
				UnitTypes: map[string]*army.UnitType{},
			}

			applyEconomyTick(gs)

			if got := gs.Regions["home"].Population; got != tt.wantPop {
				t.Fatalf("nüfus büyümemeli, got=%d", got)
			}
		})
	}
}

func TestApplyEconomyTickAppliesGrainShortageStabilityEffects(t *testing.T) {
	gs := &state.GameState{
		Month:           4,
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", Gold: 100, Grain: 10},
		},
		Regions: map[world.RegionID]*world.Region{
			"home": {
				ID:              "home",
				OwnerID:         "player",
				Population:      200,
				Satisfaction:    50,
				TaxRate:         50,
				BaseGoldIncome:  100,
				BaseGrainOutput: 0,
			},
		},
		Armies:    map[army.ArmyID]*army.Army{},
		UnitTypes: map[string]*army.UnitType{},
	}

	report := applyEconomyTick(gs)
	status := gs.GrainEconomy["player"]
	if status.SupplyLevel != state.GrainSupplyCritical || status.MonthsOfSupply != 0 {
		t.Fatalf("kritik tahıl durumu bekleniyordu, got=%+v", status)
	}
	if gs.Factions["player"].Grain != 0 || gs.Factions["player"].Gold != 145 {
		t.Fatalf("kritik rezerv etkileri uygulanmadı, faction=%+v", gs.Factions["player"])
	}
	if gs.Regions["home"].Satisfaction != 48 {
		t.Fatalf("kritik tahıl rezervi memnuniyeti 2 azaltmalıydı, got=%d", gs.Regions["home"].Satisfaction)
	}
	if report.PlayerGrainStatus.SupplyLevel != state.GrainSupplyCritical {
		t.Fatalf("oyuncu tahıl raporu kritik durumu taşımadı, got=%+v", report.PlayerGrainStatus)
	}
}

func TestApplyEconomyTickCombinesSatisfactionModifiers(t *testing.T) {
	regions := make(map[world.RegionID]*world.Region, 21)
	for i := 1; i <= 21; i++ {
		rid := world.RegionID(fmt.Sprintf("region-%d", i))
		regions[rid] = &world.Region{ID: rid, OwnerID: "player", Satisfaction: 50, TaxRate: 50}
	}
	regions["region-1"].Buildings = []string{"market", "farm", "barracks", "port"}

	gs := &state.GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", Grain: 100},
			"enemy":  {ID: "enemy"},
		},
		Regions: regions,
		Armies: map[army.ArmyID]*army.Army{
			"garrison": {
				ID: "garrison", OwnerID: "player", RegionID: "region-1",
				Units: repeatedUnits("guard", 1, 100),
			},
		},
		UnitTypes: map[string]*army.UnitType{
			"guard": {ID: "guard", Attack: 100, HP: 100},
		},
		BuildingTypes: map[string]*city.Building{
			"market":   {ID: "market", SatBonus: 1},
			"farm":     {ID: "farm", SatBonus: 1},
			"barracks": {ID: "barracks", SatBonus: -1},
			"port":     {ID: "port", SatBonus: 1},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("player", "enemy"): {
				FactionA: "player", FactionB: "enemy", Stance: faction.StanceWar,
			},
		},
	}

	applyEconomyTick(gs)

	// Bölge 1: bina +2 + savaş -1 + genişleme -1 + ordu gücü 100 => +10.
	if got := gs.Regions["region-1"].Satisfaction; got != 60 {
		t.Fatalf("toplam memnuniyet deltası bölge 1 için +10 olmalıydı, got=%d", got)
	}
	// Diğer bölgeler yalnızca savaş yorgunluğu ve 20+ bölge cezasını alır.
	if got := gs.Regions["region-2"].Satisfaction; got != 48 {
		t.Fatalf("savaş ve genişleme cezaları toplam -2 olmalıydı, got=%d", got)
	}
}

func TestRegionArmySatisfactionBonusScalesAndCaps(t *testing.T) {
	gs := &state.GameState{
		UnitTypes: map[string]*army.UnitType{
			"guard": {ID: "guard", Attack: 75, HP: 100},
		},
		Armies: map[army.ArmyID]*army.Army{
			"a": {ID: "a", OwnerID: "player", RegionID: "home", Units: repeatedUnits("guard", 1, 100)},
		},
	}
	region := &world.Region{ID: "home", OwnerID: "player"}
	if got := regionArmySatisfactionBonus(gs, region); got != 7 {
		t.Fatalf("75 güç +7 bonus vermeliydi, got=%d", got)
	}
	gs.Armies["b"] = &army.Army{ID: "b", OwnerID: "player", RegionID: "home", Units: repeatedUnits("guard", 1, 100)}
	if got := regionArmySatisfactionBonus(gs, region); got != 10 {
		t.Fatalf("ordu bonusu +10 ile sınırlanmalıydı, got=%d", got)
	}
}

func TestApplyEconomyTickAppliesAnnualSatisfactionDecayAtYearEnd(t *testing.T) {
	tests := []struct {
		name  string
		month int
		want  int
	}{
		{name: "aralık", month: 12, want: 49},
		{name: "yıl içi", month: 11, want: 50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gs := &state.GameState{
				Month: tt.month,
				Factions: map[faction.FactionID]*faction.Faction{
					"player": {ID: "player", Grain: 100},
				},
				Regions: map[world.RegionID]*world.Region{
					"home": {ID: "home", OwnerID: "player", TaxRate: 50, Satisfaction: 50},
				},
				Armies:    map[army.ArmyID]*army.Army{},
				UnitTypes: map[string]*army.UnitType{},
			}

			applyEconomyTick(gs)

			if got := gs.Regions["home"].Satisfaction; got != tt.want {
				t.Fatalf("memnuniyet %d. ay için %d olmalıydı, got=%d", tt.month, tt.want, got)
			}
		})
	}
}

func TestApplyEconomyTickConsumesCivilianGrain(t *testing.T) {
	gs := &state.GameState{
		Month: 4,
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", Grain: 100},
		},
		Regions: map[world.RegionID]*world.Region{
			"home": {
				ID:              "home",
				OwnerID:         "player",
				Population:      40,
				BaseGrainOutput: 0,
			},
		},
		Armies:    map[army.ArmyID]*army.Army{},
		UnitTypes: map[string]*army.UnitType{},
	}

	applyEconomyTick(gs)

	if got := gs.Factions["player"].Grain; got != 98 {
		t.Fatalf("40 nüfus 2 tahıl tüketmeliydi, got=%d", got)
	}
}

func TestApplyEconomyTickAddsTradeIncome(t *testing.T) {
	gs := &state.GameState{
		Month: 4,
		Factions: map[faction.FactionID]*faction.Faction{
			"a": {ID: "a", Religion: religion.Catholic, Gold: 10, Grain: 0, Spice: 10},
			"b": {ID: "b", Religion: religion.Catholic, Gold: 30, Grain: 0, Spice: 0},
		},
		Regions: map[world.RegionID]*world.Region{
			"a1": {ID: "a1", OwnerID: "a", TaxRate: 50, Satisfaction: 50},
			"b1": {ID: "b1", OwnerID: "b", TaxRate: 50, Satisfaction: 50},
		},
		TradeRoutes: []*economy.TradeRoute{
			{FromFactionID: "a", ToFactionID: "b", Good: economy.GoodSpice, AmountPerTurn: 2, GoldPerUnit: 12},
		},
		Armies:    map[army.ArmyID]*army.Army{},
		UnitTypes: map[string]*army.UnitType{},
	}

	applyEconomyTick(gs)

	// a: 10 (başlangıç) + 24 (2 spice * 12 gold, b'den ödeme) = 34, spice: 10 - 2 = 8
	if gs.Factions["a"].Gold != 34 {
		t.Fatalf("ticaret geliri altına eklenmedi, got=%d", gs.Factions["a"].Gold)
	}
	if gs.Factions["a"].Spice != 8 {
		t.Fatalf("ticaret rotası malı kaynaktan çıkarmalıydı, got=%d", gs.Factions["a"].Spice)
	}
	// b: 30 (başlangıç) - 24 (ticaret ödemesi) = 6, spice: 0 + 2 = 2
	if gs.Factions["b"].Gold != 6 {
		t.Fatalf("ticaret rotası hedeften altın çıkarmalıydı, got=%d", gs.Factions["b"].Gold)
	}
	if gs.Factions["b"].Spice != 2 {
		t.Fatalf("ticaret rotası malı hedefe eklemeliydi, got=%d", gs.Factions["b"].Spice)
	}
}

func TestApplyEconomyTickAppliesMarketGoldBonusToPassiveTrade(t *testing.T) {
	gs := &state.GameState{
		Month: 4,
		Factions: map[faction.FactionID]*faction.Faction{
			"a": {
				ID: "a",
				Research: faction.ResearchState{
					Completed: map[string]bool{"guilds": true},
				},
			},
		},
		Regions: map[world.RegionID]*world.Region{
			"a1": {ID: "a1", OwnerID: "a", TaxRate: 50, Satisfaction: 50, TradeCapacity: 5},
		},
		Armies: map[army.ArmyID]*army.Army{},
		TechTypes: map[string]*tech.Technology{
			"guilds": {ID: "guilds", Effects: tech.Effects{MarketGoldMod: 1}},
		},
	}

	applyEconomyTick(gs)

	if gs.Factions["a"].Gold != 22 {
		t.Fatalf("market bonus pasif ticaret gelirini ikiye katlamalıydı, got=%d", gs.Factions["a"].Gold)
	}
}

func TestApplyEconomyTickTransfersVassalTributeToOverlord(t *testing.T) {
	gs := &state.GameState{
		Month: 4,
		Factions: map[faction.FactionID]*faction.Faction{
			"lord":   {ID: "lord", Gold: 50},
			"vassal": {ID: "vassal", Gold: 0, OverlordID: "lord"},
		},
		Regions: map[world.RegionID]*world.Region{
			"v1": {ID: "v1", OwnerID: "vassal", TaxRate: 50, Satisfaction: 50, BaseGoldIncome: 20},
		},
		Armies:    map[army.ArmyID]*army.Army{},
		UnitTypes: map[string]*army.UnitType{},
	}

	applyEconomyTick(gs)

	if got := gs.Factions["vassal"].Gold; got != 8 {
		t.Fatalf("vassal gelirinin %%20 haraç sonrası 8 kalmalıydı, got=%d", got)
	}
	if got := gs.Factions["lord"].Gold; got != 52 {
		t.Fatalf("overlord haracı almalıydı, got=%d", got)
	}
}

func TestApplySeasonEffectsAddsNavalMoveBonus(t *testing.T) {
	gs := &state.GameState{
		Month: 4,
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {
				ID: "player",
				Research: faction.ResearchState{
					Completed: map[string]bool{"navigation": true},
				},
			},
		},
		Armies: map[army.ArmyID]*army.Army{
			"fleet": {ID: "fleet", OwnerID: "player", RegionID: "sea_1", IsNaval: true},
			"land":  {ID: "land", OwnerID: "player", RegionID: "land_1"},
		},
		Regions: map[world.RegionID]*world.Region{
			"sea_1":  {ID: "sea_1", IsSea: true},
			"land_1": {ID: "land_1", OwnerID: "player"},
		},
		TechTypes: map[string]*tech.Technology{
			"navigation": {ID: "navigation", Effects: tech.Effects{NavalMoveBonus: 1}},
		},
	}

	applySeasonEffects(gs)

	if gs.Armies["fleet"].MaxMovePoints != 3 {
		t.Fatalf("naval move bonus filoya +1 hareket vermeliydi, got=%d", gs.Armies["fleet"].MaxMovePoints)
	}
	if gs.Armies["land"].MaxMovePoints != 2 {
		t.Fatalf("naval move bonus kara ordusunu etkilememeliydi, got=%d", gs.Armies["land"].MaxMovePoints)
	}
}

func TestApplySeasonEffectsUsesSlowestUnitAndClimate(t *testing.T) {
	gs := &state.GameState{
		Month: 6,
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player"},
		},
		Armies: map[army.ArmyID]*army.Army{
			"cavalry":  {ID: "cavalry", OwnerID: "player", Units: []army.Unit{{TypeID: "cav"}}},
			"infantry": {ID: "infantry", OwnerID: "player", Units: []army.Unit{{TypeID: "inf"}}},
			"siege":    {ID: "siege", OwnerID: "player", Units: []army.Unit{{TypeID: "siege"}}},
			"mixed":    {ID: "mixed", OwnerID: "player", Units: []army.Unit{{TypeID: "cav"}, {TypeID: "siege"}}},
		},
		UnitTypes: map[string]*army.UnitType{
			"cav":   {ID: "cav", Category: army.CategoryCavalry, MovementPoints: 3},
			"inf":   {ID: "inf", Category: army.CategoryInfantry, MovementPoints: 2},
			"siege": {ID: "siege", Category: army.CategorySiege, MovementPoints: 1},
		},
	}

	applySeasonEffects(gs)
	if got := gs.Armies["cavalry"].MaxMovePoints; got != 3 {
		t.Fatalf("yazın yalnız süvari ordusu 3 ilerlemeliydi, got=%d", got)
	}
	if got := gs.Armies["infantry"].MaxMovePoints; got != 2 {
		t.Fatalf("yazın yalnız piyade ordusu 2 ilerlemeliydi, got=%d", got)
	}
	if got := gs.Armies["siege"].MaxMovePoints; got != 1 {
		t.Fatalf("yazın yalnız kuşatma ordusu 1 ilerlemeliydi, got=%d", got)
	}
	if got := gs.Armies["mixed"].MaxMovePoints; got != 1 {
		t.Fatalf("karışık ordu en yavaş birim olan kuşatma hızını kullanmalıydı, got=%d", got)
	}

	gs.Month = 9 // sonbahar: %95 iklim çarpanı
	applySeasonEffects(gs)
	if got := gs.Armies["cavalry"].MaxMovePoints; got != 2 {
		t.Fatalf("sonbaharda süvari hareketi iklimle 2 olmalıydı, got=%d", got)
	}
}

func TestApplySeasonEffectsAddsCommanderMoveBonus(t *testing.T) {
	tactician := army.NewCommander("cmd_tactician", "Taktisyen")
	tactician.Experience = army.CommanderLevel3XP
	tactician.Normalize()

	gs := &state.GameState{
		Month: 4,
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player"},
		},
		Armies: map[army.ArmyID]*army.Army{
			"land": {ID: "land", OwnerID: "player", RegionID: "land_1", Commander: tactician},
		},
		Regions: map[world.RegionID]*world.Region{
			"land_1": {ID: "land_1", OwnerID: "player"},
		},
	}

	applySeasonEffects(gs)

	if gs.Armies["land"].MaxMovePoints != 3 {
		t.Fatalf("taktisyen komutan kara ordusuna +1 hareket vermeliydi, got=%d", gs.Armies["land"].MaxMovePoints)
	}
}

func TestApplyReligionConversionUsesTechSpeedBonus(t *testing.T) {
	gs := &state.GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"owner": {
				ID:       "owner",
				Religion: religion.Catholic,
				Research: faction.ResearchState{
					Completed: map[string]bool{"proselytism": true},
				},
			},
		},
		Regions: map[world.RegionID]*world.Region{
			"r1": {ID: "r1", OwnerID: "owner", Religion: string(religion.Sunni), Satisfaction: 60},
		},
		TechTypes: map[string]*tech.Technology{
			"proselytism": {ID: "proselytism", Effects: tech.Effects{ConversionSpeedMod: 1}},
		},
	}

	applyReligionConversion(gs)

	if gs.Regions["r1"].ConversionTurns != 2 {
		t.Fatalf("conversion speed bonus bir turda +2 ilerletmeliydi, got=%d", gs.Regions["r1"].ConversionTurns)
	}
}

func TestStartResearchSwitchesActiveResearch(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {
				ID:   "player",
				Gold: 100,
				Research: faction.ResearchState{
					ActiveID:  "old",
					TurnsLeft: 3,
					Completed: map[string]bool{},
				},
			},
		},
		TechTypes: map[string]*tech.Technology{
			"old": {ID: "old", NameTR: "Eski", GoldCost: 20, TurnsRequired: 4},
			"new": {ID: "new", NameTR: "Yeni", GoldCost: 30, TurnsRequired: 5},
		},
	}
	g := &Game{gs: gs, renderer: &render.Renderer{}}

	g.startResearch("new")

	player := gs.Factions["player"]
	if player.Research.ActiveID != "new" {
		t.Fatalf("active research yeni tech'e gecmeliydi, got=%q", player.Research.ActiveID)
	}
	if player.Research.TurnsLeft != 5 {
		t.Fatalf("yeni turns left ayarlanmadi, got=%d", player.Research.TurnsLeft)
	}
	if player.Research.PausedTurns["old"] != 3 {
		t.Fatalf("eski research kalan turuyla pause'a alinmaliydi, got=%d", player.Research.PausedTurns["old"])
	}
	if player.Gold != 70 {
		t.Fatalf("switch sadece yeni research maliyetini dusmeli, got=%d", player.Gold)
	}
}

func TestStartResearchResumesPausedResearchWithoutChargingAgain(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {
				ID:   "player",
				Gold: 70,
				Research: faction.ResearchState{
					Completed:   map[string]bool{},
					PausedTurns: map[string]int{"old": 3},
				},
			},
		},
		TechTypes: map[string]*tech.Technology{
			"old": {ID: "old", NameTR: "Eski", GoldCost: 20, TurnsRequired: 4},
		},
	}
	g := &Game{gs: gs, renderer: &render.Renderer{}}

	g.startResearch("old")

	player := gs.Factions["player"]
	if player.Research.ActiveID != "old" || player.Research.TurnsLeft != 3 {
		t.Fatalf("paused research kaldigi yerden devam etmeli, got=%s/%d", player.Research.ActiveID, player.Research.TurnsLeft)
	}
	if player.Gold != 70 {
		t.Fatalf("resume tekrar altin dusmemeli, got=%d", player.Gold)
	}
	if _, ok := player.Research.PausedTurns["old"]; ok {
		t.Fatal("resume edilen paused kayit silinmeliydi")
	}
}

func TestApplyEconomyTickRegionalLogisticsAttritionEscalates(t *testing.T) {
	gs := &state.GameState{
		Month:           4,
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", Grain: 40},
		},
		Regions: map[world.RegionID]*world.Region{
			"gelibolu": {ID: "gelibolu", OwnerID: "player", TaxRate: 50, Satisfaction: 50, BaseGrainOutput: 4},
		},
		Armies: map[army.ArmyID]*army.Army{
			"a1": {ID: "a1", OwnerID: "player", RegionID: "gelibolu", Units: repeatedUnits("inf", 5, 100)},
			"a2": {ID: "a2", OwnerID: "player", RegionID: "gelibolu", Units: repeatedUnits("inf", 5, 100)},
			"a3": {ID: "a3", OwnerID: "player", RegionID: "gelibolu", Units: repeatedUnits("inf", 5, 100)},
			"a4": {ID: "a4", OwnerID: "player", RegionID: "gelibolu", Units: repeatedUnits("inf", 5, 100)},
		},
		UnitTypes: map[string]*army.UnitType{
			"inf": {ID: "inf", GrainUpkeep: 2},
		},
	}

	report1 := applyEconomyTick(gs)
	firstStatus := gs.ArmyLogistics["a1"]
	if firstStatus.TotalHPDamage == 0 || firstStatus.DamagePerUnit == 0 {
		t.Fatal("ilk tur bölgesel ikmal zayiatı bekleniyordu")
	}
	if gs.Armies["a1"].OverCapacityTurns != 1 {
		t.Fatalf("ilk tur over-capacity sayacı 1 olmalı, got=%d", gs.Armies["a1"].OverCapacityTurns)
	}
	if len(report1.PlayerLogisticsAlerts) != 1 {
		t.Fatalf("oyuncu için tek lojistik uyarısı bekleniyordu, got=%d", len(report1.PlayerLogisticsAlerts))
	}

	report2 := applyEconomyTick(gs)
	secondStatus := gs.ArmyLogistics["a1"]
	if secondStatus.DamagePerUnit <= firstStatus.DamagePerUnit {
		t.Fatalf("uzun süreli yığılmada zayiat artmalı, first=%d second=%d", firstStatus.DamagePerUnit, secondStatus.DamagePerUnit)
	}
	if gs.Armies["a1"].OverCapacityTurns != 2 {
		t.Fatalf("ikinci tur over-capacity sayacı 2 olmalı, got=%d", gs.Armies["a1"].OverCapacityTurns)
	}
	if len(report2.PlayerLogisticsAlerts) != 1 || report2.PlayerLogisticsAlerts[0].Overload <= 0 {
		t.Fatalf("ikinci tur da oyuncu lojistik uyarısı sürmeliydi, got=%+v", report2.PlayerLogisticsAlerts)
	}
}

func TestRegionalLogisticsUsesProductionAfterCivilianDemand(t *testing.T) {
	gs := &state.GameState{
		Month:           4,
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", Grain: 0},
		},
		Regions: map[world.RegionID]*world.Region{
			"home": {
				ID: "home", OwnerID: "player", Population: 100,
				BaseGrainOutput: 20,
			},
		},
		Armies: map[army.ArmyID]*army.Army{
			"field": {
				ID: "field", OwnerID: "player", RegionID: "home",
				Units: repeatedUnits("inf", 8, 100),
			},
		},
		UnitTypes: map[string]*army.UnitType{
			"inf": {ID: "inf", GrainUpkeep: 2},
		},
	}

	applyEconomyTick(gs)

	status := gs.RegionLogistics["home"]
	if status.LocalProduction != 15 {
		t.Fatalf("yerel askeri üretim sivil talep sonrası 15 olmalıydı, got=%+v", status)
	}
	if status.Demand != 16 || status.Overload != 1 {
		t.Fatalf("sivil talep sonrası bölgesel askeri açık hesaplanmalıydı, got=%+v", status)
	}
}

func TestRegionalLogisticsReducesPortBufferUnderBlockade(t *testing.T) {
	gs := &state.GameState{
		Month:           4,
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", Grain: 20},
			"enemy":  {ID: "enemy"},
		},
		Regions: map[world.RegionID]*world.Region{
			"port": {ID: "port", OwnerID: "player", Neighbors: []world.RegionID{"sea"}, Settlements: []world.Settlement{{Type: world.SettlementPort}}},
			"sea":  {ID: "sea", IsSea: true},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("enemy", "player"): {FactionA: "enemy", FactionB: "player", Stance: faction.StanceWar},
		},
		Armies: map[army.ArmyID]*army.Army{
			"field": {ID: "field", OwnerID: "player", RegionID: "port", Units: repeatedUnits("inf", 1, 100)},
			"fleet": {ID: "fleet", OwnerID: "enemy", RegionID: "sea", IsNaval: true, Units: []army.Unit{{TypeID: "warship", CurrentHP: army.MaxUnitHP}}},
		},
		UnitTypes: map[string]*army.UnitType{
			"inf":     {ID: "inf", GrainUpkeep: 2},
			"warship": {ID: "warship", Category: army.CategoryNavalWar},
		},
	}

	applyEconomyTick(gs)

	status := gs.RegionLogistics["port"]
	if status.BlockadePercent != 50 || status.SettlementBuffer != 3 {
		t.Fatalf("liman tamponu abluka ile yarıya düşmeliydi, got=%+v", status)
	}
}

func TestApplyEconomyTickRegionalLogisticsResetsWhenCapacityRecovers(t *testing.T) {
	gs := &state.GameState{
		Month:           4,
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", Grain: 20},
		},
		Regions: map[world.RegionID]*world.Region{
			"safe": {ID: "safe", OwnerID: "player", TaxRate: 50, Satisfaction: 50, BaseGrainOutput: 12},
		},
		Armies: map[army.ArmyID]*army.Army{
			"a1": {ID: "a1", OwnerID: "player", RegionID: "safe", Units: repeatedUnits("inf", 1, 100), OverCapacityTurns: 3},
		},
		UnitTypes: map[string]*army.UnitType{
			"inf": {ID: "inf", GrainUpkeep: 2},
		},
	}

	report := applyEconomyTick(gs)

	if gs.Armies["a1"].OverCapacityTurns != 0 {
		t.Fatalf("kapasite yeterliyken sayaç sıfırlanmalı, got=%d", gs.Armies["a1"].OverCapacityTurns)
	}
	if status := gs.RegionLogistics["safe"]; status.Overload > 0 {
		t.Fatalf("güvenli bölgede overload olmamalı, got=%+v", status)
	}
	if _, ok := gs.ArmyLogistics["a1"]; ok {
		t.Fatal("kapasite yeterliyken orduya zayiat kaydı yazılmamalı")
	}
	if len(report.PlayerLogisticsAlerts) != 0 {
		t.Fatalf("kapasite yeterliyken oyuncu uyarısı olmamalı, got=%+v", report.PlayerLogisticsAlerts)
	}
}

func TestCheckEliminationsRemovesArmiesAndRelations(t *testing.T) {
	gs := &state.GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"a": {ID: "a"},
			"b": {ID: "b"},
			"c": {ID: "c"},
		},
		Regions: map[world.RegionID]*world.Region{
			"b1": {ID: "b1", OwnerID: "b"},
			"c1": {ID: "c1", OwnerID: "c"},
		},
		Armies: map[army.ArmyID]*army.Army{
			"a1": {ID: "a1", OwnerID: "a"},
			"b1": {ID: "b1", OwnerID: "b"},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("a", "b"): {FactionA: "a", FactionB: "b", Stance: faction.StanceWar},
			faction.RelationKey("a", "c"): {FactionA: "a", FactionB: "c", Stance: faction.StancePeace},
			faction.RelationKey("b", "c"): {FactionA: "b", FactionB: "c", Stance: faction.StanceAllied},
		},
		DiplomaticOffers: []state.DiplomaticOffer{
			{FromFactionID: "a", ToFactionID: "b", Action: "propose_peace"},
			{FromFactionID: "b", ToFactionID: "c", Action: "propose_trade"},
		},
		TradeRoutes: []*economy.TradeRoute{
			{FromFactionID: "a", ToFactionID: "b"},
			{FromFactionID: "b", ToFactionID: "c"},
		},
		AIPlans: map[faction.FactionID]*state.AIPlanState{
			"a": {ObjectiveID: "expand:b", Kind: state.AIObjectiveExpand, TargetFactionID: "b"},
			"b": {ObjectiveID: "defend:c", Kind: state.AIObjectiveDefend, TargetFactionID: "c"},
		},
	}

	checkEliminations(gs)

	if !gs.Factions["a"].IsEliminated {
		t.Fatal("bölgesi kalmayan fraksiyon elenmiş işaretlenmeli")
	}
	if _, ok := gs.Armies["a1"]; ok {
		t.Fatal("elenen fraksiyonun ordusu temizlenmeliydi")
	}
	if _, ok := gs.Relations[faction.RelationKey("a", "b")]; ok {
		t.Fatal("elenen fraksiyonun ilişkileri temizlenmeliydi (a|b)")
	}
	if _, ok := gs.Relations[faction.RelationKey("a", "c")]; ok {
		t.Fatal("elenen fraksiyonun ilişkileri temizlenmeliydi (a|c)")
	}
	if _, ok := gs.Relations[faction.RelationKey("b", "c")]; !ok {
		t.Fatal("elenmeyen fraksiyonlar arası ilişki korunmalı")
	}
	if len(gs.DiplomaticOffers) != 1 || gs.DiplomaticOffers[0].FromFactionID != "b" {
		t.Fatalf("elenen fraksiyonun teklifleri temizlenmeliydi, got=%+v", gs.DiplomaticOffers)
	}
	if len(gs.TradeRoutes) != 1 || gs.TradeRoutes[0].FromFactionID != "b" || gs.TradeRoutes[0].ToFactionID != "c" {
		t.Fatalf("elenen fraksiyonun ticaret rotalari temizlenmeliydi, got=%+v", gs.TradeRoutes)
	}
	if gs.AIPlans["a"] != nil || gs.AIPlans["b"] == nil {
		t.Fatalf("elenen fraksiyonun AI planı temizlenmeli, diğer planlar korunmalı: %+v", gs.AIPlans)
	}
}

func TestCheckEliminationsRemovesSeaOnlyFactionWithFleets(t *testing.T) {
	gs := &state.GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"a": {ID: "a"},
			"b": {ID: "b"},
		},
		Regions: map[world.RegionID]*world.Region{
			"sea_a":  {ID: "sea_a", OwnerID: "a", IsSea: true},
			"land_b": {ID: "land_b", OwnerID: "b", IsSea: false},
		},
		Armies: map[army.ArmyID]*army.Army{
			"a_fleet": {ID: "a_fleet", OwnerID: "a", IsNaval: true},
			"a_land":  {ID: "a_land", OwnerID: "a", IsNaval: false},
			"b_land":  {ID: "b_land", OwnerID: "b", IsNaval: false},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("a", "b"): {FactionA: "a", FactionB: "b", Stance: faction.StanceWar},
		},
	}

	checkEliminations(gs)

	if !gs.Factions["a"].IsEliminated {
		t.Fatal("sadece deniz bölgesi kalan fraksiyon elenmeli")
	}
	if _, ok := gs.Armies["a_fleet"]; ok {
		t.Fatal("elenen fraksiyonun donanması temizlenmeliydi")
	}
	if _, ok := gs.Armies["a_land"]; ok {
		t.Fatal("elenen fraksiyonun kara ordusu temizlenmeliydi")
	}
	if _, ok := gs.Relations[faction.RelationKey("a", "b")]; ok {
		t.Fatal("elenen fraksiyonun diplomasi ilişkileri temizlenmeliydi")
	}
}

func TestApplySeasonEffectsReplenishesFriendlyLandArmy(t *testing.T) {
	gs := &state.GameState{
		Month:           4,
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player"},
			"ally":   {ID: "ally"},
		},
		Regions: map[world.RegionID]*world.Region{
			"home":      {ID: "home", OwnerID: "player"},
			"enemy":     {ID: "enemy", OwnerID: "enemy"},
			"ally_port": {ID: "ally_port", OwnerID: "ally"},
			"sea":       {ID: "sea", IsSea: true},
		},
		Armies: map[army.ArmyID]*army.Army{
			"home_army":  {ID: "home_army", OwnerID: "player", RegionID: "home", Units: []army.Unit{{TypeID: "inf", CurrentHP: 65}}},
			"enemy_army": {ID: "enemy_army", OwnerID: "player", RegionID: "enemy", Units: []army.Unit{{TypeID: "inf", CurrentHP: 65}}},
			"fleet":      {ID: "fleet", OwnerID: "player", RegionID: "sea", IsNaval: true, Units: []army.Unit{{TypeID: "transport", CurrentHP: 65}}},
			"docked_self": {
				ID:             "docked_self",
				OwnerID:        "player",
				RegionID:       "sea",
				DockedRegionID: "home",
				IsNaval:        true,
				Units:          []army.Unit{{TypeID: "transport", CurrentHP: 65}},
				EmbarkedUnits:  []army.Unit{{TypeID: "inf", CurrentHP: 72}},
			},
			"docked_ally": {
				ID:             "docked_ally",
				OwnerID:        "player",
				RegionID:       "sea",
				DockedRegionID: "ally_port",
				IsNaval:        true,
				Units:          []army.Unit{{TypeID: "transport", CurrentHP: 65}},
				EmbarkedUnits:  []army.Unit{{TypeID: "inf", CurrentHP: 72}},
			},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("ally", "player"): {FactionA: "ally", FactionB: "player", Stance: faction.StanceAllied},
		},
	}

	applySeasonEffects(gs)

	if got := gs.Armies["home_army"].Units[0].CurrentHP; got != 75 {
		t.Fatalf("dost topraktaki ordu iyilesmeli, got=%d", got)
	}
	if got := gs.Armies["enemy_army"].Units[0].CurrentHP; got != 65 {
		t.Fatalf("dost olmayan toprakta iyilesme olmamali, got=%d", got)
	}
	if got := gs.Armies["fleet"].Units[0].CurrentHP; got != 65 {
		t.Fatalf("donanma kara takviyesi almamali, got=%d", got)
	}
	if got := gs.Armies["docked_self"].Units[0].CurrentHP; got != 75 {
		t.Fatalf("kendi limanındaki donanma iyilesmeli, got=%d", got)
	}
	if got := gs.Armies["docked_self"].EmbarkedUnits[0].CurrentHP; got != 82 {
		t.Fatalf("kendi limanındaki taşınan birlik de iyilesmeli, got=%d", got)
	}
	if got := gs.Armies["docked_ally"].Units[0].CurrentHP; got != 70 {
		t.Fatalf("müttefik limanındaki donanma yari hizda iyilesmeli, got=%d", got)
	}
	if got := gs.Armies["docked_ally"].EmbarkedUnits[0].CurrentHP; got != 77 {
		t.Fatalf("müttefik limanındaki taşınan birlik yari hizda iyilesmeli, got=%d", got)
	}
}

func TestApplyEmbarkedVoyageAttritionStartsAfterGraceTurns(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		Regions: map[world.RegionID]*world.Region{
			"sea": {ID: "sea", IsSea: true},
		},
		Armies: map[army.ArmyID]*army.Army{
			"fleet": {
				ID:               "fleet",
				OwnerID:          "player",
				RegionID:         "sea",
				IsNaval:          true,
				EmbarkedUnits:    repeatedUnits("inf", 2, 100),
				TurnsWithoutPort: 3,
			},
		},
		ArmyLogistics: map[army.ArmyID]state.ArmyLogisticsStatus{},
	}

	alerts := applyEmbarkedVoyageAttrition(gs)

	if len(alerts) != 1 {
		t.Fatalf("grace sonrasi tek deniz attrition uyarisi bekleniyordu, got=%d", len(alerts))
	}
	if got := gs.Armies["fleet"].TurnsWithoutPort; got != 4 {
		t.Fatalf("limansiz tur sayaci 4 olmali, got=%d", got)
	}
	for i, u := range gs.Armies["fleet"].EmbarkedUnits {
		if u.CurrentHP != 96 {
			t.Fatalf("embarked birim %d 4 HP kaybetmeliydi, got=%d", i, u.CurrentHP)
		}
	}
	if status := gs.ArmyLogistics["fleet"]; status.DamagePerUnit != 4 || status.TotalHPDamage != 8 {
		t.Fatalf("army logistics deniz attrition bilgisini tasimaliydi, got=%+v", status)
	}
}

func TestApplyEmbarkedVoyageAttritionResetsAtPort(t *testing.T) {
	gs := &state.GameState{
		Regions: map[world.RegionID]*world.Region{
			"sea":  {ID: "sea", IsSea: true},
			"home": {ID: "home", OwnerID: "player"},
		},
		Armies: map[army.ArmyID]*army.Army{
			"fleet": {
				ID:               "fleet",
				OwnerID:          "player",
				RegionID:         "sea",
				DockedRegionID:   "home",
				IsNaval:          true,
				EmbarkedUnits:    repeatedUnits("inf", 1, 92),
				TurnsWithoutPort: 5,
			},
		},
		ArmyLogistics: map[army.ArmyID]state.ArmyLogisticsStatus{},
	}

	alerts := applyEmbarkedVoyageAttrition(gs)

	if len(alerts) != 0 {
		t.Fatalf("limandaki filo attrition almamaliydi, got=%d", len(alerts))
	}
	if got := gs.Armies["fleet"].TurnsWithoutPort; got != 0 {
		t.Fatalf("limana baglaninca limansiz tur sayaci sifirlanmali, got=%d", got)
	}
	if got := gs.Armies["fleet"].EmbarkedUnits[0].CurrentHP; got != 92 {
		t.Fatalf("limandaki embarked birlik hasar almamaliydi, got=%d", got)
	}
	if _, ok := gs.ArmyLogistics["fleet"]; ok {
		t.Fatalf("limandaki filo icin deniz attrition kaydi yazilmamali")
	}
}
