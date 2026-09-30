package ai

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/diplomacy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestAIPartitionedMinorGoldUsesActualFactionShare(t *testing.T) {
	parent := &world.Region{ID: "parent", OwnerID: "sovereign"}
	minor := &world.Region{
		ID: "minor", OwnerID: "operator", IsMinorRegion: true,
		IsPrivileged: true, ParentRegionID: parent.ID, BaseGoldIncome: 10,
		TaxRate: 100, Satisfaction: 50,
	}
	gs := &state.GameState{Regions: map[world.RegionID]*world.Region{
		parent.ID: parent,
		minor.ID:  minor,
	}}
	production := gs.RegionProductionSummary(minor)
	if got := aiFactionGoldFromProduction(gs, "sovereign", minor, production); got != 5 {
		t.Fatalf("egemen AI minor altın payı = %d, 5 bekleniyordu", got)
	}
	if got := aiFactionGoldFromProduction(gs, "operator", minor, production); got != 5 {
		t.Fatalf("işletmeci AI minor altın payı = %d, 5 bekleniyordu", got)
	}
}

func TestAISovereignRevokesValuableMinorPrivilege(t *testing.T) {
	gs := aiPrivilegeTestState(0, faction.StancePeace)
	gs.Turn = state.DefaultMinorPrivilegeProtectionTurns + 1
	steps := make([]TurnStep, 0, 1)

	if !aiManageMinorPrivilegesWithSteps(gs, "sovereign", &steps) {
		t.Fatal("egemen AI kârlı imtiyazı kaldırmadı")
	}
	minor := gs.Regions["minor"]
	if minor.IsPrivileged {
		t.Fatal("AI imtiyazı kaldırmasına rağmen bölge imtiyazlı kaldı")
	}
	if minor.OwnerID != "sovereign" {
		t.Fatalf("imtiyaz kaldırılan bölgenin sahibi = %q, sovereign bekleniyordu", minor.OwnerID)
	}
	if relation := diplomacyRelationForTest(gs, "sovereign", "operator"); relation == nil || diplomacy.RelationScore(gs, "sovereign", "operator") != -10 {
		t.Fatalf("imtiyaz kaldırma ilişki cezası uygulanmadı: %#v", relation)
	}
	if len(steps) != 1 || steps[0].TargetRegion != "minor" {
		t.Fatalf("AI imtiyaz adımı oluşturmadı: %#v", steps)
	}
}

func TestAIPreservesValuablePrivilegeForTrustedOperator(t *testing.T) {
	gs := aiPrivilegeTestState(60, faction.StanceAllied)
	gs.Turn = state.DefaultMinorPrivilegeProtectionTurns + 1
	if aiManageMinorPrivilegesWithSteps(gs, "sovereign", nil) {
		t.Fatal("AI müttefik işletmeciye ait imtiyazı gereksiz yere kaldırdı")
	}
	minor := gs.Regions["minor"]
	if !minor.IsPrivileged || minor.OwnerID != "operator" {
		t.Fatalf("güvenilir işletmecinin imtiyazı korunmadı: %#v", minor)
	}
}

func TestAIPreservesNewMinorPrivilegeDuringProtectionWindow(t *testing.T) {
	gs := aiPrivilegeTestState(0, faction.StancePeace)
	gs.Turn = 10
	gs.Regions["minor"].PrivilegeGrantedTurn = 1

	if aiManageMinorPrivilegesWithSteps(gs, "sovereign", nil) {
		t.Fatal("AI koruma süresi dolmadan yeni imtiyazı kaldırdı")
	}
	if !gs.Regions["minor"].IsPrivileged {
		t.Fatal("koruma süresindeki imtiyaz korunmadı")
	}
}

func TestAIRevokesNewMinorPrivilegeImmediatelyWhenAtWarWithOperator(t *testing.T) {
	gs := aiPrivilegeTestState(0, faction.StanceWar)
	gs.Turn = 10
	gs.Regions["minor"].PrivilegeGrantedTurn = 1

	if !aiManageMinorPrivilegesWithSteps(gs, "sovereign", nil) {
		t.Fatal("AI savaş varken koruma süresindeki imtiyazı kaldırmadı")
	}
}

func TestAISovereignPrivilegeRevokeEvictsOperatorArmy(t *testing.T) {
	gs := aiPrivilegeTestState(0, faction.StancePeace)
	gs.Regions["minor"].WorldX = 100
	gs.Regions["operator_home"].WorldX = 120
	gs.Armies = map[army.ArmyID]*army.Army{
		"operator_army": {ID: "operator_army", OwnerID: "operator", RegionID: "minor"},
	}

	if !aiManageMinorPrivilegesWithSteps(gs, "sovereign", nil) {
		t.Fatal("egemen AI imtiyazı kaldırmadı")
	}
	if got := gs.Armies["operator_army"].RegionID; got != "operator_home" {
		t.Fatalf("tahliye edilen işletmeci ordusu = %q, operator_home bekleniyordu", got)
	}
}

func TestAISovereignPrivilegeRevokeEliminatesLandlessOperator(t *testing.T) {
	gs := aiPrivilegeTestState(0, faction.StancePeace)
	delete(gs.Regions, "operator_home")

	if !aiManageMinorPrivilegesWithSteps(gs, "sovereign", nil) {
		t.Fatal("egemen AI imtiyazı kaldırmadı")
	}
	if !gs.Factions["operator"].IsEliminated {
		t.Fatal("topraksız kalan işletmeci AI revocation sonrasında elenmedi")
	}
}

func aiPrivilegeTestState(relationScore int, stance faction.DiplomaticStance) *state.GameState {
	return &state.GameState{
		Turn: state.DefaultMinorPrivilegeProtectionTurns + 1,
		Relations: map[string]*faction.Relation{
			faction.RelationKey("sovereign", "operator"): {
				FactionA: "sovereign", FactionB: "operator",
				ScoreAToB: relationScore, ScoreBToA: relationScore, Stance: stance,
			},
		},
		Factions: map[faction.FactionID]*faction.Faction{
			"sovereign": {ID: "sovereign", NameTR: "Egemen"},
			"operator":  {ID: "operator", NameTR: "İşletmeci"},
		},
		Regions: map[world.RegionID]*world.Region{
			"parent": {ID: "parent", OwnerID: "sovereign"},
			"minor": {
				ID: "minor", NameTR: "Minor", OwnerID: "operator", BaseGoldIncome: 10,
				IsMinorRegion: true, IsPrivileged: true, ParentRegionID: "parent",
				PrivilegeGrantedTurn: 1,
				TaxRate:              100, Satisfaction: 50,
			},
			"operator_home": {ID: "operator_home", OwnerID: "operator"},
		},
	}
}

func diplomacyRelationForTest(gs *state.GameState, a, b faction.FactionID) *faction.Relation {
	if gs == nil {
		return nil
	}
	return gs.Relations[faction.RelationKey(a, b)]
}
