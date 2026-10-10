package save

import (
	"encoding/json"
	"testing"

	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
)

func TestTradeDisabledChangePersistsAcrossSaveFormats(t *testing.T) {
	fid := faction.FactionID("free_folk")
	base := &faction.Faction{ID: fid, TradeDisabled: true}
	current := &faction.Faction{ID: fid, TradeDisabled: false}

	delta, ok := makeFactionSaveState(current, base)
	if !ok || delta.TradeDisabled == nil || *delta.TradeDisabled {
		t.Fatalf("compact save false değişikliğini saklamadı: %#v", delta.TradeDisabled)
	}
	data, err := json.Marshal(delta)
	if err != nil {
		t.Fatalf("compact faction delta serileştirilemedi: %v", err)
	}
	var decoded factionSaveState
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("compact faction delta çözülemedi: %v", err)
	}
	restored := &faction.Faction{ID: fid, TradeDisabled: true}
	applyFactionSaveState(restored, decoded)
	if restored.TradeDisabled {
		t.Fatal("compact save yüklenince Free Folk ticareti kapalı kaldı")
	}

	debugState := makeDebugCampaignSaveState(&state.GameState{
		Factions: map[faction.FactionID]*faction.Faction{fid: current},
	})
	debugFaction := debugState.Factions[fid]
	if debugFaction.TradeDisabled == nil || *debugFaction.TradeDisabled {
		t.Fatalf("debug save false değişikliğini saklamadı: %#v", debugFaction.TradeDisabled)
	}
	legacyConverted := convertLegacyCampaignSaveState(debugState)
	legacyDelta := legacyConverted.Factions[fid]
	if legacyDelta.TradeDisabled == nil || *legacyDelta.TradeDisabled {
		t.Fatalf("eski/debug save dönüşümü false değerini kaybetti: %#v", legacyDelta.TradeDisabled)
	}
}

func TestLegacySaveWithoutTradeDisabledPreservesScenarioDefault(t *testing.T) {
	fid := faction.FactionID("free_folk")
	legacy := convertLegacyCampaignSaveState(legacyCampaignSaveState{
		Factions: map[faction.FactionID]legacyFactionSaveState{
			fid: {},
		},
	})
	loaded := &faction.Faction{ID: fid, TradeDisabled: true}
	applyFactionSaveState(loaded, legacy.Factions[fid])
	if !loaded.TradeDisabled {
		t.Fatal("trade_disabled alanı olmayan eski save senaryo varsayılanını değiştirdi")
	}
}
