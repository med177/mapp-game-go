package ai

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/religion"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/tech"
	"mapp-game-go/internal/world"
)

func TestAIResearchExpandPrioritizesRequiredSiegeUnlock(t *testing.T) {
	gs, ctx := aiResearchTestState(state.AIObjectiveExpand)
	gs.Regions["target"].Buildings = []string{"walls"}
	gs.TechTypes = map[string]*tech.Technology{
		"siege_unlock": aiTestTechnology("siege_unlock", tech.CategoryMilitary, 25, 4, tech.Effects{SiegeAttackMod: 0.05}),
		"small_income": aiTestTechnology("small_income", tech.CategoryEconomy, 10, 2, tech.Effects{GoldPerRegion: 1}),
	}
	gs.UnitTypes["catapult"] = &army.UnitType{
		ID: "catapult", Category: army.CategorySiege, Attack: 30, Defense: 3, Morale: 30,
		GoldCost: 200, GrainUpkeep: 3, TurnsRequired: 3, RequiredTech: "siege_unlock",
		RequiredBldg: "barracks", RequiredBldgLevel: 1,
	}

	if got := aiSelectResearchTechnology(gs, gs.Factions["ai"], nil, ctx); got == nil || got.ID != "siege_unlock" {
		t.Fatalf("tahkimli genişleme ve eksik kuşatma desteği birim açılımını seçmeliydi: got=%v", researchID(got))
	}
}

func TestAIResearchDefendValuesRealLandDefenseEffect(t *testing.T) {
	gs, ctx := aiResearchTestState(state.AIObjectiveDefend)
	gs.TechTypes = map[string]*tech.Technology{
		"a_attack":  aiTestTechnology("a_attack", tech.CategoryMilitary, 20, 3, tech.Effects{InfantryAttackMod: 0.10}),
		"z_defense": aiTestTechnology("z_defense", tech.CategoryMilitary, 20, 3, tech.Effects{LandDefenseMod: 0.10}),
	}

	if got := aiSelectResearchTechnology(gs, gs.Factions["ai"], nil, ctx); got == nil || got.ID != "z_defense" {
		t.Fatalf("savunma planı eşit modda gerçek kara savunmasını seçmeliydi: got=%v", researchID(got))
	}
}

func TestAIResearchUsesGrainBottleneck(t *testing.T) {
	gs, ctx := aiResearchTestState(state.AIObjectiveConsolidate)
	gs.Factions["ai"].Grain = 5
	gs.Regions["home"].BaseGrainOutput = 20
	gs.UnitTypes["hungry"] = &army.UnitType{ID: "hungry", Category: army.CategoryInfantry, GrainUpkeep: 30}
	gs.Armies["field"].Units = []army.Unit{{TypeID: "hungry", CurrentHP: army.MaxUnitHP}}
	gs.TechTypes = map[string]*tech.Technology{
		"crop": aiTestTechnology("crop", tech.CategoryEconomy, 20, 3, tech.Effects{GrainMod: 0.20}),
		"tax":  aiTestTechnology("tax", tech.CategoryEconomy, 20, 3, tech.Effects{GoldPerRegion: 3}),
	}

	if got := aiSelectResearchTechnology(gs, gs.Factions["ai"], nil, ctx); got == nil || got.ID != "crop" {
		t.Fatalf("tahıl üretimi bakımın altındayken tahıl teknolojisi seçilmeliydi: got=%v", researchID(got))
	}
}

func TestAIResearchConsolidateUsesStabilityAndReligionNeed(t *testing.T) {
	gs, ctx := aiResearchTestState(state.AIObjectiveConsolidate)
	gs.Factions["ai"].Religion = religion.Sunni
	gs.Regions["home"].Religion = string(religion.Orthodox)
	gs.Regions["home"].Satisfaction = 20
	gs.TechTypes = map[string]*tech.Technology{
		"stability": aiTestTechnology("stability", tech.CategoryReligion, 20, 3, tech.Effects{SatisfactionBonus: 4, ConversionSpeedMod: 1}),
		"income":    aiTestTechnology("income", tech.CategoryEconomy, 20, 3, tech.Effects{GoldPerRegion: 3}),
	}

	if got := aiSelectResearchTechnology(gs, gs.Factions["ai"], nil, ctx); got == nil || got.ID != "stability" {
		t.Fatalf("istikrarsız farklı din bölgesinde memnuniyet/dönüşüm teknolojisi seçilmeliydi: got=%v", researchID(got))
	}
}

func TestAIResearchLandlockedFactionDeprioritizesNavalTech(t *testing.T) {
	gs, ctx := aiResearchTestState(state.AIObjectiveConsolidate)
	gs.TechTypes = map[string]*tech.Technology{
		"naval":    aiTestTechnology("naval", tech.CategoryNaval, 10, 2, tech.Effects{NavalMoveBonus: 1}),
		"military": aiTestTechnology("military", tech.CategoryMilitary, 10, 2, tech.Effects{}),
	}

	if got := aiSelectResearchTechnology(gs, gs.Factions["ai"], nil, ctx); got == nil || got.ID != "military" {
		t.Fatalf("kıyısız devlet saf deniz teknolojisini geri plana atmalıydı: got=%v", researchID(got))
	}
}

func TestAIResearchUnitUnlockFollowsCompositionDeficit(t *testing.T) {
	gs, ctx := aiResearchTestState(state.AIObjectiveExpand)
	gs.Armies["field"].Units = aiUnitsOfType("base_inf", 8)
	gs.TechTypes = map[string]*tech.Technology{
		"a_infantry_unlock": aiTestTechnology("a_infantry_unlock", tech.CategoryMilitary, 20, 3, tech.Effects{}),
		"z_cavalry_unlock":  aiTestTechnology("z_cavalry_unlock", tech.CategoryMilitary, 20, 3, tech.Effects{}),
	}
	gs.UnitTypes["new_inf"] = aiResearchUnlockUnit("new_inf", army.CategoryInfantry, "a_infantry_unlock")
	gs.UnitTypes["new_cav"] = aiResearchUnlockUnit("new_cav", army.CategoryCavalry, "z_cavalry_unlock")

	if got := aiSelectResearchTechnology(gs, gs.Factions["ai"], nil, ctx); got == nil || got.ID != "z_cavalry_unlock" {
		t.Fatalf("piyade dolu kompozisyonda süvari açılımı seçilmeliydi: got=%v", researchID(got))
	}
}

func TestAIResearchRewardsImmediatelyUnlockedFollowOnTech(t *testing.T) {
	gs, ctx := aiResearchTestState(state.AIObjectiveConsolidate)
	gs.TechTypes = map[string]*tech.Technology{
		"a_plain":      aiTestTechnology("a_plain", tech.CategoryEconomy, 20, 3, tech.Effects{}),
		"z_foundation": aiTestTechnology("z_foundation", tech.CategoryEconomy, 20, 3, tech.Effects{}),
		"future": {
			ID: "future", Category: tech.CategoryEconomy, GoldCost: 20, TurnsRequired: 3,
			Requires: []string{"z_foundation"}, Effects: tech.Effects{GoldPerRegion: 3},
		},
	}

	if got := aiSelectResearchTechnology(gs, gs.Factions["ai"], nil, ctx); got == nil || got.ID != "z_foundation" {
		t.Fatalf("eşit adayda doğrudan sonraki teknoloji açan temel seçilmeliydi: got=%v", researchID(got))
	}
}

func TestAIResearchIgnoresPlayerOnlyIntelValue(t *testing.T) {
	gs, ctx := aiResearchTestState(state.AIObjectiveConsolidate)
	gs.TechTypes = map[string]*tech.Technology{
		"intel": aiTestTechnology("intel", tech.CategoryDiplomacy, 20, 3, tech.Effects{RevealEnemyStrength: true}),
		"peace": aiTestTechnology("peace", tech.CategoryDiplomacy, 20, 3, tech.Effects{PeaceRelationBonus: 4}),
	}

	if got := aiSelectResearchTechnology(gs, gs.Factions["ai"], nil, ctx); got == nil || got.ID != "peace" {
		t.Fatalf("yalnız oyuncu UI'sına yarayan istihbarat AI için sahte değer üretmemeliydi: got=%v", researchID(got))
	}
}

func TestAIResearchKeepsActiveResearch(t *testing.T) {
	gs, ctx := aiResearchTestState(state.AIObjectiveExpand)
	gs.TechTypes = map[string]*tech.Technology{
		"current": aiTestTechnology("current", tech.CategoryEconomy, 10, 3, tech.Effects{GoldPerRegion: 1}),
		"better":  aiTestTechnology("better", tech.CategoryMilitary, 10, 3, tech.Effects{InfantryAttackMod: 0.20}),
	}
	gs.Factions["ai"].Research.ActiveID = "current"
	gs.Factions["ai"].Research.TurnsLeft = 2
	goldBefore := gs.Factions["ai"].Gold

	aiResearchWithStrategicContextAndSteps(gs, "ai", nil, ctx, nil)
	if gs.Factions["ai"].Research.ActiveID != "current" || gs.Factions["ai"].Research.TurnsLeft != 2 || gs.Factions["ai"].Gold != goldBefore {
		t.Fatalf("aktif araştırma yarıda değiştirilmemeliydi: %+v", gs.Factions["ai"].Research)
	}
}

func TestAIResearchPreservesLegacyScenarioPriority(t *testing.T) {
	gs, ctx := aiResearchTestState(state.AIObjectiveConsolidate)
	gs.ScenarioID = "legacy_scenario"
	gs.TechTypes = map[string]*tech.Technology{
		"military": aiTestTechnology("military", tech.CategoryMilitary, 20, 4, tech.Effects{}),
		"economy":  aiTestTechnology("economy", tech.CategoryEconomy, 10, 2, tech.Effects{GoldPerRegion: 20}),
	}

	if got := aiSelectResearchTechnology(gs, gs.Factions["ai"], nil, ctx); got == nil || got.ID != "military" {
		t.Fatalf("diğer senaryolar sabit askeri önceliği korumalıydı: got=%v", researchID(got))
	}
}

func aiResearchTestState(kind state.AIObjectiveKind) (*state.GameState, *StrategicContext) {
	gs := &state.GameState{
		ScenarioID: "1300_ottoman_rise",
		Year:       1300,
		Month:      6,
		Factions: map[faction.FactionID]*faction.Faction{
			"ai": {
				ID: "ai", Gold: 1000, Grain: 500, Iron: 200, Timber: 200, Stone: 200,
				Research: faction.ResearchState{Completed: map[string]bool{}},
			},
			"enemy": {ID: "enemy", Research: faction.ResearchState{Completed: map[string]bool{}}},
		},
		Regions: map[world.RegionID]*world.Region{
			"home": {
				ID: "home", OwnerID: "ai", Terrain: world.TerrainPlain, Satisfaction: 70,
				BaseGoldIncome: 20, BaseGrainOutput: 30, BaseIronOutput: 4, BaseTimberOutput: 4,
				BaseStoneOutput: 4, Buildings: []string{"barracks", "barracks", "barracks"},
			},
			"target": {ID: "target", OwnerID: "enemy", Terrain: world.TerrainPlain, Satisfaction: 70},
		},
		Armies: map[army.ArmyID]*army.Army{
			"field": {ID: "field", OwnerID: "ai", RegionID: "home", Units: aiUnitsOfType("base_inf", 4)},
		},
		UnitTypes: map[string]*army.UnitType{
			"base_inf": {ID: "base_inf", Category: army.CategoryInfantry, Attack: 10, Defense: 10, Morale: 50, GrainUpkeep: 2},
		},
		TechTypes: make(map[string]*tech.Technology),
		AIPlans: map[faction.FactionID]*state.AIPlanState{
			"ai": {Kind: kind, TargetFactionID: "enemy", TargetRegionIDs: []world.RegionID{"target"}},
		},
	}
	ctx := &StrategicContext{
		FactionID: "ai", gs: gs, OwnedLandRegionIDs: []world.RegionID{"home"}, WarEnemies: []faction.FactionID{"enemy"},
		ArmyAssignments: map[army.ArmyID]AIArmyAssignment{
			"field": {Role: AIArmyRoleAssault, FrontFactionID: "enemy", AnchorRegionID: "target"},
		},
	}
	return gs, ctx
}

func aiTestTechnology(id string, category tech.Category, goldCost, turns int, effects tech.Effects) *tech.Technology {
	return &tech.Technology{ID: id, Category: category, GoldCost: goldCost, TurnsRequired: turns, Effects: effects}
}

func aiResearchUnlockUnit(id string, category army.UnitCategory, requiredTech string) *army.UnitType {
	return &army.UnitType{
		ID: id, Category: category, Attack: 18, Defense: 12, Morale: 60, GoldCost: 150,
		GrainUpkeep: 3, TurnsRequired: 2, RequiredTech: requiredTech, RequiredBldg: "barracks", RequiredBldgLevel: 1,
	}
}

func researchID(technology *tech.Technology) string {
	if technology == nil {
		return ""
	}
	return technology.ID
}
