package diplomacy

import (
	"testing"

	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestAssessPeaceDesireKeepsFirstFourWarTurnsIneligibleDuringEmergency(t *testing.T) {
	gs := &state.GameState{
		Turn: 1,
		Factions: map[faction.FactionID]*faction.Faction{
			"actor":    {ID: "actor"},
			"opponent": {ID: "opponent"},
		},
		Regions: map[world.RegionID]*world.Region{
			"actor-region": {ID: "actor-region", OwnerID: "actor"},
		},
		Relations: map[string]*faction.Relation{},
	}
	ForceRelation(gs, "actor", "opponent", faction.StanceWar, 0)

	assessment := AssessPeaceDesire(gs, "actor", "opponent")
	if !assessment.Emergency {
		t.Fatal("fixture başkent/askerî acil durum üretmeliydi")
	}
	if assessment.WarTurns != 0 {
		t.Fatalf("savaş süresi 0 olmalıydı, got=%d", assessment.WarTurns)
	}
	if assessment.Eligible {
		t.Fatal("ilk dört savaş turunda acil durum barış teklifini açmamalı")
	}

	gs.Turn = peaceMinimumWarTurns + 1
	assessment = AssessPeaceDesire(gs, "actor", "opponent")
	if !assessment.Eligible {
		t.Fatal("dördüncü savaş turundan sonra barış değerlendirmesi açılmalı")
	}
}

func TestResolveAutoWarCallsBreaksAllianceWhenBlockedAllyCannotJoin(t *testing.T) {
	gs := &state.GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"caller": {ID: "caller"},
			"ally":   {ID: "ally"},
			"enemy":  {ID: "enemy"},
		},
		Relations: map[string]*faction.Relation{},
	}
	ForceRelation(gs, "caller", "ally", faction.StanceAllied, 0)
	ForceRelation(gs, "ally", "enemy", faction.StancePeace, 0)
	gs.RecordTruce("ally", "enemy")

	outcomes := resolveAutoWarCalls(gs, "caller", "enemy", "caller")
	if len(outcomes) != 1 || !outcomes[0].AllianceBroken {
		t.Fatalf("bloke olan müttefik ret ve ittifak kopmasıyla raporlanmalıydı, got=%+v", outcomes)
	}
	if rel := Relation(gs, "caller", "ally"); rel == nil || rel.Stance != faction.StancePeace {
		t.Fatalf("savaşa katılamayan müttefikle ittifak korunmamalıydı, relation=%+v", rel)
	}
}
