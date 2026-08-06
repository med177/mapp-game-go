package ai

import (
	"testing"

	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestAIAdjustTaxesProtectsUnrestAndRaisesHealthyRevenue(t *testing.T) {
	gs := &state.GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"ai": {ID: "ai"},
		},
		Regions: map[world.RegionID]*world.Region{
			"rebellion-risk": {ID: "rebellion-risk", OwnerID: "ai", Satisfaction: 25, TaxRate: 80},
			"unstable":       {ID: "unstable", OwnerID: "ai", Satisfaction: 45, TaxRate: 80},
			"healthy":        {ID: "healthy", OwnerID: "ai", Satisfaction: 80, TaxRate: 40},
			"stable":         {ID: "stable", OwnerID: "ai", Satisfaction: 60, TaxRate: 40},
			"neutral":        {ID: "neutral", OwnerID: "other", Satisfaction: 20, TaxRate: 100},
			"sea":            {ID: "sea", OwnerID: "ai", IsSea: true, Satisfaction: 20, TaxRate: 100},
		},
	}

	aiAdjustTaxesWithSteps(gs, "ai", nil)

	if got := gs.Regions["rebellion-risk"].TaxRate; got != 60 {
		t.Fatalf("isyan riski yüksek bölgede vergi 20 puan düşmeliydi, got=%d", got)
	}
	if got := gs.Regions["unstable"].TaxRate; got != 70 {
		t.Fatalf("düşük memnuniyetli bölgede vergi 10 puan düşmeliydi, got=%d", got)
	}
	if got := gs.Regions["healthy"].TaxRate; got != 50 {
		t.Fatalf("iyi memnuniyetli bölgede vergi 10 puan artmalıydı, got=%d", got)
	}
	if got := gs.Regions["stable"].TaxRate; got != 40 {
		t.Fatalf("orta memnuniyetli bölgede vergi değişmemeliydi, got=%d", got)
	}
	if got := gs.Regions["neutral"].TaxRate; got != 100 {
		t.Fatalf("başka devletin bölgesine dokunulmamalıydı, got=%d", got)
	}
	if got := gs.Regions["sea"].TaxRate; got != 100 {
		t.Fatalf("deniz bölgesinde vergi değişmemeliydi, got=%d", got)
	}
}

func TestAIAdjustTaxesAccountsForIndependentWarFatigue(t *testing.T) {
	gs := &state.GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"ai":    {ID: "ai"},
			"enemy": {ID: "enemy"},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("ai", "enemy"): {
				FactionA: "ai", FactionB: "enemy", Stance: faction.StanceWar,
			},
		},
		Regions: map[world.RegionID]*world.Region{
			"front": {ID: "front", OwnerID: "ai", Satisfaction: 76, TaxRate: 40},
		},
	}

	aiAdjustTaxesWithSteps(gs, "ai", nil)
	if got := gs.Regions["front"].TaxRate; got != 40 {
		t.Fatalf("savaş cezası sonrası 73 projeksiyonunda vergi artmamalıydı, got=%d", got)
	}
}

func TestAIAdjustTaxesDoesNotCreateVisibleTurnSteps(t *testing.T) {
	gs := &state.GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"ai": {ID: "ai"},
		},
		Regions: map[world.RegionID]*world.Region{
			"healthy": {ID: "healthy", OwnerID: "ai", Satisfaction: 80, TaxRate: 40},
		},
	}
	steps := make([]TurnStep, 0, 1)

	aiAdjustTaxesWithSteps(gs, "ai", &steps)

	if len(steps) != 0 {
		t.Fatalf("vergi ayarı HAMLELER için adım üretmemeli, got=%+v", steps)
	}
	if got := gs.Regions["healthy"].TaxRate; got != 50 {
		t.Fatalf("vergi state'i değişmeye devam etmeli, got=%d", got)
	}
}
