package state

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/world"
)

func TestConvertInvalidNavalBlockadesToPatrol(t *testing.T) {
	gs := navalMissionTransitionState()
	blockade := gs.Armies["blockade"]

	if got := gs.ConvertInvalidNavalBlockadesToPatrol(); got != 0 {
		t.Fatalf("geçerli abluka için dönüşüm sayısı = %d, want 0", got)
	}
	if blockade.NavalMission.Kind != army.NavalMissionBlockade {
		t.Fatalf("geçerli abluka görevi değişmemeli, got %q", blockade.NavalMission.Kind)
	}

	gs.Regions["enemy_port"].OwnerID = "player"
	if got := gs.ConvertInvalidNavalBlockadesToPatrol(); got != 1 {
		t.Fatalf("geçersiz abluka için dönüşüm sayısı = %d, want 1", got)
	}
	if blockade.NavalMission.Kind != army.NavalMissionPatrol {
		t.Fatalf("geçersiz abluka devriyeye dönüşmeli, got %q", blockade.NavalMission.Kind)
	}
	if blockade.NavalMission.TargetRegionID != "sea" {
		t.Fatalf("devriye hedefi mevcut deniz olmalı, got %q", blockade.NavalMission.TargetRegionID)
	}
}

func TestNavalFishingMissionRequiresOwnedCoastAndProducesPerUnitGrain(t *testing.T) {
	gs := &GameState{
		UnitTypes: map[string]*army.UnitType{
			"fishing_ship": {ID: "fishing_ship", Category: army.CategoryNavalTrade, FishingGrainIncome: 30},
			"transport":    {ID: "transport", Category: army.CategoryNavalTrans},
		},
		Regions: map[world.RegionID]*world.Region{
			"coast":     {ID: "coast", OwnerID: "player", Neighbors: []world.RegionID{"sea"}},
			"sea":       {ID: "sea", IsSea: true, Neighbors: []world.RegionID{"coast"}},
			"far_coast": {ID: "far_coast", OwnerID: "player"},
		},
		Armies: map[army.ArmyID]*army.Army{
			"fleet": {
				ID: "fleet", OwnerID: "player", IsNaval: true, RegionID: "sea",
				Units: []army.Unit{
					{TypeID: "fishing_ship", CurrentHP: 100},
					{TypeID: "fishing_ship", CurrentHP: 100},
					{TypeID: "fishing_ship", CurrentHP: 0},
					{TypeID: "transport", CurrentHP: 100},
				},
			},
		},
	}
	mission := army.NavalMission{Kind: army.NavalMissionFishing, TargetRegionID: "sea"}
	if ok, reason := gs.CanAssignNavalMission("fleet", mission); !ok {
		t.Fatalf("fishing mission rejected beside owned coast: %s", reason)
	}
	if got, want := gs.NavalFishingGrainIncomeForFleet(gs.Armies["fleet"]), 60; got != want {
		t.Fatalf("fleet fishing income = %d, want %d", got, want)
	}
	if got, want := gs.NavalFishingNetGrainIncomeForFleet(gs.Armies["fleet"]), 60; got != want {
		t.Fatalf("fleet net fishing income = %d, want %d", got, want)
	}

	fleet := gs.Armies["fleet"]
	fleet.NavalMission = &mission
	if got, want := gs.NavalFishingGrainIncomeForFaction("player"), 60; got != want {
		t.Fatalf("faction fishing income = %d, want %d", got, want)
	}

	gs.Regions["coast"].OwnerID = "other"
	if ok, _ := gs.CanAssignNavalMission("fleet", mission); ok {
		t.Fatal("fishing mission should be rejected without an adjacent owned coast")
	}
	if got := gs.NavalFishingGrainIncomeForFaction("player"); got != 0 {
		t.Fatalf("fishing income without owned coast = %d, want 0", got)
	}
}

func TestNavalFishingNetIncomeSubtractsEachFishingUnitUpkeep(t *testing.T) {
	gs := &GameState{
		UnitTypes: map[string]*army.UnitType{
			"warship": {ID: "warship", Category: army.CategoryNavalWar, FishingGrainIncome: 40, GrainUpkeep: 20},
			"merchant_ship": {
				ID: "merchant_ship", Category: army.CategoryNavalTrade,
				FishingGrainIncome: 30, GrainUpkeep: 3,
			},
		},
	}
	fleet := &army.Army{
		IsNaval: true,
		Units: []army.Unit{
			{TypeID: "warship", CurrentHP: army.MaxUnitHP},
			{TypeID: "warship", CurrentHP: army.MaxUnitHP},
			{TypeID: "warship", CurrentHP: army.MaxUnitHP},
			{TypeID: "warship", CurrentHP: army.MaxUnitHP},
			{TypeID: "warship", CurrentHP: army.MaxUnitHP},
			{TypeID: "warship", CurrentHP: army.MaxUnitHP},
			{TypeID: "merchant_ship", CurrentHP: army.MaxUnitHP},
			{TypeID: "warship"},
		},
	}
	if got, want := gs.NavalFishingNetGrainIncomeForFleet(fleet), 147; got != want {
		t.Fatalf("fleet net fishing income = %d, want %d (six warships, merchant ship, excluding sunk ship)", got, want)
	}
	fleet.MaxMovePoints = 3
	fleet.MovePoints = 2
	if got, want := gs.NavalFishingNetGrainIncomeForFleet(fleet), 86; got != want {
		t.Fatalf("moving fleet net fishing income = %d, want %d after 150%% upkeep", got, want)
	}
}

func TestFactionGrainPreviewIncludesFishingAndEffectiveArmyUpkeep(t *testing.T) {
	const fid faction.FactionID = "player"
	fleet := &army.Army{
		ID: "fleet", OwnerID: string(fid), IsNaval: true, RegionID: "sea",
		MaxMovePoints: 3, MovePoints: 3,
		NavalMission: &army.NavalMission{Kind: army.NavalMissionFishing, TargetRegionID: "sea"},
		Units: []army.Unit{
			{TypeID: "warship", CurrentHP: army.MaxUnitHP},
			{TypeID: "warship", CurrentHP: army.MaxUnitHP},
			{TypeID: "warship", CurrentHP: army.MaxUnitHP},
			{TypeID: "warship", CurrentHP: army.MaxUnitHP},
			{TypeID: "warship", CurrentHP: army.MaxUnitHP},
			{TypeID: "warship", CurrentHP: army.MaxUnitHP},
		},
	}
	gs := &GameState{
		PlayerFactionID: fid,
		UnitTypes: map[string]*army.UnitType{
			"warship": {ID: "warship", Category: army.CategoryNavalWar, FishingGrainIncome: 40, GrainUpkeep: 20},
		},
		Regions: map[world.RegionID]*world.Region{
			"coast": {ID: "coast", OwnerID: string(fid), Neighbors: []world.RegionID{"sea"}},
			"sea":   {ID: "sea", IsSea: true, Neighbors: []world.RegionID{"coast"}},
		},
		Armies: map[army.ArmyID]*army.Army{"fleet": fleet},
	}
	if got, want := gs.FactionGrainNetChange(fid), 120; got != want {
		t.Fatalf("stationary six-ship net grain change = %d, want %d", got, want)
	}
	fleet.MovePoints = 2
	if got, want := gs.FactionGrainNetChange(fid), 60; got != want {
		t.Fatalf("moving six-ship net grain change = %d, want %d", got, want)
	}
}

func TestFishingMissionClearsAfterRelocation(t *testing.T) {
	fleet := &army.Army{
		ID: "fleet", IsNaval: true, RegionID: "other_sea",
		NavalMission: &army.NavalMission{Kind: army.NavalMissionFishing, TargetRegionID: "sea"},
	}
	gs := &GameState{}
	if !gs.ClearNavalMissionAfterRelocation(fleet, "sea") {
		t.Fatal("fishing mission should clear when fleet leaves its assigned sea")
	}
	if fleet.NavalMission != nil {
		t.Fatal("fishing mission should be removed after relocation")
	}
}

func TestSupplyCargoLoadsAssignsAndUnloadsAtCapital(t *testing.T) {
	gs := &GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", CapitalSettlementID: "capital_port", Grain: 100},
		},
		UnitTypes: map[string]*army.UnitType{
			"transport": {ID: "transport", Category: army.CategoryNavalTrans, CarryCapacity: 5},
			"soldier":   {ID: "soldier", Category: army.CategoryInfantry, GrainUpkeep: 10},
		},
		Regions: map[world.RegionID]*world.Region{
			"capital": {ID: "capital", OwnerID: "player", Neighbors: []world.RegionID{"sea"}, Settlements: []world.Settlement{{ID: "capital_port", Type: world.SettlementPort}}},
			"sea":     {ID: "sea", IsSea: true, Neighbors: []world.RegionID{"capital", "coast"}},
			"coast":   {ID: "coast", OwnerID: "player", Neighbors: []world.RegionID{"sea"}, Settlements: []world.Settlement{{ID: "coast_port", Type: world.SettlementPort}}},
		},
		Armies: map[army.ArmyID]*army.Army{
			"fleet": {ID: "fleet", OwnerID: "player", IsNaval: true, DockedRegionID: "capital", RegionID: "sea", Units: []army.Unit{{TypeID: "transport"}}},
			"army":  {ID: "army", OwnerID: "player", RegionID: "coast", Units: []army.Unit{{TypeID: "soldier", CurrentHP: 100}}},
		},
	}
	if got := gs.SupplyCargoCapacityForTurns("fleet", 5); got != 250 {
		t.Fatalf("5 turluk ikmal kapasitesi = %d, want 250", got)
	}

	ok, reason := gs.LoadSupplyCargoAtCapital("fleet", economy.ResourceCost{Grain: 20})
	if !ok || reason != "" {
		t.Fatalf("yükleme sonucu = (%v, %q)", ok, reason)
	}
	if got := gs.Factions["player"].Grain; got != 80 {
		t.Fatalf("yükleme sonrası tahıl = %d, want 80", got)
	}

	fleet := gs.Armies["fleet"]
	fleet.DockedRegionID = ""
	fleet.NavalMission = &army.NavalMission{Kind: army.NavalMissionSupplyArmy, TargetArmyID: "army"}
	if ok, reason := gs.CanAssignNavalMission("fleet", *fleet.NavalMission); !ok {
		t.Fatalf("ikmal görevi reddedildi: %s", reason)
	}

	fleet.DockedRegionID = "capital"
	if cargo, ok, reason := gs.UnloadSupplyCargoAtCapital("fleet"); !ok || cargo.Grain != 20 || reason != "" {
		t.Fatalf("boşaltma sonucu = (%+v, %v, %q)", cargo, ok, reason)
	}
	if fleet.NavalMission != nil {
		t.Fatal("ikmal yükü boşaltılınca ordu bağı temizlenmeli")
	}
	if got := gs.Factions["player"].Grain; got != 100 {
		t.Fatalf("boşaltma sonrası tahıl = %d, want 100", got)
	}
}

func TestSupplyCargoCapacityForUnitTypeIsPerUnit(t *testing.T) {
	units := make([]army.Unit, 16)
	for i := range units {
		units[i] = army.Unit{TypeID: "transport"}
	}
	gs := &GameState{
		UnitTypes: map[string]*army.UnitType{
			"transport": {ID: "transport", Category: army.CategoryNavalTrans, CarryCapacity: 4},
			"dragon":    {ID: "dragon", Category: army.CategoryDragon, GrainUpkeep: 25},
		},
		Armies: map[army.ArmyID]*army.Army{
			"fleet": {ID: "fleet", IsNaval: true, Units: units},
		},
	}

	if got := gs.SupplyCargoCapacityForUnitType("transport", 5); got != 500 {
		t.Fatalf("tek nakliye gemisinin 5 turluk ikmal kapasitesi = %d, want 500", got)
	}
	if got := gs.SupplyCargoCapacityForTurns("fleet", 5); got != 2500 {
		t.Fatalf("filonun 5 turluk ikmal kapasitesi = %d, want 2500", got)
	}
}

func TestLoadSupplyCargoForTurnsAddsToExistingCargoAndLeavesOverflowInStock(t *testing.T) {
	gs := &GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", CapitalSettlementID: "capital_port", Grain: 200},
		},
		UnitTypes: map[string]*army.UnitType{
			"transport": {ID: "transport", Category: army.CategoryNavalTrans, CarryCapacity: 5},
			"soldier":   {ID: "soldier", Category: army.CategoryInfantry, GrainUpkeep: 10},
		},
		Regions: map[world.RegionID]*world.Region{
			"capital": {ID: "capital", OwnerID: "player", Settlements: []world.Settlement{{ID: "capital_port", Type: world.SettlementPort}}},
		},
		Armies: map[army.ArmyID]*army.Army{
			"fleet": {
				ID: "fleet", OwnerID: "player", IsNaval: true, DockedRegionID: "capital",
				Units:       []army.Unit{{TypeID: "transport"}},
				SupplyCargo: economy.ResourceCost{Grain: 80},
			},
		},
	}

	cargo, ok, reason := gs.LoadSupplyCargoForTurns("fleet", 5)
	if !ok || reason != "" {
		t.Fatalf("mevcut yüke ek yükleme sonucu = (%+v, %v, %q)", cargo, ok, reason)
	}
	if cargo.Grain != 170 {
		t.Fatalf("yüklenen tahıl = %d, want 170", cargo.Grain)
	}
	if got := gs.Armies["fleet"].SupplyCargo.Grain; got != 250 {
		t.Fatalf("filonun toplam ikmal yükü = %d, want 250", got)
	}
	if got := gs.Factions["player"].Grain; got != 30 {
		t.Fatalf("merkez ambarındaki tahıl = %d, want 30", got)
	}
}

func TestLoadSupplyCargoForTurnsLoadsOnlyAvailableStock(t *testing.T) {
	gs := &GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", CapitalSettlementID: "capital_port", Grain: 40},
		},
		UnitTypes: map[string]*army.UnitType{
			"transport": {ID: "transport", Category: army.CategoryNavalTrans, CarryCapacity: 5},
			"soldier":   {ID: "soldier", Category: army.CategoryInfantry, GrainUpkeep: 10},
		},
		Regions: map[world.RegionID]*world.Region{
			"capital": {ID: "capital", OwnerID: "player", Settlements: []world.Settlement{{ID: "capital_port", Type: world.SettlementPort}}},
		},
		Armies: map[army.ArmyID]*army.Army{
			"fleet": {
				ID: "fleet", OwnerID: "player", IsNaval: true, DockedRegionID: "capital",
				Units:       []army.Unit{{TypeID: "transport"}},
				SupplyCargo: economy.ResourceCost{Grain: 80},
			},
		},
	}

	cargo, ok, reason := gs.LoadSupplyCargoForTurns("fleet", 5)
	if !ok || reason != "" {
		t.Fatalf("stok miktarına kadar yükleme sonucu = (%+v, %v, %q)", cargo, ok, reason)
	}
	if cargo.Grain != 40 || gs.Armies["fleet"].SupplyCargo.Grain != 120 || gs.Factions["player"].Grain != 0 {
		t.Fatalf("stoktan fazla yükleme yapıldı: cargo=%+v fleet=%+v stock=%d", cargo, gs.Armies["fleet"].SupplyCargo, gs.Factions["player"].Grain)
	}
}

func TestSupplyCargoUsesNearestOwnedPortWhenCapitalRegionHasNoPort(t *testing.T) {
	gs := &GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", CapitalSettlementID: "capital_city", Grain: 100},
		},
		UnitTypes: map[string]*army.UnitType{
			"transport": {ID: "transport", Category: army.CategoryNavalTrans, CarryCapacity: 5},
			"soldier":   {ID: "soldier", Category: army.CategoryInfantry, GrainUpkeep: 10},
		},
		Regions: map[world.RegionID]*world.Region{
			"capital":      {ID: "capital", OwnerID: "player", WorldX: 0, WorldY: 0, Settlements: []world.Settlement{{ID: "capital_city", Type: world.SettlementCity}}},
			"central_port": {ID: "central_port", OwnerID: "player", WorldX: 1, WorldY: 1, Neighbors: []world.RegionID{"sea"}, Settlements: []world.Settlement{{ID: "central_port_settlement", Type: world.SettlementPort}}},
			"sea":          {ID: "sea", IsSea: true, Neighbors: []world.RegionID{"central_port"}},
		},
		Armies: map[army.ArmyID]*army.Army{
			"fleet": {ID: "fleet", OwnerID: "player", IsNaval: true, DockedRegionID: "central_port", RegionID: "sea", Units: []army.Unit{{TypeID: "transport"}}},
		},
	}

	if got := gs.SupplyCargoPortRegion("player"); got == nil || got.ID != "central_port" {
		t.Fatalf("canonical merkez limanı = %v, want central_port", got)
	}
	if ok, reason := gs.LoadSupplyCargoAtCapital("fleet", economy.ResourceCost{Grain: 20}); !ok || reason != "" {
		t.Fatalf("başkent dışındaki canonical merkez limanında yükleme reddedildi: %v, %q", ok, reason)
	}
	if got := gs.Armies["fleet"].SupplyCargo.Grain; got != 20 {
		t.Fatalf("canonical merkez limanı kargosu = %d, want 20", got)
	}
}

func TestConvertInvalidNavalBlockadesKeepsUnrelatedMissions(t *testing.T) {
	gs := navalMissionTransitionState()
	gs.Armies["patrol"] = &army.Army{
		ID:       "patrol",
		OwnerID:  "player",
		RegionID: "sea",
		IsNaval:  true,
		NavalMission: &army.NavalMission{
			Kind:           army.NavalMissionPatrol,
			TargetRegionID: "sea",
		},
	}
	gs.Regions["enemy_port"].OwnerID = "player"

	if got := gs.ConvertInvalidNavalBlockadesToPatrol(); got != 1 {
		t.Fatalf("dönüşüm sayısı = %d, want 1", got)
	}
	if got := gs.Armies["patrol"].NavalMission.Kind; got != army.NavalMissionPatrol {
		t.Fatalf("mevcut devriye görevi korunmalı, got %q", got)
	}
}

func navalMissionTransitionState() *GameState {
	return &GameState{
		Regions: map[world.RegionID]*world.Region{
			"enemy_port": {
				ID:          "enemy_port",
				OwnerID:     "enemy",
				Neighbors:   []world.RegionID{"sea"},
				Settlements: []world.Settlement{{ID: "enemy_harbor", Type: world.SettlementPort}},
			},
			"sea": {ID: "sea", IsSea: true, Neighbors: []world.RegionID{"enemy_port"}},
		},
		Armies: map[army.ArmyID]*army.Army{
			"blockade": {
				ID:       "blockade",
				OwnerID:  "player",
				RegionID: "sea",
				IsNaval:  true,
				NavalMission: &army.NavalMission{
					Kind:           army.NavalMissionBlockade,
					TargetRegionID: "sea",
				},
			},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("player", "enemy"): {
				FactionA: "player",
				FactionB: "enemy",
				Stance:   faction.StanceWar,
			},
		},
	}
}
