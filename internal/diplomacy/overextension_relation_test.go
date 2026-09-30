package diplomacy

import (
	"testing"

	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/religion"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

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
				FactionA: "expander",
				FactionB: "neighbor",
				Score:    35,
				Stance:   faction.StanceTrade,
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
	if rel.Score >= tradeRelationThreshold {
		t.Fatalf("aşırı genişleme sonrası ticaret ilişkisi yeterince düşmedi: %d", rel.Score)
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
	if rel.Score != 25 || rel.PassiveRelationModifier != 25 {
		t.Fatalf("saldırısız ilişki trendi = score %d modifier %d, want 25/25", rel.Score, rel.PassiveRelationModifier)
	}
	gs.Turn = 30
	ApplyRelationDecay(gs)
	if rel.Score != 25 || rel.PassiveRelationModifier != 25 {
		t.Fatalf("aynı turda ikinci trend uygulandı: score %d modifier %d", rel.Score, rel.PassiveRelationModifier)
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
		if rel.Score != -1 || rel.PassiveRelationModifier != -1 {
			t.Fatalf("%s ilişkisi = score %d modifier %d, want -1/-1", other, rel.Score, rel.PassiveRelationModifier)
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

	if got := gs.Relations[faction.RelationKey("attacker", "shia_one")].Score; got != -3 {
		t.Fatalf("karşı din ilişkisi = %d, want -3", got)
	}
	if got := gs.Relations[faction.RelationKey("attacker", "sunni_one")].Score; got != 1 {
		t.Fatalf("aynı din ilişkisi = %d, want +1", got)
	}
	if got := gs.Relations[faction.RelationKey("attacker", "orthodox")].Score; got != 0 {
		t.Fatalf("ilgisiz din ilişkisi = %d, want 0", got)
	}
	if got := gs.Relations[faction.RelationKey("attacker", "catholic_target")].Score; got != -3 {
		t.Fatalf("ikinci karşı din ilişkisi = %d, want -3", got)
	}
}
