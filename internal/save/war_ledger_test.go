package save

import (
	"encoding/json"
	"testing"

	"mapp-game-go/internal/state"
)

func TestCompactCampaignStatePreservesWarLedger(t *testing.T) {
	original := campaignSaveState{
		Turn:       9,
		ScenarioID: "1300_ottoman_rise",
		WarLedgers: map[string]*state.WarLedger{
			"a|b": {
				FactionA:           "a",
				FactionB:           "b",
				StartedTurn:        3,
				InitialRegionsA:    4,
				InitialRegionsB:    3,
				CasualtiesA:        5,
				CasualtiesB:        2,
				RegionsCapturedA:   1,
				LastBattleTurn:     8,
				LastPeaceOfferTurn: 7,
			},
		},
		OfferRejectionTurns: map[string]int{
			"a|b|propose_peace": 8,
		},
	}
	raw, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := decodeCampaignSaveState(raw)
	if err != nil {
		t.Fatal(err)
	}
	restored := &state.GameState{}
	applyCampaignSaveState(restored, decoded)
	ledger := restored.WarLedgerFor("b", "a")
	if ledger == nil || ledger.StartedTurn != 3 || ledger.CasualtiesA != 5 || ledger.RegionsCapturedA != 1 || ledger.LastPeaceOfferTurn != 7 {
		t.Fatalf("war ledger save/load kaybı: %+v", ledger)
	}
	if restored.OfferRejectionTurns["a|b|propose_peace"] != 8 {
		t.Fatalf("diplomasi ret cooldown save/load kaybı: %+v", restored.OfferRejectionTurns)
	}
}
