package ai

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/diplomacy"
	"mapp-game-go/internal/economy"
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

func TestAISovereignPreservesPrivilegeWhenTradeRoyaltyCoversLocalShare(t *testing.T) {
	gs := aiPrivilegeTestState(0, faction.StancePeace)
	gs.BaseGoldValues = map[economy.GoodType]int{economy.GoodCloth: 8}
	gs.TradeRoutes = []*economy.TradeRoute{{
		FromFactionID: "sovereign", ToFactionID: "operator",
		Good: economy.GoodCloth, AmountPerTurn: 3, IsPrivilegedMinor: true,
	}}
	gs.Turn = state.DefaultMinorPrivilegeProtectionTurns + 1

	if aiManageMinorPrivilegesWithSteps(gs, "sovereign", nil) {
		t.Fatal("rota telifi yerel payı karşıladığı halde AI imtiyazı kaldırdı")
	}
	if !gs.Regions["minor"].IsPrivileged {
		t.Fatal("ticari avantajı kârlı olan imtiyaz korunmadı")
	}
}

func TestAIGoldProductionIncludesExpectedPrivilegedTradeIncome(t *testing.T) {
	gs := aiPrivilegeTestState(0, faction.StancePeace)
	gs.BaseGoldValues = map[economy.GoodType]int{economy.GoodCloth: 8}
	gs.TradeRoutes = []*economy.TradeRoute{{
		FromFactionID: "sovereign", ToFactionID: "operator",
		Good: economy.GoodCloth, AmountPerTurn: 3, IsPrivilegedMinor: true,
	}}

	if got := gs.ExpectedPrivilegedTradeIncomeForFaction("operator"); got != 24 {
		t.Fatalf("işletmeci beklenen imtiyaz ticaret geliri = %d, 24 bekleniyordu", got)
	}
	if got := gs.ExpectedPrivilegedTradeIncomeForFaction("sovereign"); got != 6 {
		t.Fatalf("egemen beklenen imtiyaz telifi = %d, 6 bekleniyordu", got)
	}
	if got := aiFactionGoldProduction(gs, "operator"); got != 29 {
		t.Fatalf("AI işletmeci toplam beklenen altın üretimi = %d, 29 bekleniyordu", got)
	}
}

func TestAISovereignOffersUnprivilegedMinorToPlayer(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"sovereign": {ID: "sovereign", NameTR: "Egemen"},
			"player":    {ID: "player", NameTR: "Oyuncu"},
		},
		Regions: map[world.RegionID]*world.Region{
			"parent": {ID: "parent", OwnerID: "sovereign"},
			"minor":  {ID: "minor", NameTR: "Minor", OwnerID: "sovereign", IsMinorRegion: true, ParentRegionID: "parent"},
		},
	}
	steps := make([]TurnStep, 0, 1)
	if !aiOfferMinorPrivilegeWithSteps(gs, "sovereign", &steps) {
		t.Fatal("AI imtiyazsız minor için oyuncuya teklif göndermedi")
	}
	if len(gs.DiplomaticOffers) != 1 {
		t.Fatalf("AI imtiyaz teklif kuyruğu = %d, 1 bekleniyordu", len(gs.DiplomaticOffers))
	}
	offer := gs.DiplomaticOffers[0]
	if offer.Action != string(diplomacy.ActionOfferMinorPrivilege) || offer.ToFactionID != "player" || offer.RegionID != "minor" {
		t.Fatalf("AI imtiyaz teklifi yanlış: %#v", offer)
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
