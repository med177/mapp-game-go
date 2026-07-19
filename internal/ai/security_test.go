package ai

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func aiSecurityTestState() *state.GameState {
	return &state.GameState{
		ScenarioID: "1300_ottoman_rise",
		Turn:       4,
		Factions: map[faction.FactionID]*faction.Faction{
			"ai": {ID: "ai", Religion: "sunni", AIAggressiveness: 55},
		},
		Regions: map[world.RegionID]*world.Region{
			"capital": {
				ID: "capital", OwnerID: "ai", Religion: "sunni", Satisfaction: 60,
				Neighbors: []world.RegionID{"middle"},
			},
			"middle": {
				ID: "middle", OwnerID: "ai", Religion: "sunni", Satisfaction: 60,
				Neighbors: []world.RegionID{"capital", "unstable"},
			},
			"unstable": {
				ID: "unstable", OwnerID: "ai", Religion: "sunni", Satisfaction: 34,
				Neighbors: []world.RegionID{"middle"},
			},
		},
		Armies: map[army.ArmyID]*army.Army{
			"small": {
				ID: "small", OwnerID: "ai", RegionID: "capital", MovePoints: 2, MaxMovePoints: 2,
				Units: aiRetreatUnits(2, army.MaxUnitHP),
			},
			"big": {
				ID: "big", OwnerID: "ai", RegionID: "capital", MovePoints: 2, MaxMovePoints: 2,
				Units: aiRetreatUnits(5, army.MaxUnitHP),
			},
		},
		UnitTypes: map[string]*army.UnitType{
			"inf": {ID: "inf", Category: army.CategoryInfantry, Attack: 10, Defense: 10, GrainUpkeep: 1},
		},
		Relations: map[string]*faction.Relation{},
		AIPlans: map[faction.FactionID]*state.AIPlanState{
			"ai": {
				ObjectiveID: "secure_realm", Kind: state.AIObjectiveConsolidate,
				TargetRegionIDs: []world.RegionID{"capital"}, StartedTurn: 1, ReassessTurn: 10,
			},
		},
	}
}

func TestSecurityAssignsSmallestFieldArmyAndUsesFriendlyRoute(t *testing.T) {
	gs := aiSecurityTestState()
	ctx := prepareStrategicContext(gs, "ai")

	assignment := ctx.ArmyAssignments["small"]
	if assignment.Role != AIArmyRoleSecurity || assignment.AnchorRegionID != "unstable" {
		t.Fatalf("en küçük uygun saha ordusu düşük memnuniyetli bölgeye atanmalıydı: %+v", assignment)
	}
	if got := ctx.ArmyAssignments["big"].Role; got == AIArmyRoleSecurity {
		t.Fatalf("daha büyük ordu gereksiz yere security rolüne alınmamalıydı: %s", got)
	}
	if ctx.ReserveAssignedPower != 0 {
		t.Fatalf("security için kullanılan rezerv gücü savaş hazırlığında hâlâ sayılmamalıydı: %d", ctx.ReserveAssignedPower)
	}
	if target := chooseBestMoveWithStrategicContext(gs, gs.Armies["small"], ctx); target != "middle" {
		t.Fatalf("security ordusu yalnız dost kara hattında ilk adıma ilerlemeliydi: %s", target)
	}
	gs.Armies["small"].RegionID = "unstable"
	if target := chooseBestMoveWithStrategicContext(gs, gs.Armies["small"], ctx); target != "" {
		t.Fatalf("security ordusu anchor'a ulaşınca memnuniyet düzelene kadar ayrılmamalıydı: %s", target)
	}
}

func TestSecurityReleasesArmyAtSameReligionThreshold(t *testing.T) {
	gs := aiSecurityTestState()
	first := prepareStrategicContext(gs, "ai")
	if first.ArmyAssignments["small"].Role != AIArmyRoleSecurity {
		t.Fatalf("test önkoşulu security rolünü üretmedi: %+v", first.ArmyAssignments)
	}

	gs.Regions["unstable"].Satisfaction = aiSecuritySatisfactionThreshold
	released := prepareStrategicContext(gs, "ai")
	if released.ArmyAssignments["small"].Role == AIArmyRoleSecurity || released.ArmyAssignments["big"].Role == AIArmyRoleSecurity {
		t.Fatalf("memnuniyet yüzde 35'e ulaşınca security rolü bırakılmalıydı: %+v", released.ArmyAssignments)
	}
}

func TestSecurityUsesHigherThresholdForReligionMismatch(t *testing.T) {
	gs := aiSecurityTestState()
	gs.Regions["unstable"].Religion = "catholic"
	gs.Regions["unstable"].Satisfaction = 44
	ctx := prepareStrategicContext(gs, "ai")
	if ctx.ArmyAssignments["small"].Role != AIArmyRoleSecurity {
		t.Fatalf("din farkında yüzde 44 security rolünü tetiklemeliydi: %+v", ctx.ArmyAssignments)
	}

	gs.Regions["unstable"].Satisfaction = aiSecurityForeignReligionThreshold
	released := prepareStrategicContext(gs, "ai")
	if released.ArmyAssignments["small"].Role == AIArmyRoleSecurity || released.ArmyAssignments["big"].Role == AIArmyRoleSecurity {
		t.Fatalf("din farkında tam yüzde 45 security rolünü bırakmalıydı: %+v", released.ArmyAssignments)
	}
}

func TestWallsAndFixedGarrisonSuppressSecurityAssignment(t *testing.T) {
	t.Run("walls", func(t *testing.T) {
		gs := aiSecurityTestState()
		gs.Regions["unstable"].Satisfaction = 10
		gs.Regions["unstable"].Buildings = []string{"walls"}
		ctx := prepareStrategicContext(gs, "ai")
		if ctx.ArmyAssignments["small"].Role == AIArmyRoleSecurity || ctx.ArmyAssignments["big"].Role == AIArmyRoleSecurity {
			t.Fatalf("surlar mevcut isyan kuralıyla bölgeyi koruduğunda mobil security ayrılmamalıydı: %+v", ctx.ArmyAssignments)
		}
	})

	t.Run("fixed garrison", func(t *testing.T) {
		gs := aiSecurityTestState()
		gs.Regions["unstable"].Satisfaction = 10
		gs.Armies["garrison"] = &army.Army{
			ID: "garrison", OwnerID: "ai", RegionID: "unstable", IsGarrison: true,
			Units: aiRetreatUnits(1, army.MaxUnitHP),
		}
		ctx := prepareStrategicContext(gs, "ai")
		if ctx.ArmyAssignments["small"].Role == AIArmyRoleSecurity || ctx.ArmyAssignments["big"].Role == AIArmyRoleSecurity {
			t.Fatalf("sabit garnizon bulunan bölge için mobil security ayrılmamalıydı: %+v", ctx.ArmyAssignments)
		}
	})
}

func TestSingleFieldArmyOnlySecuresImmediateRebellionRisk(t *testing.T) {
	gs := aiSecurityTestState()
	delete(gs.Armies, "big")

	preventive := prepareStrategicContext(gs, "ai")
	if preventive.ArmyAssignments["small"].Role == AIArmyRoleSecurity {
		t.Fatalf("tek saha ordusu yüzde 34 önleyici güvenlikte dondurulmamalıydı: %+v", preventive.ArmyAssignments["small"])
	}

	gs.Regions["unstable"].Satisfaction = 29
	immediate := prepareStrategicContext(gs, "ai")
	if immediate.ArmyAssignments["small"].Role != AIArmyRoleSecurity {
		t.Fatalf("tek saha ordusu gerçek isyan eşiğinde bölgeyi korumalıydı: %+v", immediate.ArmyAssignments["small"])
	}
}

func TestImmediateSecurityUsesArmyThatCanArriveThisTurn(t *testing.T) {
	gs := aiSecurityTestState()
	gs.Regions["unstable"].Satisfaction = 20
	gs.Armies["small"].MovePoints = 1

	ctx := prepareStrategicContext(gs, "ai")
	if got := ctx.ArmyAssignments["big"]; got.Role != AIArmyRoleSecurity || got.AnchorRegionID != "unstable" {
		t.Fatalf("acil riskte uzaktaki küçük yerine bu tur erişebilen ordu seçilmeliydi: %+v", ctx.ArmyAssignments)
	}
}

func TestSecurityDoesNotStealCriticalFrontDefense(t *testing.T) {
	gs := aiSecurityTestState()
	ctx := buildStrategicContext(gs, "ai")
	ctx.ArmyAssignments = map[army.ArmyID]AIArmyAssignment{
		"small": {Role: AIArmyRoleDefense, AnchorRegionID: "capital", FrontFactionID: "enemy", Reason: "tehdit altındaki cephe"},
		"big":   {Role: AIArmyRoleAssault, AnchorRegionID: "capital"},
	}

	applySecurityAssignments(ctx)
	if got := ctx.ArmyAssignments["small"]; got.Role != AIArmyRoleDefense || got.FrontFactionID != "enemy" {
		t.Fatalf("kritik cephe savunması security için bozulmamalıydı: %+v", got)
	}
	if got := ctx.ArmyAssignments["big"].Role; got != AIArmyRoleSecurity {
		t.Fatalf("uygun diğer ordu security görevini almalıydı: %+v", ctx.ArmyAssignments)
	}
}

func TestRetreatOverridesSecurityForWornArmy(t *testing.T) {
	gs := aiSecurityTestState()
	gs.Armies["small"].Units = aiRetreatUnits(2, 40)

	ctx := prepareStrategicContext(gs, "ai")
	if got := ctx.ArmyAssignments["small"]; got.Role != AIArmyRoleRetreat {
		t.Fatalf("ağır yıpranmış ordunun hayatta kalma retreat'i security rolünden öncelikli olmalıydı: %+v", got)
	}
}
