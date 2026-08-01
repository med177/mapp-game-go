package render

import (
	"strings"
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/combat"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestNavalShowsFriendlyDisembark(t *testing.T) {
	gs := &state.GameState{}
	fleet := &army.Army{
		ID:            "fleet",
		OwnerID:       "p1",
		RegionID:      "sea_1",
		IsNaval:       true,
		EmbarkedUnits: []army.Unit{{TypeID: "inf", CurrentHP: 100}},
	}

	if !navalShowsFriendlyDisembark(gs, fleet, &world.Region{ID: "land_a", OwnerID: "p1"}) {
		t.Fatal("kendi kara bölgesi için IN davranışı bekleniyordu")
	}
	if !navalShowsFriendlyDisembark(gs, fleet, &world.Region{ID: "land_b"}) {
		t.Fatal("boş kara bölgesi için IN davranışı bekleniyordu")
	}
	if navalShowsFriendlyDisembark(gs, fleet, &world.Region{ID: "land_e", OwnerID: "p2"}) {
		t.Fatal("düşman kara bölgesi için friendly IN davranışı olmamalı")
	}
	if navalShowsFriendlyDisembark(gs, fleet, &world.Region{ID: "sea_1", IsSea: true}) {
		t.Fatal("deniz bölgesi için friendly IN davranışı olmamalı")
	}
}

func TestNavalLandMoveTargetSettlementShowsPortsAndLandingCenters(t *testing.T) {
	port := world.Settlement{ID: "port", Type: world.SettlementPort}
	center := world.Settlement{ID: "center", Type: world.SettlementCity, IsCenter: true}
	town := world.Settlement{ID: "town", Type: world.SettlementTown}

	if !navalLandMoveTargetSettlement(port, false) {
		t.Fatal("ordusuz donanma için liman settlement işaretlenmeliydi")
	}
	if navalLandMoveTargetSettlement(center, false) {
		t.Fatal("ordusuz donanma için merkez settlement işaretlenmemeliydi")
	}
	if !navalLandMoveTargetSettlement(center, true) {
		t.Fatal("çıkarma taşıyan donanma için merkez settlement işaretlenmeliydi")
	}
	if !navalLandMoveTargetSettlement(port, true) {
		t.Fatal("çıkarma taşıyan donanma için liman settlement docking hedefi olarak işaretlenmeliydi")
	}
	if navalLandMoveTargetSettlement(town, false) || navalLandMoveTargetSettlement(town, true) {
		t.Fatal("sıradan settlement donanma kara hedefi olmamalıydı")
	}
}

func TestNavalLandMoveTargetAtMatchesRenderedSettlementMode(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "p1",
		Regions: map[world.RegionID]*world.Region{
			"sea": {ID: "sea", IsSea: true, Neighbors: []world.RegionID{"coast"}},
			"coast": {ID: "coast", OwnerID: "p1", Terrain: world.TerrainCoast, Buildings: []string{"port"}, Neighbors: []world.RegionID{"sea"}, Settlements: []world.Settlement{
				{ID: "port", Type: world.SettlementPort},
				{ID: "center", Type: world.SettlementCity, IsCenter: true},
			}},
		},
	}
	fleet := &army.Army{ID: "fleet", OwnerID: "p1", IsNaval: true, RegionID: "sea", Units: []army.Unit{{TypeID: "transport"}}}
	gs.Armies = map[army.ArmyID]*army.Army{"fleet": fleet}
	r := &Renderer{
		gs: gs,
		worldMap: &WorldMap{settlementAnchor: map[settlementAnchorKey][2]int{
			{Region: "coast", Index: 0}: {100, 100},
			{Region: "coast", Index: 1}: {140, 100},
		}},
		camScale: 1,
	}

	portX, portY := r.worldToScreen(100, 100)
	centerX, centerY := r.worldToScreen(140, 100)
	if got, settlementID, ok := r.navalLandMoveTargetAt(portX, portY, fleet); !ok || got != "coast" || settlementID != "port" {
		t.Fatalf("boş filo liman settlement'ını hedeflemeli: region=%q settlement=%q ok=%t", got, settlementID, ok)
	}
	if got, settlementID, ok := r.navalLandMoveTargetAt(centerX, centerY, fleet); ok || got != "" || settlementID != "" {
		t.Fatalf("boş filo merkez çıkarma settlement'ını hedeflememeli: region=%q settlement=%q ok=%t", got, settlementID, ok)
	}

	fleet.EmbarkedUnits = []army.Unit{{TypeID: "inf"}}
	if got, settlementID, ok := r.navalLandMoveTargetAt(centerX, centerY, fleet); !ok || got != "coast" || settlementID != "center" {
		t.Fatalf("taşıyan filo merkez settlement'ını çıkarma hedeflemeli: region=%q settlement=%q ok=%t", got, settlementID, ok)
	}
	if got, settlementID, ok := r.navalLandMoveTargetAt(portX, portY, fleet); !ok || got != "coast" || settlementID != "port" {
		t.Fatalf("taşıyan filo liman settlement'ını dock hedefi olarak seçmeli: region=%q settlement=%q ok=%t", got, settlementID, ok)
	}
}

func TestEmbarkPromptRequiresSelectedArmyMovementPoints(t *testing.T) {
	gs := &state.GameState{
		UnitTypes: map[string]*army.UnitType{
			"inf":       {ID: "inf", Embarkable: true},
			"transport": {ID: "transport", Category: army.CategoryNavalTrans, CarryCapacity: 2},
		},
		Regions: map[world.RegionID]*world.Region{
			"land": {ID: "land", Neighbors: []world.RegionID{"sea"}},
			"sea":  {ID: "sea", IsSea: true},
		},
	}
	selected := &army.Army{
		ID:         "land_army",
		OwnerID:    "p1",
		RegionID:   "land",
		MovePoints: 0,
		Units:      []army.Unit{{TypeID: "inf"}},
	}
	fleet := &army.Army{
		ID:       "fleet",
		OwnerID:  "p1",
		RegionID: "sea",
		IsNaval:  true,
		Units:    []army.Unit{{TypeID: "transport"}},
	}

	if embarkableFleetForSelectedArmy(gs, selected, fleet) {
		t.Fatal("hareket puanı biten seçili orduda BIN göstergesi görünmemeli")
	}

	selected.MovePoints = 1
	if !embarkableFleetForSelectedArmy(gs, selected, fleet) {
		t.Fatal("hareket puanı olan seçili orduda uygun filo için BIN göstergesi görünmeliydi")
	}
}

func TestBattlePlanIntentCoversNavalAndAmphibiousCombat(t *testing.T) {
	r := &Renderer{}

	landingFleet := &army.Army{
		ID:            "fleet",
		OwnerID:       "p1",
		RegionID:      "sea_1",
		IsNaval:       true,
		EmbarkedUnits: []army.Unit{{TypeID: "inf", CurrentHP: 100}},
	}
	enemyArmy := &army.Army{
		ID:       "enemy_land",
		OwnerID:  "p2",
		RegionID: "land_a",
	}
	action, context, ok := r.battlePlanIntent(landingFleet, &world.Region{ID: "land_a", OwnerID: "p2"}, enemyArmy)
	if !ok {
		t.Fatal("düşman kıyıya çıkarma için savaş planı bekleniyordu")
	}
	if action != ActionDisembarkArmy {
		t.Fatalf("çıkarma için beklenen aksiyon disembark olmalı, got=%s", action)
	}
	if context != combat.BattleContextAmphibious {
		t.Fatalf("çıkarma için beklenen context amphibious olmalı, got=%s", context)
	}

	navalFleet := &army.Army{
		ID:       "fleet",
		OwnerID:  "p1",
		RegionID: "sea_1",
		IsNaval:  true,
		Units:    []army.Unit{{TypeID: "ship", CurrentHP: 100}},
	}
	enemyFleet := &army.Army{
		ID:       "enemy_fleet",
		OwnerID:  "p2",
		RegionID: "sea_2",
		IsNaval:  true,
		Units:    []army.Unit{{TypeID: "ship", CurrentHP: 100}},
	}
	action, context, ok = r.battlePlanIntent(navalFleet, &world.Region{ID: "sea_2", IsSea: true}, enemyFleet)
	if !ok {
		t.Fatal("düşman donanmaya karşı deniz savaş planı bekleniyordu")
	}
	if action != ActionMoveArmy {
		t.Fatalf("deniz savaşı için beklenen aksiyon move olmalı, got=%s", action)
	}
	if context != combat.BattleContextNaval {
		t.Fatalf("deniz savaşı için beklenen context naval olmalı, got=%s", context)
	}
}

func TestRenderTargetRequiresAmphibiousSiegeLanding(t *testing.T) {
	gs := &state.GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"p1": {ID: "p1"}, "p2": {ID: "p2"},
		},
	}
	fleet := &army.Army{
		ID: "fleet", OwnerID: "p1", IsNaval: true,
		EmbarkedUnits: []army.Unit{{TypeID: "inf"}},
	}
	fort := &world.Region{ID: "fort", OwnerID: "p2", Buildings: []string{"walls"}}
	if !renderTargetRequiresAmphibiousSiegeLanding(gs, fleet, fort) {
		t.Fatal("düşman tahkimli kıyı için amfibi kuşatma inişi bekleniyordu")
	}
	fort.OwnerID = ""
	if renderTargetRequiresAmphibiousSiegeLanding(gs, fleet, fort) {
		t.Fatal("sahipsiz tahkimli kıyı amfibi düşman kuşatması sayılmamalı")
	}
}

func TestOpenBattlePlanUsesEmbarkedUnitsForAmphibiousPreview(t *testing.T) {
	gs := &state.GameState{
		UnitTypes: map[string]*army.UnitType{
			"inf":  {ID: "inf", Attack: 12, Defense: 10, Morale: 50},
			"ship": {ID: "ship", Attack: 30, Defense: 20, Morale: 50},
		},
		Factions: map[faction.FactionID]*faction.Faction{
			"p1": {ID: "p1", NameTR: "Ahiler"},
			"p2": {ID: "p2", NameTR: "Düşman"},
		},
	}
	r := &Renderer{gs: gs}
	fleet := &army.Army{
		ID:            "fleet",
		OwnerID:       "p1",
		IsNaval:       true,
		Units:         []army.Unit{{TypeID: "ship", CurrentHP: 100}},
		EmbarkedUnits: []army.Unit{{TypeID: "inf", CurrentHP: 100}, {TypeID: "inf", CurrentHP: 100}, {TypeID: "inf", CurrentHP: 100}},
	}
	defender := &army.Army{
		ID:      "enemy",
		OwnerID: "p2",
		Units:   []army.Unit{{TypeID: "inf", CurrentHP: 100}, {TypeID: "inf", CurrentHP: 100}},
	}
	target := &world.Region{ID: "land_a", NameTR: "Liman", Terrain: world.TerrainCoast}

	r.openBattlePlan(fleet, target, defender, ActionDisembarkArmy, combat.BattleContextAmphibious)

	landing := &army.Army{
		OwnerID: fleet.OwnerID,
		Units:   fleet.EmbarkedUnits,
	}
	expected := combat.PreviewBattleWithContextMods(landing, defender, target.Terrain, gs.UnitTypes, combat.TechMods{}, combat.TechMods{}, combat.BattleContextAmphibious, combat.BattleStanceBalanced)
	shipBased := combat.PreviewBattleWithContextMods(fleet, defender, target.Terrain, gs.UnitTypes, combat.TechMods{}, combat.TechMods{}, combat.BattleContextAmphibious, combat.BattleStanceBalanced)

	if r.battlePlan.previews[1].AttackStrength != expected.AttackStrength {
		t.Fatalf("çıkarma preview gücü embarked birliklerden gelmeli, got=%d want=%d", r.battlePlan.previews[1].AttackStrength, expected.AttackStrength)
	}
	if expected.AttackStrength == shipBased.AttackStrength {
		t.Fatalf("test kurulumu gemi ve çıkarma gücünü ayırt etmeliydi")
	}
}

func TestOpenBattlePlanUsesAmphibiousCommanderForPreviewAndSummary(t *testing.T) {
	landingCommander := &army.Commander{
		ID:     "cmd_landing",
		Name:   "Turgut Bey",
		Level:  3,
		Traits: []army.CommanderTrait{army.CommanderTraitVeteran, army.CommanderTraitAggressor},
	}
	defenderCommander := &army.Commander{
		ID:     "cmd_def",
		Name:   "Nikola",
		Level:  2,
		Traits: []army.CommanderTrait{army.CommanderTraitDefender},
	}
	gs := &state.GameState{
		UnitTypes: map[string]*army.UnitType{
			"inf":  {ID: "inf", Attack: 12, Defense: 10, Morale: 50},
			"ship": {ID: "ship", Attack: 30, Defense: 20, Morale: 50},
		},
		Factions: map[faction.FactionID]*faction.Faction{
			"p1": {ID: "p1", NameTR: "Ahiler"},
			"p2": {ID: "p2", NameTR: "Düşman"},
		},
		Armies: map[army.ArmyID]*army.Army{
			"fleet": {
				ID:                "fleet",
				OwnerID:           "p1",
				IsNaval:           true,
				Units:             []army.Unit{{TypeID: "ship", CurrentHP: 100}},
				EmbarkedUnits:     []army.Unit{{TypeID: "inf", CurrentHP: 100}, {TypeID: "inf", CurrentHP: 100}},
				EmbarkedCommander: landingCommander,
			},
		},
	}
	r := &Renderer{gs: gs}
	fleet := gs.Armies["fleet"]
	defender := &army.Army{
		ID:        "enemy",
		OwnerID:   "p2",
		Units:     []army.Unit{{TypeID: "inf", CurrentHP: 100}, {TypeID: "inf", CurrentHP: 100}},
		Commander: defenderCommander,
	}
	target := &world.Region{ID: "land_a", NameTR: "Liman", Terrain: world.TerrainCoast}

	r.openBattlePlan(fleet, target, defender, ActionDisembarkArmy, combat.BattleContextAmphibious)

	landing := &army.Army{
		OwnerID:   fleet.OwnerID,
		Units:     fleet.EmbarkedUnits,
		Commander: landingCommander,
	}
	expected := combat.PreviewBattleWithContextMods(landing, defender, target.Terrain, gs.UnitTypes, combat.TechMods{}, combat.TechMods{}, combat.BattleContextAmphibious, combat.BattleStanceBalanced)

	if r.battlePlan.previews[1].AttackStrength != expected.AttackStrength {
		t.Fatalf("çıkarma preview gücü embarked komutanıyla hesaplanmalı, got=%d want=%d", r.battlePlan.previews[1].AttackStrength, expected.AttackStrength)
	}
	if !strings.Contains(r.battlePlan.attackerSummary, "Turgut Bey") || !strings.Contains(r.battlePlan.attackerSummary, "Moral +8%") {
		t.Fatalf("saldıran komutan özeti görünmeli, got=%q", r.battlePlan.attackerSummary)
	}
	if !strings.Contains(r.battlePlan.defenderSummary, "Nikola") || !strings.Contains(r.battlePlan.defenderSummary, "Savunma +6%") {
		t.Fatalf("savunan komutan özeti görünmeli, got=%q", r.battlePlan.defenderSummary)
	}
}
