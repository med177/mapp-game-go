package ai

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/city"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/scenario"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestAIForceRequirementsScaleLandReserveWithPopulationAndPlan(t *testing.T) {
	gs := &state.GameState{
		ScenarioID: "1300_ottoman_rise",
		Factions:   map[faction.FactionID]*faction.Faction{"ai": {ID: "ai"}},
		Regions: map[world.RegionID]*world.Region{
			"a": {ID: "a", OwnerID: "ai", Population: 800},
			"b": {ID: "b", OwnerID: "ai", Population: 400},
		},
		Armies: map[army.ArmyID]*army.Army{
			"field": {ID: "field", OwnerID: "ai", RegionID: "a", Units: []army.Unit{{TypeID: "inf", CurrentHP: army.MaxUnitHP}}},
		},
		UnitTypes: map[string]*army.UnitType{"inf": {ID: "inf", Category: army.CategoryInfantry}},
		AIPlans:   map[faction.FactionID]*state.AIPlanState{"ai": {Kind: state.AIObjectiveExpand}},
	}

	requirement := aiForceRequirements(gs, "ai", nil)
	// 1.200 / 200 = 6 temel birlik, genişleme planı bunu %125 ile 8'e çıkarır.
	if requirement.LandTarget != 8 || requirement.LandPresent != 1 {
		t.Fatalf("nüfus/genişleme kara rezerv hedefi yanlış: %+v", requirement)
	}
	if got := aiLandReserveShortfall(gs, "ai", nil); got != 7 {
		t.Fatalf("beklenen kara rezerv açığı 7 olmalıydı, got=%d", got)
	}
}

func TestAIExpansionReserveExceedsHistoricalTargetProjectedLandPower(t *testing.T) {
	gs := &state.GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"france": {ID: "france", Research: faction.ResearchState{Completed: map[string]bool{}}},
			"hre":    {ID: "hre"},
		},
		Regions: map[world.RegionID]*world.Region{
			"paris": {ID: "paris", OwnerID: "france", Population: 1000, Satisfaction: 100, BaseGrainOutput: 100, Buildings: []string{"barracks"}},
			"reims": {ID: "reims", OwnerID: "france", Population: 1000, Satisfaction: 100, BaseGrainOutput: 100, Buildings: []string{"barracks"}},
			"hre_a": {ID: "hre_a", OwnerID: "hre", Population: 500},
			"hre_b": {ID: "hre_b", OwnerID: "hre", Population: 500},
		},
		Armies: map[army.ArmyID]*army.Army{
			"hre_field": {ID: "hre_field", OwnerID: "hre", RegionID: "hre_a", Units: []army.Unit{{TypeID: "enemy", CurrentHP: army.MaxUnitHP}, {TypeID: "enemy", CurrentHP: army.MaxUnitHP}, {TypeID: "enemy", CurrentHP: army.MaxUnitHP}, {TypeID: "enemy", CurrentHP: army.MaxUnitHP}}},
		},
		UnitTypes: map[string]*army.UnitType{
			"infantry": {ID: "infantry", Category: army.CategoryInfantry, Attack: 20, Morale: 50, RequiredBldg: "barracks", RequiredBldgLevel: 1},
			"enemy":    {ID: "enemy", Category: army.CategoryInfantry, Attack: 30, Morale: 50},
		},
		AIPlans: map[faction.FactionID]*state.AIPlanState{"france": {Kind: state.AIObjectiveExpand, TargetFactionID: "hre"}},
	}
	for i := 0; i < 8; i++ {
		id := world.RegionID("france_reserve_" + string(rune('a'+i)))
		gs.Regions[id] = &world.Region{ID: id, OwnerID: "france"}
	}

	requirement := aiForceRequirements(gs, "france", nil)
	// HRE'nin dört birimi 140 güçtedir; Fransa'nın genişleme hedefi %135 ile
	// en az 189 güç ister. 25 güçlük piyade ile bu sekiz birlik eder; nüfus
	// tabanı ise 13 birlik verdiği için ilk aşamada nüfus tabanı korunur.
	if requirement.ObjectiveEnemyLandPower != 140 || requirement.ObjectiveLandPowerTarget != 189 {
		t.Fatalf("hedef devletin kara gücü yanlış türetildi: %+v", requirement)
	}
	if requirement.LandTarget < 13 {
		t.Fatalf("tarihsel hedef gücü nüfus tabanını asla düşürmemeli: %+v", requirement)
	}

	// Daha güçlü bir HRE filosu değil, kara ordusu Fransa'nın kara hedefini
	// büyütmelidir; 16 ek 35-güçlük birim hedefi 35'e taşır.
	for i := 0; i < 16; i++ {
		gs.Armies[army.ArmyID("hre_extra_"+string(rune('a'+i)))] = &army.Army{ID: army.ArmyID("hre_extra_" + string(rune('a'+i))), OwnerID: "hre", RegionID: "hre_b", Units: []army.Unit{{TypeID: "enemy", CurrentHP: army.MaxUnitHP}}}
	}
	requirement = aiForceRequirements(gs, "france", nil)
	if requirement.LandTarget <= 13 || requirement.ObjectiveEnemyLandPower <= 140 {
		t.Fatalf("güçlenen tarihsel hedef Fransa'nın birlik hedefini artırmalıydı: %+v", requirement)
	}
}

func TestAIReserveRecruitmentQueuesOnlyPopulationBasedShortfall(t *testing.T) {
	gs := &state.GameState{
		ScenarioID: "1300_ottoman_rise",
		Factions:   map[faction.FactionID]*faction.Faction{"ai": {ID: "ai", Gold: 1000, Grain: 1000, Research: faction.ResearchState{Completed: map[string]bool{}}}},
		Regions: map[world.RegionID]*world.Region{
			"home": {ID: "home", OwnerID: "ai", Population: 800, Satisfaction: 100, BaseGrainOutput: 100, Buildings: []string{"barracks"}},
		},
		Armies: map[army.ArmyID]*army.Army{},
		UnitTypes: map[string]*army.UnitType{
			"inf": {ID: "inf", Category: army.CategoryInfantry, GoldCost: 80, GrainCost: 10, GrainUpkeep: 2, RequiredBldg: "barracks", RequiredBldgLevel: 1, TurnsRequired: 2},
		},
		AIPlans: map[faction.FactionID]*state.AIPlanState{"ai": {Kind: state.AIObjectiveConsolidate}},
	}

	aiRecruitAndBuildWithStrategicContextAndSteps(gs, "ai", nil, nil, nil)
	if got := aiPendingLandUnitCount(gs, "ai"); got != 1 {
		t.Fatalf("tek-seviyeli kışla hattında ilk rezerv emri bekleniyordu, got=%d (%+v)", got, gs.ProductionQueue)
	}
	if got := aiLandReserveShortfall(gs, "ai", nil); got != 3 {
		t.Fatalf("üretim hattı açıldıktan sonra kalan rezerv hedefi korunmalıydı, got=%d", got)
	}
}

func TestAINavalFocusUsesScenarioDoctrineInsteadOfHardcodedFactionIDs(t *testing.T) {
	gs := &state.GameState{
		Factions: map[faction.FactionID]*faction.Faction{"maritime": {ID: "maritime"}},
		Regions: map[world.RegionID]*world.Region{
			"c1": {ID: "c1", OwnerID: "maritime", Neighbors: []world.RegionID{"s1"}},
			"c2": {ID: "c2", OwnerID: "maritime", Neighbors: []world.RegionID{"s2"}},
			"c3": {ID: "c3", OwnerID: "maritime", Neighbors: []world.RegionID{"s3"}},
			"c4": {ID: "c4", OwnerID: "maritime", Neighbors: []world.RegionID{"s4"}},
			"s1": {ID: "s1", IsSea: true}, "s2": {ID: "s2", IsSea: true},
			"s3": {ID: "s3", IsSea: true}, "s4": {ID: "s4", IsSea: true},
		},
		AIStrategies: map[string]scenario.AIFactionStrategy{"maritime": {FactionID: "maritime", NavalFocus: true}},
	}
	if got := aiForceRequirements(gs, "maritime", nil).WarshipTarget; got != 8 {
		t.Fatalf("JSON naval_focus dört kıyı için sekiz savaş gemisi hedeflemeliydi, got=%d", got)
	}
}

func TestAINavalReserveBuildsPortBeforeWarship(t *testing.T) {
	gs := &state.GameState{
		ScenarioID: "1300_ottoman_rise",
		Factions:   map[faction.FactionID]*faction.Faction{"ai": {ID: "ai", Gold: 5000, Grain: 500, Iron: 500, Timber: 500, Research: faction.ResearchState{Completed: map[string]bool{"navigation": true, "naval_doctrine": true}}}},
		Regions: map[world.RegionID]*world.Region{
			"coast": {ID: "coast", OwnerID: "ai", Population: 1000, Neighbors: []world.RegionID{"sea"}},
			"sea":   {ID: "sea", IsSea: true, Neighbors: []world.RegionID{"coast"}},
		},
		Armies: map[army.ArmyID]*army.Army{},
		UnitTypes: map[string]*army.UnitType{
			"warship": {ID: "warship", Category: army.CategoryNavalWar, GoldCost: 400, GrainCost: 8, IronCost: 12, TimberCost: 40, RequiredBldg: "port", RequiredBldgLevel: 3, TurnsRequired: 4, RequiredTech: []string{"navigation", "naval_doctrine"}},
		},
		BuildingTypes: map[string]*city.Building{"port": {ID: "port", GoldCost: 120, TimberCost: 20, MaxPerRegion: 3, TurnsRequired: 2}},
	}

	aiProduceNavalReserve(gs, "ai", nil, nil, nil)
	if len(gs.ProductionQueue) != 1 || gs.ProductionQueue[0].Kind != aiProductionKindBuilding || gs.ProductionQueue[0].TypeID != "port" {
		t.Fatalf("savaş gemisinden önce eksik liman seviyesi kuyruğa alınmalıydı: %+v", gs.ProductionQueue)
	}

	gs.ProductionQueue = nil
	gs.Regions["coast"].Buildings = []string{"port", "port", "port"}
	aiProduceNavalReserve(gs, "ai", nil, nil, nil)
	if len(gs.ProductionQueue) != 1 || gs.ProductionQueue[0].Kind != aiProductionKindUnit || gs.ProductionQueue[0].TypeID != "warship" {
		t.Fatalf("liman hazır olduğunda savaş gemisi kuyruğa alınmalıydı: %+v", gs.ProductionQueue)
	}
}
