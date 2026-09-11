package ai

import (
	"testing"

	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestAIAdjustTaxesUsesNextSatisfactionDelta(t *testing.T) {
	gs := &state.GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"ai": {ID: "ai", Grain: 100},
		},
		Regions: map[world.RegionID]*world.Region{
			"declining": {
				ID: "declining", OwnerID: "ai", Satisfaction: 54, TaxRate: 50,
			},
		},
	}

	aiAdjustTaxesWithSteps(gs, "ai", nil)

	if got, want := gs.Regions["declining"].TaxRate, 30; got != want {
		t.Fatalf("gelecek memnuniyet düşüşü vergi acil indirimini tetiklemeli: got %d, want %d", got, want)
	}
}

func TestAIAdjustTaxesRelievesAlreadyLowSatisfaction(t *testing.T) {
	gs := &state.GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"ai": {ID: "ai", Grain: 100},
		},
		Regions: map[world.RegionID]*world.Region{
			"low": {
				ID: "low", OwnerID: "ai", Satisfaction: 45, TaxRate: 40,
			},
		},
	}

	aiAdjustTaxesWithSteps(gs, "ai", nil)

	if got, want := gs.Regions["low"].TaxRate, 30; got != want {
		t.Fatalf("düşük memnuniyet vergi indirimini tetiklemeli: got %d, want %d", got, want)
	}
}
