package diplomacy

import (
	"testing"

	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/religion"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestAddRelationScoreBothPreservesDirectionalScores(t *testing.T) {
	gs := &state.GameState{
		Relations: map[string]*faction.Relation{
			faction.RelationKey("a", "b"): {
				FactionA: "a", FactionB: "b", ScoreAToB: 10, ScoreBToA: -20,
			},
		},
	}

	AddRelationScoreBoth(gs, "a", "b", 5)
	if got := RelationScore(gs, "a", "b"); got != 15 {
		t.Fatalf("A->B karşılıklı değişim sonrası = %d, want 15", got)
	}
	if got := RelationScore(gs, "b", "a"); got != -15 {
		t.Fatalf("B->A karşılıklı değişim sonrası = %d, want -15", got)
	}
}

func TestApplyRelationDecayPenalizesTradeAndCancelsItBelowThreshold(t *testing.T) {
	gs := &state.GameState{
		Turn:                         1,
		AggressiveExpansionLastTurns: 12,
		Regions: map[world.RegionID]*world.Region{
			"owned": {ID: "owned", OwnerID: "expander"},
		},
		Factions: map[faction.FactionID]*faction.Faction{
			"expander": {ID: "expander"},
			"neighbor": {ID: "neighbor"},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("expander", "neighbor"): {
				FactionA:  "expander",
				FactionB:  "neighbor",
				ScoreAToB: 35,
				ScoreBToA: 35,
				Stance:    faction.StanceTrade,
			},
		},
		TradeRoutes: []*economy.TradeRoute{
			{FromFactionID: "expander", ToFactionID: "neighbor", AmountPerTurn: 1},
		},
	}
	for i := 0; i < 30; i++ {
		gs.RecordRegionAcquisition("expander", faction.FactionID("old_"+string(rune('a'+i))))
	}

	for turn := 1; turn <= 4; turn++ {
		gs.Turn = turn
		ApplyRelationDecay(gs)
	}

	rel := gs.Relations[faction.RelationKey("expander", "neighbor")]
	if got := RelationScore(gs, "neighbor", "expander"); got >= tradeRelationThreshold {
		t.Fatalf("aşırı genişleme sonrası expander görüşü yeterince düşmedi: %d", got)
	}
	if rel.Stance != faction.StancePeace {
		t.Fatalf("ticaret eşiğin altında barışa dönmedi: %q", rel.Stance)
	}
	if HasTradeRouteBetween(gs, "expander", "neighbor") {
		t.Fatal("ilişki eşiğin altına inince ticaret rotası korunmuş")
	}
}

func TestApplyRelationDecayAddsPassiveRelationTrendUpTo25(t *testing.T) {
	gs := &state.GameState{
		Turn: 1,
		Factions: map[faction.FactionID]*faction.Faction{
			"quiet":  {ID: "quiet"},
			"friend": {ID: "friend"},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("quiet", "friend"): {
				FactionA: "quiet",
				FactionB: "friend",
				Stance:   faction.StancePeace,
			},
		},
	}

	for turn := 1; turn <= 30; turn++ {
		gs.Turn = turn
		ApplyRelationDecay(gs)
	}

	rel := gs.Relations[faction.RelationKey("quiet", "friend")]
	if rel.ScoreFrom("quiet") != 25 || rel.PassiveModifierFrom("quiet") != 25 {
		t.Fatalf("saldırısız ilişki trendi = score %d modifier %d, want 25/25", rel.ScoreFrom("quiet"), rel.PassiveModifierFrom("quiet"))
	}
	gs.Turn = 30
	ApplyRelationDecay(gs)
	if rel.ScoreFrom("quiet") != 25 || rel.PassiveModifierFrom("quiet") != 25 {
		t.Fatalf("aynı turda ikinci trend uygulandı: score %d modifier %d", rel.ScoreFrom("quiet"), rel.PassiveModifierFrom("quiet"))
	}
}

func TestApplyRelationDecayPenalizesEveryOtherRelationAfterRegionAttack(t *testing.T) {
	gs := &state.GameState{
		Turn: 1,
		Factions: map[faction.FactionID]*faction.Faction{
			"attacker": {ID: "attacker"},
			"one":      {ID: "one"},
			"two":      {ID: "two"},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("attacker", "one"): {
				FactionA: "attacker",
				FactionB: "one",
				Stance:   faction.StancePeace,
			},
			faction.RelationKey("attacker", "two"): {
				FactionA: "attacker",
				FactionB: "two",
				Stance:   faction.StancePeace,
			},
		},
	}

	gs.RecordFactionRegionAttack("attacker")
	ApplyRelationDecay(gs)

	for _, other := range []faction.FactionID{"one", "two"} {
		rel := gs.Relations[faction.RelationKey("attacker", other)]
		if got := RelationScore(gs, "attacker", other); got != -1 || rel.PassiveModifierFrom("attacker") != -1 {
			t.Fatalf("%s saldıran görüşü = score %d modifier %d, want -1/-1", other, got, rel.PassiveModifierFrom("attacker"))
		}
		if got := RelationScore(gs, other, "attacker"); got != 1 || rel.PassiveModifierFrom(other) != 1 {
			t.Fatalf("%s saldırmayan görüşü = score %d modifier %d, want +1/+1", other, got, rel.PassiveModifierFrom(other))
		}
	}
}

func TestReligionAttackRelationTrendTargetsReligiousGroupsOncePerTurn(t *testing.T) {
	gs := &state.GameState{
		Turn: 1,
		Factions: map[faction.FactionID]*faction.Faction{
			"attacker":        {ID: "attacker", Religion: religion.Sunni},
			"shia_one":        {ID: "shia_one", Religion: religion.Shia},
			"shia_two":        {ID: "shia_two", Religion: religion.Shia},
			"catholic_target": {ID: "catholic_target", Religion: religion.Catholic},
			"sunni_one":       {ID: "sunni_one", Religion: religion.Sunni},
			"orthodox":        {ID: "orthodox", Religion: religion.Orthodox},
		},
		Relations: map[string]*faction.Relation{},
	}
	for _, other := range []faction.FactionID{"shia_one", "sunni_one", "orthodox", "catholic_target"} {
		gs.Relations[faction.RelationKey("attacker", other)] = &faction.Relation{
			FactionA: "attacker",
			FactionB: other,
			Stance:   faction.StancePeace,
		}
	}

	gs.RecordFactionRegionAttackAgainst("attacker", "shia_one")
	gs.RecordFactionRegionAttackAgainst("attacker", "shia_two")
	gs.RecordFactionRegionAttackAgainst("attacker", "catholic_target")
	for _, other := range []faction.FactionID{"shia_one", "sunni_one", "orthodox", "catholic_target"} {
		applyReligionAttackRelationTrend(gs, gs.Relations[faction.RelationKey("attacker", other)])
	}

	if got := RelationScore(gs, "attacker", "shia_one"); got != -3 {
		t.Fatalf("karşı din ilişkisi = %d, want -3", got)
	}
	if got := RelationScore(gs, "shia_one", "attacker"); got != 0 {
		t.Fatalf("hedef dininin saldıran görüşü = %d, want 0", got)
	}
	if got := RelationScore(gs, "attacker", "sunni_one"); got != 1 {
		t.Fatalf("aynı din ilişkisi = %d, want +1", got)
	}
	if got := RelationScore(gs, "sunni_one", "attacker"); got != 0 {
		t.Fatalf("aynı din devletinin saldıran görüşü = %d, want 0", got)
	}
	if got := RelationScore(gs, "attacker", "orthodox"); got != 0 {
		t.Fatalf("ilgisiz din ilişkisi = %d, want 0", got)
	}
	if got := RelationScore(gs, "attacker", "catholic_target"); got != -3 {
		t.Fatalf("ikinci karşı din ilişkisi = %d, want -3", got)
	}
}
